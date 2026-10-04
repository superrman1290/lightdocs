#!/usr/bin/env bash
set -Eeuo pipefail

if [[ "${1:-}" != "--yes" || -z "${2:-}" ]]; then
  echo "Usage: $0 --yes /absolute/path/to/lightdocs-YYYYMMDDTHHMMSSZ.dump.gz" >&2
  echo "This replaces the current database contents." >&2
  exit 2
fi

backup_file="$2"
[[ -f "$backup_file" ]] || { echo "Backup file not found: $backup_file" >&2; exit 2; }

COMPOSE_FILE="${COMPOSE_FILE:-/opt/lightdocs/current/deploy/docker-compose.production.yml}"
POSTGRES_ENV_FILE="${POSTGRES_ENV_FILE:-/etc/lightdocs/postgres.env}"
set -a
. "$POSTGRES_ENV_FILE"
set +a

gzip -cd -- "$backup_file" | docker compose --env-file "$POSTGRES_ENV_FILE" -f "$COMPOSE_FILE" exec -T postgres \
  pg_restore -U "$POSTGRES_USER" -d "$POSTGRES_DB" --clean --if-exists --no-owner

printf 'Database restore completed from: %s\n' "$backup_file"
