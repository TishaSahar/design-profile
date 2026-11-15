#!/usr/bin/env bash
# Checks that the website serves the frontend completely and that the API is
# reachable through it:
#   1. /health.json (written by build-webui-manifest.sh) names the expected release;
#   2. every file it lists is served with the same SHA-256, and / serves index.html;
#   3. /api/v1/health answers through the same host, i.e. nginx reaches the backend.
#
# On the VPS the requests go to the local nginx (RESOLVE_TO=127.0.0.1) even when
# SITE_URL is the public domain, so DNS is not involved and TLS is still checked.
# Set RESOLVE_TO= (empty) to check a site from any other machine.
#
# Usage: bash check-webui.sh <site-url> [expected-release]
set -euo pipefail

SITE_URL="${1%/}"
EXPECTED="${2:-}"
RESOLVE_TO="${RESOLVE_TO-127.0.0.1}"

fail() {
  echo "Frontend check failed: $*" >&2
  exit 1
}

sha256() {
  if command -v sha256sum >/dev/null; then sha256sum "$1"; else shasum -a 256 "$1"; fi | cut -d' ' -f1
}

hostport="${SITE_URL#*://}"
hostport="${hostport%%/*}"
host="${hostport%%:*}"
resolve=()
if [ -n "$RESOLVE_TO" ] && [ "$host" != "$RESOLVE_TO" ] && [ "$host" != "localhost" ]; then
  # Cover a redirect from http to https as well as an explicit port.
  resolve=(--resolve "$host:80:$RESOLVE_TO" --resolve "$host:443:$RESOLVE_TO")
  if [ "$hostport" != "$host" ]; then
    resolve+=(--resolve "$host:${hostport#*:}:$RESOLVE_TO")
  fi
fi

fetch() {
  curl -fsSL --max-time 10 ${resolve[@]+"${resolve[@]}"} -o "$2" "$SITE_URL/$1"
}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

fetch health.json "$tmp/health.json" || fail "cannot load $SITE_URL/health.json"
release=$(sed -n 's/^  "release": "\(.*\)",$/\1/p' "$tmp/health.json")
if [ -n "$EXPECTED" ] && [ "$release" != "$EXPECTED" ]; then
  fail "the site serves release '${release:-unknown}', expected '$EXPECTED'; does nginx root point to /opt/design-profile/current/webui?"
fi

count=0
index_hash=""
while read -r path hash; do
  fetch "$path" "$tmp/file" || fail "cannot load /$path"
  [ "$(sha256 "$tmp/file")" = "$hash" ] || fail "/$path differs from release $release"
  if [ "$path" = "index.html" ]; then
    index_hash="$hash"
  fi
  count=$((count + 1))
done < <(sed -n 's/^    "\(.*\)": "\([0-9a-f]\{64\}\)",\{0,1\}$/\1 \2/p' "$tmp/health.json")

[ "$count" -gt 0 ] || fail "health.json lists no files"
[ -n "$index_hash" ] || fail "health.json does not list index.html"
fetch "" "$tmp/root" || fail "cannot load $SITE_URL/"
[ "$(sha256 "$tmp/root")" = "$index_hash" ] || fail "$SITE_URL/ does not serve index.html of release $release"

fetch api/v1/health "$tmp/api" || fail "the API is not reachable through the site ($SITE_URL/api/v1/health)"

echo "Frontend OK: release $release, $count files, API reachable through $SITE_URL"
