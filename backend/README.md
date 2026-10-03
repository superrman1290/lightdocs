# LightDocs Backend

Go + Gin + PostgreSQL 16 backend for the LightDocs frontend.

## Requirements

- Go 1.23+
- PostgreSQL 16+

## Run

From this directory, copy `.env.example` to `.env` and set `DATABASE_URL`.

```bash
go mod tidy
gofmt -w cmd internal
go test ./...
go run ./cmd/server
```

Apply the schema before starting the server:

```bash
go run ./cmd/migrate -schema ../db/schema.sql
```

If the database already contains articles and images, rebuild the normalized
article-image relations once after applying the schema:

```powershell
go run ./cmd/backfillimages
```

Create the first administrator without storing a plaintext password in the repository:

```powershell
$env:DATABASE_URL="postgres://lightdocs:change-this-password@localhost:5432/lightdocs?sslmode=disable"
$env:SEED_ADMIN_USERNAME="admin"
$env:SEED_ADMIN_PASSWORD="change-this-password"
go run ./cmd/seed
```

The seed command refuses to overwrite an existing user unless `SEED_OVERWRITE=true` is explicitly set.

The server exposes `/api/v1/health` and the REST API documented in `../docs/backend-api.md`.

The initial administrator is intentionally not seeded by the application. Insert a user with a bcrypt or Argon2id hash through a deployment script, never with a plaintext password.
