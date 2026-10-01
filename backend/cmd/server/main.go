package main

import (
	"context"
	"log"
	"time"

	"github.com/superrman1290/lightdocs/backend/internal/api"
	"github.com/superrman1290/lightdocs/backend/internal/config"
	"github.com/superrman1290/lightdocs/backend/internal/store"
)

func main() {
	cfg := config.Load()
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer db.Close()

	server := api.New(cfg, db.DB)
	log.Printf("LightDocs API listening on %s", cfg.HTTPAddr)
	if err := server.Router(cfg.FrontendOrigins).Run(cfg.HTTPAddr); err != nil {
		log.Fatal(err)
	}
}
