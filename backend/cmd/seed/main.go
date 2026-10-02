package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	databaseURL := requiredEnv("DATABASE_URL")
	username := env("SEED_ADMIN_USERNAME", "admin")
	password := requiredEnv("SEED_ADMIN_PASSWORD")
	overwrite := strings.EqualFold(os.Getenv("SEED_OVERWRITE"), "true")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer conn.Close(ctx)

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}

	var existingID int64
	err = conn.QueryRow(ctx, `SELECT id FROM users WHERE username = $1`, username).Scan(&existingID)
	if err == nil && !overwrite {
		fmt.Printf("user %q already exists; keeping it\n", username)
	}
	if err != nil && err != pgx.ErrNoRows {
		log.Fatalf("check user: %v", err)
	}

	if err == nil && overwrite {
		_, err = conn.Exec(ctx, `UPDATE users SET password_hash = $1, status = 'active', updated_at = now() WHERE id = $2`, string(hash), existingID)
		if err != nil {
			log.Fatalf("update admin: %v", err)
		}
		fmt.Printf("updated admin user %q\n", username)
	} else if err == pgx.ErrNoRows {
		_, err = conn.Exec(ctx, `INSERT INTO users (username, password_hash, role, status) VALUES ($1, $2, 'admin', 'active')`, username, string(hash))
		if err != nil {
			log.Fatalf("create admin: %v", err)
		}
		fmt.Printf("created admin user %q\n", username)
	}

	_, _ = conn.Exec(ctx, `
		INSERT INTO site_settings (id, site_name, site_title, logo_url)
		VALUES (1, '轻文档', '轻文档 - 专注技术教程的个人文档网站', '/favicon.svg')
		ON CONFLICT (id) DO NOTHING`)
	_, _ = conn.Exec(ctx, `INSERT INTO security_settings (id) VALUES (1) ON CONFLICT (id) DO NOTHING`)
	for index, name := range []string{"前端开发", "后端开发", "Docker", "数据库", "项目开发"} {
		_, err = conn.Exec(ctx, `
			INSERT INTO categories (name, parent_id, sort_order)
			VALUES ($1, NULL, $2)
			ON CONFLICT DO NOTHING`, name, index+1)
		if err != nil {
			log.Fatalf("seed category %q: %v", name, err)
		}
	}
	var dockerID int64
	if err := conn.QueryRow(ctx, `SELECT id FROM categories WHERE name = 'Docker' AND parent_id IS NULL LIMIT 1`).Scan(&dockerID); err == nil {
		_, err = conn.Exec(ctx, `
			INSERT INTO articles (title, slug, category_id, status, tags, content, summary, published_at)
			SELECT 'Docker Compose 入门', 'docker-compose-guide', $1, 'published', ARRAY['Docker','部署'], '# Docker Compose 入门', '介绍 Docker Compose 的基本使用方法。', now()
			WHERE NOT EXISTS (SELECT 1 FROM articles WHERE slug = 'docker-compose-guide' AND deleted_at IS NULL)`, dockerID)
		if err != nil {
			log.Fatalf("seed demo article: %v", err)
		}
	}
}

func requiredEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		log.Fatalf("%s is required", key)
	}
	return value
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
