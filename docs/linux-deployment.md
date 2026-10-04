# Linux Server Deployment

This guide targets Ubuntu 22.04 or 24.04 with one public domain, Nginx, a Go
API managed by systemd, and PostgreSQL 16 in Docker. Replace `docs.example.com`
with the actual domain everywhere.

## 1. Prepare the Host

Install the required runtime packages and create the non-login service user:

```bash
sudo apt update
sudo apt install -y nginx certbot docker.io docker-compose-plugin ca-certificates
sudo systemctl enable --now docker nginx
sudo useradd --system --home /opt/lightdocs --shell /usr/sbin/nologin lightdocs
sudo install -d -o lightdocs -g lightdocs /var/lib/lightdocs/uploads
sudo install -d -o root -g lightdocs -m 0750 /etc/lightdocs
sudo install -d -o root -g root /var/backups/lightdocs/postgres
```

Build artifacts can be copied from CI or built on the server. If building on
the server, also install Node.js 20+ and Go 1.23+.

## 2. Install the Application

Place the checked-out release under `/opt/lightdocs/current`, then grant the
service account read access:

```bash
sudo install -d -o lightdocs -g lightdocs /opt/lightdocs
sudo chown -R lightdocs:lightdocs /opt/lightdocs/current
sudo -u lightdocs bash /opt/lightdocs/current/deploy/scripts/build-production.sh
```

The production build uses `VITE_API_URL=/api/v1`, so browser API requests stay
on the same HTTPS domain and are proxied by Nginx.

## 3. Configure PostgreSQL

Create `/etc/lightdocs/postgres.env` from
`deploy/env/postgres.production.env.example`. Use a randomly generated password
and restrict its permissions:

```bash
sudo install -m 0600 -o root -g lightdocs \
  /opt/lightdocs/current/deploy/env/postgres.production.env.example \
  /etc/lightdocs/postgres.env
sudo editor /etc/lightdocs/postgres.env
sudo docker compose --env-file /etc/lightdocs/postgres.env \
  -f /opt/lightdocs/current/deploy/docker-compose.production.yml up -d
sudo docker compose --env-file /etc/lightdocs/postgres.env \
  -f /opt/lightdocs/current/deploy/docker-compose.production.yml ps
```

The production Compose file binds PostgreSQL to `127.0.0.1` only. Do not expose
port `5432` through the public firewall.

## 4. Configure the Go API

Create `/etc/lightdocs/lightdocs.env` from
`deploy/env/lightdocs.production.env.example`. Set the same database password
used in `postgres.env`, the real public origin, and the persistent upload path:

```bash
sudo install -m 0640 -o root -g lightdocs \
  /opt/lightdocs/current/deploy/env/lightdocs.production.env.example \
  /etc/lightdocs/lightdocs.env
sudo editor /etc/lightdocs/lightdocs.env
```

Important production values are:

```ini
GIN_MODE=release
HTTP_ADDR=127.0.0.1:8080
FRONTEND_ORIGINS=https://docs.example.com
UPLOAD_DIRECTORY=/var/lib/lightdocs/uploads
```

Run database initialization before enabling the API. The migration binary
creates an empty database from `db/schema.sql` once and subsequently applies
only new files in `db/migrations/`:

```bash
sudo -u lightdocs bash -c '
  set -a
  . /etc/lightdocs/lightdocs.env
  set +a
  cd /opt/lightdocs/current/backend
  ./lightdocs-migrate
'
```

Create the initial administrator once. Do not put the temporary seed password
in shell history; use an interactive secret mechanism where possible:

```bash
sudo -u lightdocs bash -c '
  set -a
  . /etc/lightdocs/lightdocs.env
  export SEED_ADMIN_USERNAME=admin
  export SEED_ADMIN_PASSWORD="replace-with-a-strong-password"
  set +a
  cd /opt/lightdocs/current/backend
  ./lightdocs-seed
'
```

For an existing installation upgraded to `article_images`, run:

```bash
sudo -u lightdocs bash -c '
  set -a
  . /etc/lightdocs/lightdocs.env
  set +a
  cd /opt/lightdocs/current/backend
  ./lightdocs-backfillimages
'
```

Install and start the systemd unit:

```bash
sudo cp /opt/lightdocs/current/deploy/systemd/lightdocs.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now lightdocs
sudo systemctl status lightdocs
curl http://127.0.0.1:8080/api/v1/health
```

View API logs with:

```bash
sudo journalctl -u lightdocs -f
```

## 5. Configure Nginx and TLS

Create the temporary HTTP site first:

```bash
sudo mkdir -p /var/www/certbot
sudo cp /opt/lightdocs/current/deploy/nginx/lightdocs.http.conf.example /etc/nginx/sites-available/lightdocs
sudo editor /etc/nginx/sites-available/lightdocs
sudo ln -s /etc/nginx/sites-available/lightdocs /etc/nginx/sites-enabled/lightdocs
sudo nginx -t && sudo systemctl reload nginx
```

After the domain DNS A/AAAA records point to this server, request the
certificate:

```bash
sudo certbot certonly --webroot -w /var/www/certbot -d docs.example.com
```

Replace the temporary site with the TLS template, update the domain, and reload:

```bash
sudo cp /opt/lightdocs/current/deploy/nginx/lightdocs.conf.example /etc/nginx/sites-available/lightdocs
sudo editor /etc/nginx/sites-available/lightdocs
sudo nginx -t && sudo systemctl reload nginx
```

Certbot installs renewal timers automatically. Verify with:

```bash
sudo certbot renew --dry-run
```

## 6. Firewall and Verification

Only expose SSH, HTTP and HTTPS. Do not expose PostgreSQL or the Go API:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'
sudo ufw enable
sudo ufw status
```

Verify the public service:

```bash
curl -I https://docs.example.com/
curl https://docs.example.com/api/v1/health
```

## 7. Backups and Restore

Install the daily backup schedule:

```bash
sudo chmod 0750 /opt/lightdocs/current/deploy/scripts/*.sh
sudo crontab /opt/lightdocs/current/deploy/cron/lightdocs-backup.cron
sudo /opt/lightdocs/current/deploy/scripts/backup-postgres.sh
```

Backups are stored in `/var/backups/lightdocs/postgres` for 14 days. Copy this
directory and `/var/lib/lightdocs/uploads` to independent storage; a database
backup alone does not include uploaded files.

Restore is destructive and requires an explicit confirmation flag:

```bash
sudo /opt/lightdocs/current/deploy/scripts/restore-postgres.sh --yes \
  /var/backups/lightdocs/postgres/lightdocs-YYYYMMDDTHHMMSSZ.dump.gz
```

Test restores on a separate server or database before relying on backups.

## 8. Release Updates

For each release:

```bash
sudo systemctl stop lightdocs
sudo -u lightdocs bash /opt/lightdocs/current/deploy/scripts/build-production.sh
sudo -u lightdocs bash -c '
  set -a
  . /etc/lightdocs/lightdocs.env
  set +a
  cd /opt/lightdocs/current/backend
  ./lightdocs-migrate
'
sudo systemctl start lightdocs
sudo nginx -t && sudo systemctl reload nginx
```

For zero-downtime releases, deploy each version to a new release directory and
switch `/opt/lightdocs/current` only after the build and migration complete.

## Operational Notes

- Keep `/var/lib/lightdocs/uploads` outside release directories. It must be
  backed up and owned by the `lightdocs` service account.
- Create a new numbered SQL migration for every database change. Never edit an
  already applied migration.
- Restrict server SSH access and keep the operating system, Docker, Nginx and
  PostgreSQL image patched.
- SVG uploads can contain active content. For an internet-facing deployment,
  either disable SVG uploads or sanitize them before storing.
