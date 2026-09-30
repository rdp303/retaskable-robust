package main

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func testCfg(url string) Config {
    return Config{ProbeURL: url, Timeout: time.Second}
}

func TestProbeOnline204(t *testing.T) {
    s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNoContent)
    }))
    defer s.Close()
    got := probe(testCfg(s.URL))
    if got.State != StateOnline { t.Fatalf("state=%s want online", got.State) }
    if got.StatusCode != 204 { t.Fatalf("status=%d want 204", got.StatusCode) }
}

func TestProbeCaptiveRedirect(t *testing.T) {
    s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Location", "/login")
        w.WriteHeader(http.StatusFound)
    }))
    defer s.Close()
    got := probe(testCfg(s.URL + "/generate_204"))
    if got.State != StateCaptive { t.Fatalf("state=%s want captive", got.State) }
    if got.StatusCode != 302 { t.Fatalf("status=%d want 302", got.StatusCode) }
    if got.PortalURL != s.URL+"/login" { t.Fatalf("portal=%q want %q", got.PortalURL, s.URL+"/login") }
}

func TestProbeCaptive200(t *testing.T) {
    s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html")
        _, _ = w.Write([]byte("<html>Sign in</html>"))
    }))
    defer s.Close()
    probeURL := s.URL + "/generate_204"
    got := probe(testCfg(probeURL))
    if got.State != StateCaptive { t.Fatalf("state=%s want captive", got.State) }
    if got.StatusCode != 200 { t.Fatalf("status=%d want 200", got.StatusCode) }
    if got.PortalURL != probeURL { t.Fatalf("portal=%q want original probe %q", got.PortalURL, probeURL) }
}

func TestProbeOffline(t *testing.T) {
    got := probe(testCfg("http://127.0.0.1:1/generate_204"))
    if got.State != StateOffline { t.Fatalf("state=%s want offline", got.State) }
}
