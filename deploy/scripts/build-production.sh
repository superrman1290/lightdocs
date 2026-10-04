#!/usr/bin/env bash
set -Eeuo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$PROJECT_ROOT"

# A relative API path keeps browser traffic on the public HTTPS origin; Nginx
# proxies it to the private Go listener.
export VITE_API_URL="${VITE_API_URL:-/api/v1}"

npm ci
npm run build

pushd backend >/dev/null
go mod download
go build -trimpath -ldflags="-s -w" -o lightdocs-server ./cmd/server
go build -trimpath -ldflags="-s -w" -o lightdocs-migrate ./cmd/migrate
go build -trimpath -ldflags="-s -w" -o lightdocs-seed ./cmd/seed
go build -trimpath -ldflags="-s -w" -o lightdocs-backfillimages ./cmd/backfillimages
popd >/dev/null

printf 'Production artifacts built: %s/dist and %s/backend/lightdocs-*\n' "$PROJECT_ROOT" "$PROJECT_ROOT"
