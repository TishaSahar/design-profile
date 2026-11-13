#!/usr/bin/env bash
# Runs on the VPS as the deploy user; .github/workflows/deploy.yml pipes it over SSH.
# Points `current` at an uploaded release, restarts the backend and checks that it
# answers. If it does not, switches back to the previous release.
#
# Usage: bash activate.sh <deploy-root> <release-id>
set -euo pipefail

ROOT="$1"
RELEASE="$2"
SERVICE="${SERVICE:-design-backend}"
HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:8080/api/v1/contacts}"
KEEP_RELEASES=5

cd "$ROOT"
if [ ! -x "releases/$RELEASE/server" ]; then
  echo "Release $RELEASE is missing or incomplete" >&2
  exit 1
fi

previous=$(readlink current || true)

activate() {
  ln -sfn "$1" current
  # -n: fail instead of waiting for a password if the sudoers rule is missing.
  sudo -n systemctl restart "$SERVICE"
}

healthy() {
  for _ in $(seq 1 15); do
    if curl -fsS -o /dev/null "$HEALTH_URL"; then
      return 0
    fi
    sleep 2
  done
  return 1
}

activate "releases/$RELEASE"
if healthy; then
  echo "Release $RELEASE is live"
  # Remove all but the newest releases, never the one in use.
  ls -1dt releases/*/ | tail -n +$((KEEP_RELEASES + 1)) | while read -r dir; do
    if [ "${dir%/}" != "releases/$RELEASE" ]; then
      rm -rf "$dir"
    fi
  done
  exit 0
fi

echo "Release $RELEASE failed the health check: $HEALTH_URL" >&2
if [ -n "$previous" ] && [ "$previous" != "releases/$RELEASE" ]; then
  echo "Rolling back to $previous" >&2
  activate "$previous"
  healthy || echo "The previous release does not answer either" >&2
fi
exit 1
