#!/usr/bin/env bash
# Writes <webui-dir>/health.json: the release id and the SHA-256 of every file of
# the frontend, so check-webui.sh can verify the site serves exactly this build.
# Run by .github/workflows/deploy.yml after the production config.js is in place.
#
# Usage: bash build-webui-manifest.sh <webui-dir> <release-id>
set -euo pipefail

DIR="$1"
RELEASE="$2"

sha256() {
  if command -v sha256sum >/dev/null; then sha256sum "$1"; else shasum -a 256 "$1"; fi | cut -d' ' -f1
}

cd "$DIR"
files=$(find . -type f ! -name health.json ! -name '.*' | sed 's|^\./||' | LC_ALL=C sort)
if [ -z "$files" ]; then
  echo "No frontend files found in $DIR" >&2
  exit 1
fi

# check-webui.sh parses this exact layout; keep one entry per line.
{
  printf '{\n  "status": "ok",\n  "release": "%s",\n  "files": {\n' "$RELEASE"
  separator=""
  while read -r file; do
    printf '%s    "%s": "%s"' "$separator" "$file" "$(sha256 "$file")"
    separator=$',\n'
  done <<< "$files"
  printf '\n  }\n}\n'
} > health.json

echo "health.json: release $RELEASE, $(printf '%s\n' "$files" | wc -l | tr -d ' ') files"
