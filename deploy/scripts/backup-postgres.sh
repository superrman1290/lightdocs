#!/usr/bin/env bash
set -Eeuo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-/opt/lightdocs/current/deploy/docker-compose.production.yml}"
POSTGRES_ENV_FILE="${POSTGRES_ENV_FILE:-/etc/lightdocs/postgres.env}"
BACKUP_DIRECTORY="${BACKUP_DIRECTORY:-/var/backups/lightdocs/postgres}"

set -a
. "$POSTGRES_ENV_FILE"
set +a

mkdir -p "$BACKUP_DIRECTORY"
timestamp="$(date -u +%Y%m%dT%H%M%SZ)"
target="$BACKUP_DIRECTORY/lightdocs-$timestamp.dump"

docker compose --env-file "$POSTGRES_ENV_FILE" -f "$COMPOSE_FILE" exec -T postgres \
  pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc > "$target"

gzip -f "$target"
find "$BACKUP_DIRECTORY" -type f -name 'lightdocs-*.dump.gz' -mtime +14 -delete
printf 'Backup written: %s.gz\n' "$target"
