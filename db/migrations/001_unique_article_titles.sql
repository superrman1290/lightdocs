-- Preserve existing duplicate articles by giving later rows an explicit copy
-- suffix before enforcing title uniqueness for active articles.
WITH ranked AS (
  SELECT id, title,
         row_number() OVER (
           PARTITION BY lower(btrim(title))
           ORDER BY created_at, id
         ) AS duplicate_number
  FROM articles
  WHERE deleted_at IS NULL
), renamed AS (
  SELECT id,
         left(btrim(title), 180) || '（副本 ' || duplicate_number || '）' AS next_title
  FROM ranked
  WHERE duplicate_number > 1
)
UPDATE articles AS a
SET title = renamed.next_title,
    updated_at = now()
FROM renamed
WHERE a.id = renamed.id;

CREATE UNIQUE INDEX IF NOT EXISTS uq_articles_title_active
  ON articles (lower(btrim(title)))
  WHERE deleted_at IS NULL;
