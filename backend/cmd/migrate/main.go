// Command migrate initializes an empty database from db/schema.sql once, then
// applies each numbered SQL file in db/migrations exactly once.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const initialSchemaVersion = "000_initial_schema"

func main() {
	schemaPath := flag.String("schema", "../db/schema.sql", "path to the initial schema SQL")
	migrationsPath := flag.String("migrations", "../db/migrations", "directory containing versioned SQL migrations")
	flag.Parse()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer conn.Close(ctx)

	if err := prepareMigrationTable(ctx, conn, *schemaPath); err != nil {
		log.Fatal(err)
	}
	if err := applyMigrationFiles(ctx, conn, *migrationsPath); err != nil {
		log.Fatal(err)
	}
	log.Println("database migrations applied")
}

func prepareMigrationTable(ctx context.Context, conn *pgx.Conn, schemaPath string) error {
	if _, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return fmt.Errorf("create migration table: %w", err)
	}

	var applied bool
	if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version=$1)`, initialSchemaVersion).Scan(&applied); err != nil {
		return fmt.Errorf("check initial schema migration: %w", err)
	}
	if applied {
		return nil
	}

	var usersTableExists bool
	if err := conn.QueryRow(ctx, `SELECT to_regclass('public.users') IS NOT NULL`).Scan(&usersTableExists); err != nil {
		return fmt.Errorf("inspect database schema: %w", err)
	}
	if usersTableExists {
		// Existing installations predate schema_migrations. Mark their current
		// baseline without rerunning the non-idempotent initial schema.
		if _, err := conn.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, initialSchemaVersion); err != nil {
			return fmt.Errorf("record existing schema baseline: %w", err)
		}
		log.Println("existing database marked as initial schema baseline")
		return nil
	}

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read initial schema: %w", err)
	}
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin initial schema transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, string(schema)); err != nil {
		return fmt.Errorf("apply initial schema: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, initialSchemaVersion); err != nil {
		return fmt.Errorf("record initial schema migration: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit initial schema migration: %w", err)
	}
	log.Println("initial schema applied")
	return nil
}

func applyMigrationFiles(ctx context.Context, conn *pgx.Conn, migrationsPath string) error {
	entries, err := os.ReadDir(migrationsPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	for _, name := range files {
		var applied bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version=$1)`, name).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied {
			continue
		}
		sql, err := os.ReadFile(filepath.Join(migrationsPath, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		tx, err := conn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err = tx.Exec(ctx, string(sql)); err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO schema_migrations(version) VALUES($1)`, name)
		}
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(ctx)
		}
		if err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		log.Printf("applied migration %s", name)
	}
	return nil
}
