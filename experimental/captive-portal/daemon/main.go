package main

import (
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "sync"
    "time"
)

type State string

const (
    StateOffline State = "offline"
    StateOnline  State = "online"
    StateCaptive State = "captive"
)

type ProbeResult struct {
    State      State  `json:"state"`
    StatusCode int    `json:"status_code,omitempty"`
    PortalURL  string `json:"portal_url,omitempty"`
    Error      string `json:"error,omitempty"`
    CheckedAt  string `json:"checked_at"`
}

type Config struct {
    ProbeURL       string
    Interval       time.Duration
    Timeout        time.Duration
    RuntimeDir     string
    PortalService  string
    SettleFailures int
}

func main() {
    cfg := Config{}
    flag.StringVar(&cfg.ProbeURL, "probe", envOr("RCP_PROBE_URL", "http://connectivitycheck.gstatic.com/generate_204"), "HTTP URL expected to return 204 when internet is open")
    flag.DurationVar(&cfg.Interval, "interval", envDuration("RCP_INTERVAL", 12*time.Second), "probe interval")
    flag.DurationVar(&cfg.Timeout, "timeout", envDuration("RCP_TIMEOUT", 5*time.Second), "per-probe timeout")
    flag.StringVar(&cfg.RuntimeDir, "runtime-dir", envOr("RCP_RUNTIME_DIR", "/run/remarkable-captive-portal"), "runtime state directory")
    flag.StringVar(&cfg.PortalService, "portal-service", envOr("RCP_PORTAL_SERVICE", "rcp-portal.service"), "systemd portal service")
    flag.IntVar(&cfg.SettleFailures, "settle", envInt("RCP_SETTLE_FAILURES", 2), "number of consecutive captive results before opening UI")
    flag.Parse()

    if err := os.MkdirAll(cfg.RuntimeDir, 0755); err != nil {
        log.Fatalf("runtime dir: %v", err)
    }

    log.Printf("rcp-daemon starting: probe=%s interval=%s", cfg.ProbeURL, cfg.Interval)
    run(cfg)
}

func run(cfg Config) {
    var lastState State
    captiveCount := 0
    ticker := time.NewTicker(cfg.Interval)
    defer ticker.Stop()

    probeAndAct := func() {
        result := probe(cfg)
        if err := writeJSON(filepath.Join(cfg.RuntimeDir, "state.json"), result); err != nil {
            log.Printf("write state: %v", err)
        }

        switch result.State {
        case StateCaptive:
            captiveCount++
            if result.PortalURL != "" {
                _ = os.WriteFile(filepath.Join(cfg.RuntimeDir, "portal-url"), []byte(result.PortalURL+"\n"), 0644)
            }
            if captiveCount >= cfg.SettleFailures && lastState != StateCaptive {
                log.Printf("captive portal detected (HTTP %d), opening %s", result.StatusCode, result.PortalURL)
                if err := systemctl("start", cfg.PortalService); err != nil {
                    log.Printf("start portal: %v", err)
                }
                lastState = StateCaptive
            }
        case StateOnline:
            captiveCount = 0
            if lastState == StateCaptive {
                log.Printf("internet access confirmed; closing portal")
                if err := systemctl("stop", cfg.PortalService); err != nil {
                    log.Printf("stop portal: %v", err)
                }
            }
            if lastState != StateOnline {
                log.Printf("connectivity state: online")
            }
            lastState = StateOnline
        case StateOffline:
            captiveCount = 0
            if lastState != StateOffline {
                log.Printf("connectivity state: offline (%s)", result.Error)
            }
            lastState = StateOffline
        }
    }

    probeAndAct()
    for range ticker.C {
        probeAndAct()
    }
}

func probe(cfg Config) ProbeResult {
    result := ProbeResult{CheckedAt: time.Now().UTC().Format(time.RFC3339)}

    ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.ProbeURL, nil)
    if err != nil {
        result.State = StateOffline
        result.Error = err.Error()
        return result
    }
    req.Header.Set("User-Agent", "reMarkable-CaptivePortal/0.1")
    req.Header.Set("Cache-Control", "no-cache")

    client := &http.Client{
        Timeout: cfg.Timeout,
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }

    resp, err := client.Do(req)
    if err != nil {
        result.State = StateOffline
        result.Error = err.Error()
        return result
    }
    defer resp.Body.Close()
    result.StatusCode = resp.StatusCode
    _, _ = io.CopyN(io.Discard, resp.Body, 4096)

    if resp.StatusCode == http.StatusNoContent {
        result.State = StateOnline
        return result
    }

    result.State = StateCaptive
    result.PortalURL = cfg.ProbeURL
    if loc := strings.TrimSpace(resp.Header.Get("Location")); loc != "" {
        if base, e1 := url.Parse(cfg.ProbeURL); e1 == nil {
            if next, e2 := url.Parse(loc); e2 == nil {
                result.PortalURL = base.ResolveReference(next).String()
            }
        }
    }
    return result
}

var ctlMu sync.Mutex

func systemctl(action, unit string) error {
    ctlMu.Lock()
    defer ctlMu.Unlock()
    cmd := exec.Command("/bin/systemctl", action, unit)
    out, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("systemctl %s %s: %w: %s", action, unit, err, strings.TrimSpace(string(out)))
    }
    return nil
}

func writeJSON(path string, v any) error {
    b, err := json.MarshalIndent(v, "", "  ")
    if err != nil { return err }
    tmp := path + ".tmp"
    if err := os.WriteFile(tmp, append(b, '\n'), 0644); err != nil { return err }
    return os.Rename(tmp, path)
}

func envOr(k, def string) string {
    if v := strings.TrimSpace(os.Getenv(k)); v != "" { return v }
    return def
}

func envDuration(k string, def time.Duration) time.Duration {
    if v := strings.TrimSpace(os.Getenv(k)); v != "" {
        if d, err := time.ParseDuration(v); err == nil { return d }
    }
    return def
}

func envInt(k string, def int) int {
    if v := strings.TrimSpace(os.Getenv(k)); v != "" {
        var n int
        if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 { return n }
    }
    return def
}
