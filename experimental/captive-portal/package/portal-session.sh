#!/bin/sh
set -eu

RUNTIME=${RCP_RUNTIME_DIR:-/run/remarkable-captive-portal}
URL_FILE="$RUNTIME/portal-url"
[ -s "$URL_FILE" ] || exit 1
URL=$(head -n 1 "$URL_FILE")

# V0 proof-of-concept: reuse MaximeRivest's existing remagic Chromium app.
# This proves rendering, form input, redirects and auto-close. The upstream
# viewer still has a URL button, so this is NOT the final locked-down build.
CHROMIUM_APP=${RCP_CHROMIUM_APP:-/home/root/xovi/exthome/appload/chromium}
TAKEOVER="$CHROMIUM_APP/chromium-takeover.sh"

if [ ! -x "$TAKEOVER" ]; then
    echo "rcp: Chromium app not found at $CHROMIUM_APP" >&2
    echo "rcp: install remagic Chromium first, or point RCP_CHROMIUM_APP at a restricted fork" >&2
    exit 2
fi

export CHROMIUM_URL="$URL"
export CHROMIUM_WAIT_SELECTOR='body'
export CHROMIUM_RENDER_TIMEOUT=45

exec /bin/sh "$TAKEOVER"
