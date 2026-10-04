-- LightDocs PostgreSQL 16 schema
-- 生产环境请通过迁移工具执行，并在部署时注入管理员初始密码哈希。

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE users (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  username VARCHAR(30) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  role VARCHAR(20) NOT NULL DEFAULT 'admin' CHECK (role IN ('admin')),
  status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'disabled')),
  failed_login_count SMALLINT NOT NULL DEFAULT 0 CHECK (failed_login_count >= 0),
  locked_until TIMESTAMPTZ,
  last_login_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE auth_sessions (
  id UUID PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash CHAR(64) NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  ip INET,
  user_agent TEXT,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_auth_sessions_user_active
  ON auth_sessions (user_id, revoked_at, expires_at);

CREATE TABLE reauth_tokens (
  id UUID PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash CHAR(64) NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_reauth_tokens_user_expiry
  ON reauth_tokens (user_id, expires_at);

CREATE TABLE site_settings (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  site_name VARCHAR(30) NOT NULL,
  site_title VARCHAR(80) NOT NULL,
  logo_url TEXT NOT NULL,
  updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE security_settings (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  max_login_failures SMALLINT NOT NULL DEFAULT 5 CHECK (max_login_failures BETWEEN 1 AND 20),
  lock_minutes INTEGER NOT NULL DEFAULT 15 CHECK (lock_minutes BETWEEN 1 AND 1440),
  session_days INTEGER NOT NULL DEFAULT 7 CHECK (session_days BETWEEN 1 AND 365),
  remember_login BOOLEAN NOT NULL DEFAULT false,
  logout_other_devices_on_password_change BOOLEAN NOT NULL DEFAULT true,
  require_password_reauth_on_save BOOLEAN NOT NULL DEFAULT false,
  updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE categories (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name VARCHAR(80) NOT NULL,
  parent_id BIGINT REFERENCES categories(id) ON DELETE RESTRICT,
  sort_order INTEGER NOT NULL DEFAULT 0,
  created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (length(btrim(name)) > 0)
);

CREATE UNIQUE INDEX uq_categories_parent_name
  ON categories (COALESCE(parent_id, 0), lower(name));

CREATE INDEX idx_categories_tree
  ON categories (parent_id, sort_order, id);

CREATE OR REPLACE FUNCTION prevent_category_cycle()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  IF NEW.parent_id IS NULL THEN
    RETURN NEW;
  END IF;

  IF NEW.parent_id = NEW.id THEN
    RAISE EXCEPTION 'category cannot be its own parent';
  END IF;

  IF EXISTS (
    WITH RECURSIVE ancestors(id) AS (
      SELECT NEW.parent_id
      UNION ALL
      SELECT c.parent_id
      FROM categories c
      JOIN ancestors a ON c.id = a.id
      WHERE c.parent_id IS NOT NULL
    )
    SELECT 1 FROM ancestors WHERE id = NEW.id
  ) THEN
    RAISE EXCEPTION 'category parent cycle detected';
  END IF;

  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_categories_prevent_cycle
  BEFORE INSERT OR UPDATE OF parent_id ON categories
  FOR EACH ROW
  EXECUTE FUNCTION prevent_category_cycle();

CREATE TABLE articles (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  title VARCHAR(200) NOT NULL,
  slug VARCHAR(160) NOT NULL,
  category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
  status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
  tags TEXT[] NOT NULL DEFAULT '{}',
  content TEXT NOT NULL DEFAULT '',
  summary VARCHAR(500) NOT NULL DEFAULT '',
  published_at TIMESTAMPTZ,
  created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  updated_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ,
  deleted_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  CHECK (length(btrim(title)) > 0),
  CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
  CHECK (
    (status = 'published' AND published_at IS NOT NULL)
    OR (status = 'draft' AND published_at IS NULL)
  )
);

CREATE UNIQUE INDEX uq_articles_slug_active
  ON articles (slug)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_articles_title_active
  ON articles (lower(btrim(title)))
  WHERE deleted_at IS NULL;

CREATE INDEX idx_articles_status_updated
  ON articles (status, updated_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_articles_category_status_updated
  ON articles (category_id, status, updated_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_articles_tags_gin
  ON articles USING GIN (tags)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_articles_title_trgm
  ON articles USING GIN (title gin_trgm_ops)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_articles_summary_trgm
  ON articles USING GIN (summary gin_trgm_ops)
  WHERE deleted_at IS NULL;

ALTER TABLE articles ADD COLUMN search_vector tsvector;

CREATE OR REPLACE FUNCTION articles_search_vector_update()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('simple', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('simple', coalesce(NEW.summary, '')), 'B') ||
    setweight(to_tsvector('simple', coalesce(NEW.content, '')), 'C') ||
    setweight(to_tsvector('simple', coalesce(array_to_string(NEW.tags, ' '), '')), 'B');
  RETURN NEW;
END;
$$;

CREATE TRIGGER trg_articles_search_vector
  BEFORE INSERT OR UPDATE OF title, summary, content, tags ON articles
  FOR EACH ROW
  EXECUTE FUNCTION articles_search_vector_update();

CREATE INDEX idx_articles_search_vector
  ON articles USING GIN (search_vector)
  WHERE deleted_at IS NULL;

CREATE TABLE images (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  storage_key TEXT NOT NULL UNIQUE,
  url TEXT NOT NULL,
  mime_type VARCHAR(100) NOT NULL,
  size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
  source VARCHAR(20) NOT NULL CHECK (source IN ('upload', 'article', 'system')),
  width INTEGER CHECK (width IS NULL OR width > 0),
  height INTEGER CHECK (height IS NULL OR height > 0),
  created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ,
  deleted_by BIGINT REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_images_source_created
  ON images (source, created_at DESC)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_images_name_trgm
  ON images USING GIN (name gin_trgm_ops)
  WHERE deleted_at IS NULL;

CREATE INDEX idx_images_deleted
  ON images (deleted_at DESC)
  WHERE deleted_at IS NOT NULL;

CREATE TABLE article_images (
  article_id BIGINT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
  image_id BIGINT NOT NULL REFERENCES images(id) ON DELETE RESTRICT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (article_id, image_id)
);

CREATE INDEX IF NOT EXISTS idx_article_images_image
  ON article_images (image_id, article_id);

CREATE TABLE recycle_bin (
  id UUID PRIMARY KEY,
  item_type VARCHAR(20) NOT NULL CHECK (item_type IN ('article', 'image')),
  item_id BIGINT NOT NULL,
  snapshot JSONB NOT NULL,
  deleted_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  deleted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_recycle_bin_type_deleted
  ON recycle_bin (item_type, deleted_at DESC);

CREATE INDEX idx_recycle_bin_deleted
  ON recycle_bin (deleted_at DESC);

CREATE TABLE audit_logs (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
  action VARCHAR(80) NOT NULL,
  resource_type VARCHAR(40),
  resource_id BIGINT,
  payload JSONB,
  ip INET,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_logs_created
  ON audit_logs (created_at DESC);

CREATE INDEX idx_audit_logs_resource
  ON audit_logs (resource_type, resource_id, created_at DESC);

-- 初始系统设置。管理员用户应先插入，再将其 id 作为 updated_by/created_by。
INSERT INTO site_settings (id, site_name, site_title, logo_url)
VALUES (1, '轻文档', '轻文档 - 专注技术教程的个人文档网站', '/favicon.svg')
ON CONFLICT (id) DO NOTHING;

INSERT INTO security_settings (id)
VALUES (1)
ON CONFLICT (id) DO NOTHING;
