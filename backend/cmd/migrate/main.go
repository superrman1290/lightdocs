package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	schemaPath := flag.String("schema", "../db/schema.sql", "path to schema SQL")
	flag.Parse()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	schema, err := os.ReadFile(*schemaPath)
	if err != nil {
		log.Fatalf("read schema: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, string(schema)); err != nil {
		log.Fatalf("apply schema: %v", err)
	}
	log.Println("schema applied")
}
