# Versioned Migrations

Place every post-initial-schema PostgreSQL change in this directory using a
lexicographically sortable filename, for example:

```text
001_add_article_visibility.sql
002_add_image_checksum.sql
```

`go run ./cmd/migrate` records each completed filename in `schema_migrations`
and never executes that filename again. Do not edit an applied migration; add a
new migration instead.

The initial schema remains in `../schema.sql`. Existing installations are
baselined as `000_initial_schema` automatically on their first run of the new
migration command.
