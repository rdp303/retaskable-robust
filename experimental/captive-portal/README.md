# reMarkable Captive Wi-Fi — V0 proof of concept

Goal: let a reMarkable complete browser-based Wi-Fi sign-in **without giving the user a standalone browser**.

## Architecture

This POC is deliberately hybrid:

1. **AppLoad package** — install/enable surface.
2. **Native systemd daemon** — probes connectivity in the background.
3. **Portal session service** — starts only when the probe indicates interception/captive state.
4. **Chromium renderer** — V0 reuses the existing remagic Chromium takeover app to prove forms/JS/cookies/redirects/input.
5. **Automatic close** — when the detector starts receiving HTTP 204 again, it stops the portal session and xochitl returns.

The final version should ship a restricted fork of the Chromium viewer with the URL control removed and should use a temporary/private profile.

## Why not AppLoad-only?

AppLoad is useful for the launcher/UI, but captive detection needs to happen even when no app is open. A systemd service is the right place for that job. The production version should move from 12-second polling to Wi-Fi/network events with a slow fallback probe.

## Build on a computer

Requires Go. The detector has no third-party Go dependencies.

```sh
./scripts/build.sh
```

This cross-compiles a static ARM64 Linux binary into `package/bin/rcp-daemon`.

## Prerequisites on Paper Pro

- Developer Mode enabled.
- XOVI + AppLoad installed.
- For this V0 only: the remagic Chromium app and its Chromium engine installed.

The current V0 points at:

`/home/root/xovi/exthome/appload/chromium/chromium-takeover.sh`

## Install through remagic/AppLoad

From this directory on the computer:

```sh
./scripts/build.sh
remagic install ./package
```

Then on the tablet open AppLoad and tap **Captive Wi-Fi** once. That first launch installs and enables the native background systemd unit.

Manual equivalent:

```sh
scp -r package root@10.11.99.1:/home/root/xovi/exthome/appload/captive-portal
ssh root@10.11.99.1 '/home/root/xovi/exthome/appload/captive-portal/enable.sh'
```

## Starbucks field test

This branch is a **field-test proof of concept**. Do not merge it into a production package yet.

Before leaving home:

```sh
ssh root@10.11.99.1 'systemctl status rcp-daemon.service --no-pager'
ssh root@10.11.99.1 'test -x /home/root/xovi/exthome/appload/chromium/chromium-takeover.sh && echo chromium-ok'
```

At Starbucks:

1. Connect the reMarkable to the Starbucks Wi-Fi SSID from normal Wi-Fi settings.
2. Leave the Wi-Fi settings screen open for up to ~30 seconds. V0 requires two captive results, 12 seconds apart.
3. Expected: the captive portal renderer takes over automatically.
4. Complete the Starbucks terms/sign-in flow.
5. Expected: within ~12 seconds of Internet access being granted, the portal renderer closes and stock reMarkable UI returns.
6. Verify cloud sync or another Internet-dependent reMarkable function works.

If anything fails, capture diagnostics over USB SSH before rebooting:

```sh
/home/root/xovi/exthome/appload/captive-portal/diagnostics.sh
```

The diagnostics script intentionally does not print saved Wi-Fi passwords or browser cookies.

## V0 success criteria

- captive state is detected without manually opening a general-purpose browser;
- the Starbucks portal renders and accepts touch/keyboard input;
- authentication results in normal Internet access for the tablet;
- the portal session closes automatically after the 204 connectivity probe succeeds;
- the tablet returns to xochitl cleanly.

## Known V0 caveats

- The upstream Chromium viewer still exposes its URL control. **Do not treat this build as the final restricted product.**
- The detector polls every 12 seconds. This is intentional for field debugging; production should use Wi-Fi/D-Bus events with a slow fallback probe.
- If the portal uses an interaction unsupported by the upstream Chromium takeover app, the field test may fail even though captive detection works.
- Systemd files are installed directly under `/etc/systemd/system`; production packaging should use ecosystem persistence/package conventions across OS updates.
