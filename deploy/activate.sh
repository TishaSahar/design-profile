#!/usr/bin/env bash
# Runs on the VPS as the deploy user; .github/workflows/deploy.yml pipes it over SSH.
# Points `current` at an uploaded release, restarts the backend and checks both
# the backend (/health, including PostgreSQL) and the website served by nginx.
# If either check fails, switches back to the previous release.
#
# Usage: bash activate.sh <deploy-root> <release-id> [site-url]
set -euo pipefail

ROOT="$1"
RELEASE="$2"
SITE_URL="${3:-http://127.0.0.1}"
SERVICE="${SERVICE:-design-backend}"
BACKEND_HEALTH_URL="${BACKEND_HEALTH_URL:-http://127.0.0.1:8080/health}"
KEEP_RELEASES=5

cd "$ROOT"
for file in server check-webui.sh webui/health.json; do
  if [ ! -e "releases/$RELEASE/$file" ]; then
    echo "Release $RELEASE is missing or incomplete: no $file" >&2
    exit 1
  fi
done
# The new release's checker also verifies a rolled-back release.
CHECK_WEBUI="$ROOT/releases/$RELEASE/check-webui.sh"

previous=$(readlink current || true)

activate() {
  ln -sfn "$1" current
  # -n: fail instead of waiting for a password if the sudoers rule is missing.
  sudo -n systemctl restart "$SERVICE"
}

backend_healthy() {
  for _ in $(seq 1 15); do
    if curl -fs -o /dev/null "$BACKEND_HEALTH_URL"; then
      echo "Backend OK: $BACKEND_HEALTH_URL"
      return 0
    fi
    sleep 2
  done
  # The body names the dependency that is down, e.g. {"checks":{"postgres":"unavailable"}}.
  echo "Backend check failed: $BACKEND_HEALTH_URL does not answer with 200;" \
    "last response: $(curl -s --max-time 5 "$BACKEND_HEALTH_URL" || echo 'no response')" >&2
  return 1
}

healthy() {
  backend_healthy && bash "$CHECK_WEBUI" "$SITE_URL" "$(basename "$1")"
}

activate "releases/$RELEASE"
if healthy "releases/$RELEASE"; then
  echo "Release $RELEASE is live"
  # Remove all but the newest releases, never the one in use.
  ls -1dt releases/*/ | tail -n +$((KEEP_RELEASES + 1)) | while read -r dir; do
    if [ "${dir%/}" != "releases/$RELEASE" ]; then
      rm -rf "$dir"
    fi
  done
  exit 0
fi

echo "Release $RELEASE failed the health checks" >&2
if [ -n "$previous" ] && [ "$previous" != "releases/$RELEASE" ]; then
  echo "Rolling back to $previous" >&2
  activate "$previous"
  healthy "$previous" || echo "The previous release does not pass the checks either" >&2
fi
exit 1
