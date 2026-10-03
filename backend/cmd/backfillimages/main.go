// Command backfillimages rebuilds article_images from existing Markdown
// content. It is safe to run repeatedly because the relation has a composite
// primary key and inserts use ON CONFLICT DO NOTHING.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_article_images_image ON article_images (image_id, article_id)`); err != nil {
		log.Fatalf("create article image index: %v", err)
	}
	command := `
		INSERT INTO article_images (article_id, image_id)
		SELECT DISTINCT a.id, i.id
		FROM articles a
		JOIN images i ON a.content LIKE '%' || i.url || '%'
		                 OR a.content LIKE '%' || i.storage_key || '%'
		ON CONFLICT (article_id, image_id) DO NOTHING`
	if _, err := conn.Exec(ctx, command); err != nil {
		log.Fatalf("backfill article image relations: %v", err)
	}
	log.Println("article image relations backfilled")
}
