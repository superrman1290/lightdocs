package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/superrman1290/lightdocs/backend/internal/auth"
	"github.com/superrman1290/lightdocs/backend/internal/config"
	"github.com/superrman1290/lightdocs/backend/internal/httpx"
)

type API struct {
	DB                     *pgxpool.Pool
	Auth                   *auth.Service
	MaxUploadBytes         int64
	MaxArticleContentBytes int64
	UploadDirectory        string
}

func New(cfg config.Config, db *pgxpool.Pool) *API {
	return &API{
		DB:                     db,
		Auth:                   &auth.Service{DB: db, AccessTokenTTL: time.Duration(cfg.AccessTokenTTLMin) * time.Minute, RememberSessionDays: cfg.RememberSessionDays},
		MaxUploadBytes:         cfg.MaxUploadBytes,
		MaxArticleContentBytes: cfg.MaxArticleContentBytes,
		UploadDirectory:        cfg.UploadDirectory,
	}
}

func (a *API) Router(origins []string) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Static("/uploads", a.UploadDirectory)
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = origins
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-Reauth-Token"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowCredentials = true
	r.Use(cors.New(corsConfig))

	r.GET("/api/v1/health", func(c *gin.Context) { httpx.OK(c, http.StatusOK, gin.H{"status": "ok"}) })
	api := r.Group("/api/v1")
	api.POST("/auth/session", a.login)
	api.GET("/public/docs/:slug", a.publicDoc)
	api.GET("/public/search", a.publicSearch)
	api.GET("/public/settings/site", a.publicSiteSettings)
	api.GET("/public/categories", a.listCategories)

	protected := api.Group("")
	protected.Use(a.Auth.Middleware())
	protected.GET("/auth/me", a.Auth.Me)
	protected.GET("/dashboard/overview", a.dashboardOverview)
	protected.DELETE("/auth/session", a.Auth.Logout)
	protected.DELETE("/auth/sessions", a.Auth.RevokeAll)
	protected.POST("/auth/re-auth", a.reauth)

	protected.GET("/articles", a.listArticles)
	protected.POST("/articles", a.createArticle)
	protected.GET("/articles/:id", a.getArticle)
	protected.PATCH("/articles/:id", a.updateArticle)
	protected.DELETE("/articles/:id", a.deleteArticle)
	protected.GET("/articles/:id/export", a.exportArticle)

	protected.GET("/categories", a.listCategories)
	protected.POST("/categories", a.createCategory)
	protected.PATCH("/categories/:id", a.updateCategory)
	protected.DELETE("/categories/:id", a.deleteCategory)

	protected.GET("/images", a.listImages)
	protected.POST("/images", a.uploadImages)
	protected.DELETE("/images", a.deleteImages)

	protected.GET("/recycle-bin", a.listRecycle)
	protected.PATCH("/recycle-bin/:id", a.updateRecycle)
	protected.DELETE("/recycle-bin/:id", a.deleteRecycle)

	protected.GET("/settings/site", a.siteSettings)
	protected.PATCH("/settings/site", a.updateSiteSettings)
	protected.GET("/settings/security", a.securitySettings)
	protected.PATCH("/settings/security", a.updateSecuritySettings)
	protected.GET("/settings/admin", a.adminSettings)
	protected.PATCH("/settings/admin", a.updateAdminSettings)
	return r
}

func (a *API) login(c *gin.Context) {
	var request struct {
		Username   string `json:"username"`
		Password   string `json:"password"`
		RememberMe bool   `json:"rememberMe"`
	}
	if c.ShouldBindJSON(&request) != nil || strings.TrimSpace(request.Username) == "" || request.Password == "" {
		httpx.Error(c, http.StatusBadRequest, 40000, "账号和密码不能为空")
		return
	}
	a.Auth.Login(c, strings.TrimSpace(request.Username), request.Password, request.RememberMe)
}

func (a *API) reauth(c *gin.Context) {
	var request struct {
		CurrentPassword string `json:"currentPassword"`
	}
	if c.ShouldBindJSON(&request) != nil || request.CurrentPassword == "" {
		httpx.Error(c, http.StatusBadRequest, 40000, "当前密码不能为空")
		return
	}
	a.Auth.Reauth(c, request.CurrentPassword)
}

type articleResponse struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	CategoryID  int64      `json:"categoryId"`
	Category    string     `json:"category"`
	Status      string     `json:"status"`
	Tags        []string   `json:"tags"`
	Content     string     `json:"content"`
	Summary     string     `json:"summary"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	PublishedAt *time.Time `json:"publishedAt,omitempty"`
}

var markdownImagePattern = regexp.MustCompile(`!\[[^\]]*\]\(([^\s)]+)(?:\s+"[^"]*")?\)`)

// articleImageReferences extracts local image URLs from Markdown. External
// images remain valid Markdown but are not managed by the images table.
func articleImageReferences(content string) []string {
	seen := make(map[string]struct{})
	refs := make([]string, 0)
	for _, match := range markdownImagePattern.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		raw := strings.TrimSpace(match[1])
		parsed, err := url.Parse(raw)
		if err != nil || parsed.Path == "" || !strings.HasPrefix(parsed.Path, "/uploads/") {
			continue
		}
		path := parsed.Path
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		refs = append(refs, path)
	}
	return refs
}

func syncArticleImages(ctx context.Context, tx pgx.Tx, articleID int64, content string) error {
	refs := articleImageReferences(content)
	if _, err := tx.Exec(ctx, `DELETE FROM article_images WHERE article_id=$1`, articleID); err != nil {
		return err
	}
	if len(refs) == 0 {
		return nil
	}
	for _, ref := range refs {
		var imageID int64
		err := tx.QueryRow(ctx, `SELECT id FROM images WHERE url=$1 OR ('/uploads/' || storage_key)=$1`, ref).Scan(&imageID)
		if err == pgx.ErrNoRows {
			// A valid external or stale local URL does not prevent the article
			// from being saved; it simply has no managed image relation.
			continue
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO article_images(article_id,image_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, articleID, imageID); err != nil {
			return err
		}
	}
	return nil
}

func normalizeAndValidateTags(tags []string) ([]string, error) {
	result := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, raw := range tags {
		tag := strings.TrimSpace(raw)
		if tag == "" {
			continue
		}
		if utf8.RuneCountInString(tag) > 8 {
			return nil, fmt.Errorf("标签「%s」不能超过 8 个字符", tag)
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result, nil
}

func (a *API) validateArticleContent(content string) error {
	if int64(len([]byte(content))) > a.MaxArticleContentBytes {
		return fmt.Errorf("正文不能超过 %d MiB", a.MaxArticleContentBytes/(1024*1024))
	}
	return nil
}

func uniqueConstraintName(err error) string {
	pgErr, ok := err.(*pgconn.PgError)
	if !ok || pgErr.Code != "23505" {
		return ""
	}
	return pgErr.ConstraintName
}

func reportArticleUniqueViolation(c *gin.Context, err error) bool {
	switch uniqueConstraintName(err) {
	case "uq_articles_title_active":
		httpx.Error(c, 409, 40904, "文章标题已存在，请使用不同的标题")
		return true
	case "uq_articles_slug_active":
		httpx.Error(c, 409, 40905, "文章 slug 已存在，请修改 slug")
		return true
	default:
		return false
	}
}

var allowedImageMIMEs = map[string]struct{}{
	"image/jpeg":    {},
	"image/png":     {},
	"image/webp":    {},
	"image/gif":     {},
	"image/svg+xml": {},
}

func validateImageFile(header *multipart.FileHeader) (string, error) {
	file, err := header.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()
	detectedMIME, err := mimetype.DetectReader(file)
	if err != nil {
		return "", err
	}
	detected := detectedMIME.String()
	if _, ok := allowedImageMIMEs[detected]; !ok {
		return "", errors.New("仅支持 JPEG、PNG、WebP、GIF 或 SVG 图片")
	}
	return detected, nil
}

func scanArticle(row pgx.Row) (articleResponse, error) {
	var item articleResponse
	err := row.Scan(&item.ID, &item.Title, &item.Slug, &item.CategoryID, &item.Category, &item.Status, &item.Tags, &item.Content, &item.Summary, &item.CreatedAt, &item.UpdatedAt, &item.PublishedAt)
	return item, err
}

const articleSelect = `SELECT a.id, a.title, a.slug, a.category_id, c.name, a.status, a.tags, a.content, a.summary, a.created_at, a.updated_at, a.published_at FROM articles a JOIN categories c ON c.id = a.category_id WHERE a.deleted_at IS NULL`

func (a *API) listArticles(c *gin.Context) {
	page, pageSize := pagination(c)
	status, keyword, categoryID := c.Query("status"), strings.TrimSpace(c.Query("keyword")), c.Query("categoryId")
	if _, public := c.Get("publicSearch"); public {
		status = "published"
	}
	where := ""
	args := []any{}
	if status != "" {
		args = append(args, status)
		where += fmt.Sprintf(" AND a.status = $%d", len(args))
	}
	if categoryID != "" {
		args = append(args, categoryID)
		where += fmt.Sprintf(" AND a.category_id = $%d", len(args))
	}
	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		n := len(args)
		where += fmt.Sprintf(" AND (a.title ILIKE $%d OR a.summary ILIKE $%d OR a.content ILIKE $%d OR EXISTS (SELECT 1 FROM unnest(a.tags) tag WHERE tag ILIKE $%d))", n, n, n, n)
	}
	var total int
	if err := a.DB.QueryRow(c, "SELECT count(*) FROM articles a WHERE a.deleted_at IS NULL"+where, args...).Scan(&total); err != nil {
		httpx.Error(c, 500, 50000, "查询文章失败")
		return
	}
	args = append(args, pageSize, (page-1)*pageSize)
	query := articleSelect + where + " ORDER BY a.updated_at DESC LIMIT $" + strconv.Itoa(len(args)-1) + " OFFSET $" + strconv.Itoa(len(args))
	rows, err := a.DB.Query(c, query, args...)
	if err != nil {
		httpx.Error(c, 500, 50000, "查询文章失败")
		return
	}
	defer rows.Close()
	list := []articleResponse{}
	for rows.Next() {
		var item articleResponse
		if err := rows.Scan(&item.ID, &item.Title, &item.Slug, &item.CategoryID, &item.Category, &item.Status, &item.Tags, &item.Content, &item.Summary, &item.CreatedAt, &item.UpdatedAt, &item.PublishedAt); err != nil {
			httpx.Error(c, 500, 50000, "读取文章失败")
			return
		}
		list = append(list, item)
	}
	httpx.OK(c, 200, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
}

func (a *API) getArticle(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	item, err := scanArticle(a.DB.QueryRow(c, articleSelect+" AND a.id = $1", id))
	if err == pgx.ErrNoRows {
		httpx.Error(c, 404, 40400, "文章不存在")
		return
	}
	if err != nil {
		httpx.Error(c, 500, 50000, "获取文章失败")
		return
	}
	httpx.OK(c, 200, item)
}

func (a *API) publicDoc(c *gin.Context) {
	item, err := scanArticle(a.DB.QueryRow(c, articleSelect+" AND a.slug = $1 AND a.status = 'published'", c.Param("slug")))
	if err == pgx.ErrNoRows {
		httpx.Error(c, 404, 40400, "文档不存在")
		return
	}
	if err != nil {
		httpx.Error(c, 500, 50000, "获取文档失败")
		return
	}
	httpx.OK(c, 200, item)
}

func (a *API) dashboardOverview(c *gin.Context) {
	var articles, categories, images int
	if err := a.DB.QueryRow(c, `SELECT count(*) FROM articles WHERE deleted_at IS NULL`).Scan(&articles); err != nil {
		httpx.Error(c, 500, 50000, "获取仪表盘统计失败")
		return
	}
	if err := a.DB.QueryRow(c, `SELECT count(*) FROM categories`).Scan(&categories); err != nil {
		httpx.Error(c, 500, 50000, "获取仪表盘统计失败")
		return
	}
	if err := a.DB.QueryRow(c, `SELECT count(*) FROM images WHERE deleted_at IS NULL`).Scan(&images); err != nil {
		httpx.Error(c, 500, 50000, "获取仪表盘统计失败")
		return
	}
	rows, err := a.DB.Query(c, articleSelect+` ORDER BY a.updated_at DESC LIMIT 5`)
	if err != nil {
		httpx.Error(c, 500, 50000, "获取最新文章失败")
		return
	}
	defer rows.Close()
	latest := []articleResponse{}
	for rows.Next() {
		var item articleResponse
		if err := rows.Scan(&item.ID, &item.Title, &item.Slug, &item.CategoryID, &item.Category, &item.Status, &item.Tags, &item.Content, &item.Summary, &item.CreatedAt, &item.UpdatedAt, &item.PublishedAt); err != nil {
			httpx.Error(c, 500, 50000, "读取最新文章失败")
			return
		}
		latest = append(latest, item)
	}
	httpx.OK(c, 200, gin.H{"statistics": gin.H{"articles": articles, "categories": categories, "images": images}, "latestArticles": latest})
}

func (a *API) publicSearch(c *gin.Context) { c.Set("publicSearch", true); a.listArticles(c) }

func (a *API) createArticle(c *gin.Context) {
	var request struct {
		Title      string   `json:"title"`
		Slug       string   `json:"slug"`
		CategoryID int64    `json:"categoryId"`
		Status     string   `json:"status"`
		Tags       []string `json:"tags"`
		Content    string   `json:"content"`
		Summary    string   `json:"summary"`
	}
	if c.ShouldBindJSON(&request) != nil || strings.TrimSpace(request.Title) == "" || strings.TrimSpace(request.Slug) == "" || request.CategoryID == 0 {
		httpx.Error(c, 400, 40000, "标题、slug 和分类不能为空")
		return
	}
	normalizedTags, err := normalizeAndValidateTags(request.Tags)
	if err != nil {
		httpx.Error(c, 422, 42201, err.Error())
		return
	}
	if err = a.validateArticleContent(request.Content); err != nil {
		httpx.Error(c, 422, 42202, err.Error())
		return
	}
	request.Tags = normalizedTags
	status := request.Status
	if status == "" {
		status = "draft"
	}
	var published any
	if status == "published" {
		published = time.Now()
	}
	tx, err := a.DB.Begin(c)
	if err != nil {
		httpx.Error(c, 500, 50000, "创建文章失败")
		return
	}
	defer tx.Rollback(c)
	var item articleResponse
	err = tx.QueryRow(c, `INSERT INTO articles (title, slug, category_id, status, tags, content, summary, published_at, created_by, updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9) RETURNING id, title, slug, category_id, (SELECT name FROM categories WHERE id = category_id), status, tags, content, summary, created_at, updated_at, published_at`, request.Title, request.Slug, request.CategoryID, status, request.Tags, request.Content, request.Summary, published, c.MustGet("lightdocs.user_id")).Scan(&item.ID, &item.Title, &item.Slug, &item.CategoryID, &item.Category, &item.Status, &item.Tags, &item.Content, &item.Summary, &item.CreatedAt, &item.UpdatedAt, &item.PublishedAt)
	if err == nil {
		err = syncArticleImages(c, tx, item.ID, request.Content)
	}
	if err == nil {
		err = tx.Commit(c)
	}
	if err != nil {
		if reportArticleUniqueViolation(c, err) {
			return
		}
		httpx.Error(c, 422, 42200, "创建文章失败")
		return
	}
	httpx.OK(c, 201, item)
}

func (a *API) updateArticle(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request struct {
		Title      *string  `json:"title"`
		Slug       *string  `json:"slug"`
		CategoryID *int64   `json:"categoryId"`
		Status     *string  `json:"status"`
		Tags       []string `json:"tags"`
		Content    *string  `json:"content"`
		Summary    *string  `json:"summary"`
	}
	if c.ShouldBindJSON(&request) != nil {
		httpx.Error(c, 400, 40000, "请求参数错误")
		return
	}
	if request.Tags != nil {
		normalizedTags, err := normalizeAndValidateTags(request.Tags)
		if err != nil {
			httpx.Error(c, 422, 42201, err.Error())
			return
		}
		request.Tags = normalizedTags
	}
	if request.Content != nil {
		if err := a.validateArticleContent(*request.Content); err != nil {
			httpx.Error(c, 422, 42202, err.Error())
			return
		}
	}
	tx, err := a.DB.Begin(c)
	if err != nil {
		httpx.Error(c, 500, 50000, "更新文章失败")
		return
	}
	defer tx.Rollback(c)
	var item articleResponse
	err = tx.QueryRow(c, `UPDATE articles SET title=COALESCE($1,title), slug=COALESCE($2,slug), category_id=COALESCE($3,category_id), status=COALESCE($4,status), tags=COALESCE($5,tags), content=COALESCE($6,content), summary=COALESCE($7,summary), published_at=CASE WHEN $4='published' THEN COALESCE(published_at,now()) WHEN $4='draft' THEN NULL ELSE published_at END, updated_by=$8, updated_at=now() WHERE id=$9 AND deleted_at IS NULL RETURNING id,title,slug,category_id,(SELECT name FROM categories WHERE id=category_id),status,tags,content,summary,created_at,updated_at,published_at`, request.Title, request.Slug, request.CategoryID, request.Status, request.Tags, request.Content, request.Summary, c.MustGet("lightdocs.user_id"), id).Scan(&item.ID, &item.Title, &item.Slug, &item.CategoryID, &item.Category, &item.Status, &item.Tags, &item.Content, &item.Summary, &item.CreatedAt, &item.UpdatedAt, &item.PublishedAt)
	if err == nil && request.Content != nil {
		err = syncArticleImages(c, tx, item.ID, *request.Content)
	}
	if err == nil {
		err = tx.Commit(c)
	}
	if err == pgx.ErrNoRows {
		httpx.Error(c, 404, 40400, "文章不存在")
		return
	}
	if err != nil {
		if reportArticleUniqueViolation(c, err) {
			return
		}
		httpx.Error(c, 422, 42200, "更新文章失败")
		return
	}
	httpx.OK(c, 200, item)
}
func (a *API) deleteArticle(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	tx, err := a.DB.Begin(c)
	if err != nil {
		httpx.Error(c, 500, 50000, "删除文章失败")
		return
	}
	defer tx.Rollback(c)
	var snapshot []byte
	if err = tx.QueryRow(c, `SELECT jsonb_build_object('id',id,'title',title,'slug',slug,'categoryId',category_id,'status',status,'tags',tags,'content',content,'summary',summary,'publishedAt',published_at,'article_image_ids',(SELECT COALESCE(jsonb_agg(image_id),'[]'::jsonb) FROM article_images WHERE article_id=articles.id)) FROM articles WHERE id=$1 AND deleted_at IS NULL`, id).Scan(&snapshot); err != nil {
		httpx.Error(c, 404, 40400, "文章不存在")
		return
	}
	_, err = tx.Exec(c, `INSERT INTO recycle_bin(id,item_type,item_id,snapshot,deleted_by) VALUES($1,'article',$2,$3,$4)`, uuid.New(), id, snapshot, c.MustGet("lightdocs.user_id"))
	if err == nil {
		_, err = tx.Exec(c, `UPDATE articles SET deleted_at=now(),deleted_by=$1,updated_at=now() WHERE id=$2`, c.MustGet("lightdocs.user_id"), id)
	}
	if err != nil {
		httpx.Error(c, 500, 50000, "删除文章失败")
		return
	}
	if err = tx.Commit(c); err != nil {
		httpx.Error(c, 500, 50000, "删除文章失败")
		return
	}
	httpx.NoContent(c)
}
func (a *API) exportArticle(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	format := c.DefaultQuery("format", "markdown")
	if format != "markdown" && format != "html" {
		httpx.Error(c, 400, 40000, "不支持的导出格式")
		return
	}
	var title, content string
	if err := a.DB.QueryRow(c, `SELECT title,content FROM articles WHERE id=$1 AND deleted_at IS NULL`, id).Scan(&title, &content); err != nil {
		httpx.Error(c, 404, 40400, "文章不存在")
		return
	}
	filename := strings.ReplaceAll(strings.ToLower(title), " ", "-")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`.md"`)
	c.Data(200, "text/markdown; charset=utf-8", []byte(content))
}

type categoryResponse struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ParentID *int64 `json:"parentId"`
	Sort     int    `json:"sort"`
}

func (a *API) listCategories(c *gin.Context) {
	rows, err := a.DB.Query(c, `SELECT id,name,parent_id,sort_order FROM categories ORDER BY parent_id NULLS FIRST, sort_order, id`)
	if err != nil {
		httpx.Error(c, 500, 50000, "查询分类失败")
		return
	}
	defer rows.Close()
	list := []categoryResponse{}
	for rows.Next() {
		var item categoryResponse
		if rows.Scan(&item.ID, &item.Name, &item.ParentID, &item.Sort) != nil {
			httpx.Error(c, 500, 50000, "读取分类失败")
			return
		}
		list = append(list, item)
	}
	httpx.OK(c, 200, list)
}
func (a *API) createCategory(c *gin.Context) {
	var request struct {
		Name     string `json:"name"`
		ParentID *int64 `json:"parentId"`
		Sort     int    `json:"sort"`
	}
	if c.ShouldBindJSON(&request) != nil || strings.TrimSpace(request.Name) == "" {
		httpx.Error(c, 400, 40000, "分类名称不能为空")
		return
	}
	var item categoryResponse
	err := a.DB.QueryRow(c, `INSERT INTO categories(name,parent_id,sort_order,created_by,updated_by) VALUES($1,$2,$3,$4,$4) RETURNING id,name,parent_id,sort_order`, request.Name, request.ParentID, request.Sort, c.MustGet("lightdocs.user_id")).Scan(&item.ID, &item.Name, &item.ParentID, &item.Sort)
	if err != nil {
		httpx.Error(c, 409, 40901, "分类创建失败")
		return
	}
	httpx.OK(c, 201, item)
}
func (a *API) updateCategory(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var request struct {
		Name     *string `json:"name"`
		ParentID *int64  `json:"parentId"`
		Sort     *int    `json:"sort"`
	}
	if c.ShouldBindJSON(&request) != nil {
		httpx.Error(c, 400, 40000, "请求参数错误")
		return
	}
	var item categoryResponse
	err := a.DB.QueryRow(c, `UPDATE categories SET name=COALESCE($1,name),parent_id=COALESCE($2,parent_id),sort_order=COALESCE($3,sort_order),updated_by=$4,updated_at=now() WHERE id=$5 RETURNING id,name,parent_id,sort_order`, request.Name, request.ParentID, request.Sort, c.MustGet("lightdocs.user_id"), id).Scan(&item.ID, &item.Name, &item.ParentID, &item.Sort)
	if err != nil {
		httpx.Error(c, 409, 40901, "更新分类失败")
		return
	}
	httpx.OK(c, 200, item)
}
func (a *API) deleteCategory(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var count int
	if err := a.DB.QueryRow(c, `SELECT count(*) FROM categories WHERE parent_id=$1`, id).Scan(&count); err != nil || count > 0 {
		httpx.Error(c, 409, 40901, "分类存在子分类或文章引用")
		return
	}
	var articles int
	if err := a.DB.QueryRow(c, `SELECT count(*) FROM articles WHERE category_id=$1`, id).Scan(&articles); err != nil || articles > 0 {
		httpx.Error(c, 409, 40901, "分类存在文章引用（包括回收站中的文章）")
		return
	}
	result, err := a.DB.Exec(c, `DELETE FROM categories WHERE id=$1`, id)
	if err != nil {
		httpx.Error(c, 409, 40901, "分类存在关联数据")
		return
	}
	if result.RowsAffected() == 0 {
		httpx.Error(c, 404, 40400, "分类不存在")
		return
	}
	httpx.NoContent(c)
}

func (a *API) listImages(c *gin.Context) {
	page, size := pagination(c)
	var total int
	keyword := strings.TrimSpace(c.Query("keyword"))
	source := c.Query("source")
	args := []any{}
	where := " deleted_at IS NULL"
	if keyword != "" {
		args = append(args, "%"+keyword+"%")
		where += fmt.Sprintf(" AND name ILIKE $%d", len(args))
	}
	if source != "" {
		args = append(args, source)
		where += fmt.Sprintf(" AND source=$%d", len(args))
	}
	if a.DB.QueryRow(c, "SELECT count(*) FROM images WHERE"+where, args...).Scan(&total) != nil {
		httpx.Error(c, 500, 50000, "查询图片失败")
		return
	}
	args = append(args, size, (page-1)*size)
	rows, err := a.DB.Query(c, "SELECT i.id,i.name,i.url,ceil(i.size_bytes/1024.0),i.size_bytes,i.mime_type,i.storage_key,i.width,i.height,i.source,i.created_at,EXISTS (SELECT 1 FROM article_images ai WHERE ai.image_id=i.id) FROM images i WHERE"+where+" ORDER BY i.created_at DESC LIMIT $"+strconv.Itoa(len(args)-1)+" OFFSET $"+strconv.Itoa(len(args)), args...)
	if err != nil {
		httpx.Error(c, 500, 50000, "查询图片失败")
		return
	}
	defer rows.Close()
	list := []gin.H{}
	for rows.Next() {
		var id int64
		var name, url, mime, key, sourceValue string
		var size int64
		var sizeBytes int64
		var width, height *int
		var created time.Time
		var referenced bool
		if rows.Scan(&id, &name, &url, &size, &sizeBytes, &mime, &key, &width, &height, &sourceValue, &created, &referenced) != nil {
			httpx.Error(c, 500, 50000, "读取图片失败")
			return
		}
		list = append(list, gin.H{"id": id, "name": name, "url": url, "size": size, "sizeBytes": sizeBytes, "mimeType": mime, "storageKey": key, "width": width, "height": height, "source": sourceValue, "createdAt": created, "referenced": referenced})
	}
	httpx.OK(c, 200, gin.H{"list": list, "total": total, "page": page, "pageSize": size})
}
func (a *API) uploadImages(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		httpx.Error(c, 400, 40000, "上传文件不能为空")
		return
	}
	list := []gin.H{}
	for _, headers := range form.File {
		for _, header := range headers {
			detectedMIME, err := validateImageFile(header)
			if err != nil {
				httpx.Error(c, 400, 40001, err.Error())
				return
			}
			key, err := a.uploadToDisk(header)
			if err != nil {
				httpx.Error(c, 413, 41300, err.Error())
				return
			}
			var id int64
			var created time.Time
			mime := detectedMIME
			err = a.DB.QueryRow(c, `INSERT INTO images(name,storage_key,url,mime_type,size_bytes,source,created_by) VALUES($1,$2,$3,$4,$5,'upload',$6) RETURNING id,created_at`, header.Filename, key, "/uploads/"+key, mime, header.Size, c.MustGet("lightdocs.user_id")).Scan(&id, &created)
			if err != nil {
				httpx.Error(c, 500, 50000, "保存图片失败")
				return
			}
			list = append(list, gin.H{"id": id, "name": header.Filename, "url": "/uploads/" + key, "size": (header.Size + 1023) / 1024, "sizeBytes": header.Size, "mimeType": mime, "storageKey": key, "source": "upload", "createdAt": created})
		}
	}
	httpx.OK(c, 201, list)
}
func (a *API) deleteImages(c *gin.Context) {
	ids := strings.Split(c.Query("ids"), ",")
	if len(ids) == 0 || ids[0] == "" {
		httpx.Error(c, 400, 40000, "ids 不能为空")
		return
	}
	for _, raw := range ids {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			continue
		}
		var referenced bool
		if err = a.DB.QueryRow(c, `SELECT EXISTS (SELECT 1 FROM article_images WHERE image_id=$1)`, id).Scan(&referenced); err != nil {
			httpx.Error(c, 500, 50000, "删除图片失败")
			return
		}
		if referenced {
			httpx.Error(c, 409, 40902, "图片正在被文章引用，请先移除文章中的图片")
			return
		}
		var snapshot []byte
		if err = a.DB.QueryRow(c, `SELECT jsonb_build_object('id',id,'name',name,'url',url,'sizeBytes',size_bytes,'mimeType',mime_type,'storageKey',storage_key,'source',source,'width',width,'height',height,'createdAt',created_at) FROM images WHERE id=$1 AND deleted_at IS NULL`, id).Scan(&snapshot); err != nil {
			continue
		}
		if _, err = a.DB.Exec(c, `INSERT INTO recycle_bin(id,item_type,item_id,snapshot,deleted_by) VALUES($1,'image',$2,$3,$4)`, uuid.New(), id, snapshot, c.MustGet("lightdocs.user_id")); err != nil {
			httpx.Error(c, 500, 50000, "删除图片失败")
			return
		}
		if _, err = a.DB.Exec(c, `UPDATE images SET deleted_at=now(),deleted_by=$1 WHERE id=$2`, c.MustGet("lightdocs.user_id"), id); err != nil {
			httpx.Error(c, 500, 50000, "删除图片失败")
			return
		}
	}
	httpx.NoContent(c)
}
func (a *API) listRecycle(c *gin.Context) {
	page, size := pagination(c)
	itemType := c.Query("type")
	args := []any{}
	where := ""
	if itemType != "" {
		args = append(args, itemType)
		where = " WHERE item_type=$1"
	}
	var total int
	if err := a.DB.QueryRow(c, "SELECT count(*) FROM recycle_bin"+where, args...).Scan(&total); err != nil {
		httpx.Error(c, 500, 50000, "查询回收站失败")
		return
	}
	args = append(args, size, (page-1)*size)
	rows, err := a.DB.Query(c, "SELECT id,item_type,item_id,snapshot,deleted_at FROM recycle_bin"+where+" ORDER BY deleted_at DESC LIMIT $"+strconv.Itoa(len(args)-1)+" OFFSET $"+strconv.Itoa(len(args)), args...)
	if err != nil {
		httpx.Error(c, 500, 50000, "查询回收站失败")
		return
	}
	defer rows.Close()
	list := []gin.H{}
	for rows.Next() {
		var id uuid.UUID
		var typ string
		var itemID int64
		var snapshot []byte
		var deletedAt time.Time
		if rows.Scan(&id, &typ, &itemID, &snapshot, &deletedAt) != nil {
			httpx.Error(c, 500, 50000, "读取回收站失败")
			return
		}
		var data any
		_ = json.Unmarshal(snapshot, &data)
		list = append(list, gin.H{"id": id.String(), "type": typ, "itemId": itemID, "data": data, "deletedAt": deletedAt})
	}
	httpx.OK(c, 200, gin.H{"list": list, "total": total, "page": page, "pageSize": size})
}
func (a *API) updateRecycle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, 400, 40000, "回收站 ID 无效")
		return
	}
	var request struct {
		Action string `json:"action"`
	}
	if c.ShouldBindJSON(&request) != nil || request.Action != "restore" {
		httpx.Error(c, 400, 40000, "只支持 restore 操作")
		return
	}
	var itemType string
	var itemID int64
	var snapshot []byte
	var deletedAt time.Time
	if err = a.DB.QueryRow(c, `SELECT item_type,item_id,snapshot,deleted_at FROM recycle_bin WHERE id=$1`, id).Scan(&itemType, &itemID, &snapshot, &deletedAt); err != nil {
		httpx.Error(c, 404, 40400, "回收站项目不存在")
		return
	}
	var table string
	if itemType == "article" {
		table = "articles"
	} else if itemType == "image" {
		table = "images"
	} else {
		httpx.Error(c, 422, 42200, "项目类型无效")
		return
	}
	tx, err := a.DB.Begin(c)
	if err != nil {
		httpx.Error(c, 500, 50000, "恢复失败")
		return
	}
	defer tx.Rollback(c)
	if _, err = tx.Exec(c, "UPDATE "+table+" SET deleted_at=NULL,deleted_by=NULL WHERE id=$1", itemID); err != nil {
		httpx.Error(c, 409, 40903, "恢复失败")
		return
	}
	if itemType == "article" {
		var articleSnapshot struct {
			ArticleImageIDs []int64 `json:"article_image_ids"`
		}
		if err = json.Unmarshal(snapshot, &articleSnapshot); err != nil {
			httpx.Error(c, 500, 50000, "恢复失败，文章快照无效")
			return
		}
		for _, imageID := range articleSnapshot.ArticleImageIDs {
			if _, err = tx.Exec(c, `INSERT INTO article_images(article_id,image_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, itemID, imageID); err != nil {
				httpx.Error(c, 409, 40903, "恢复失败，文章图片关系无效")
				return
			}
		}
	}
	if _, err = tx.Exec(c, `DELETE FROM recycle_bin WHERE id=$1`, id); err != nil {
		httpx.Error(c, 500, 50000, "恢复失败，回收站记录未删除")
		return
	}
	if err = tx.Commit(c); err != nil {
		httpx.Error(c, 500, 50000, "恢复失败")
		return
	}
	var result any
	_ = json.Unmarshal(snapshot, &result)
	httpx.OK(c, 200, gin.H{
		"id":        id.String(),
		"type":      itemType,
		"itemId":    itemID,
		"data":      result,
		"deletedAt": deletedAt,
	})
}
func (a *API) deleteRecycle(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpx.Error(c, 400, 40000, "回收站 ID 无效")
		return
	}
	var itemType string
	var itemID int64
	if err = a.DB.QueryRow(c, `SELECT item_type,item_id FROM recycle_bin WHERE id=$1`, id).Scan(&itemType, &itemID); err != nil {
		httpx.Error(c, 404, 40400, "回收站项目不存在")
		return
	}
	table := map[string]string{"article": "articles", "image": "images"}[itemType]
	if table == "" {
		httpx.Error(c, 422, 42200, "回收站项目类型无效")
		return
	}
	if _, err = a.DB.Exec(c, "DELETE FROM "+table+" WHERE id=$1", itemID); err != nil {
		httpx.Error(c, 409, 40902, "资源仍被引用，无法永久删除")
		return
	}
	if _, err = a.DB.Exec(c, `DELETE FROM recycle_bin WHERE id=$1`, id); err != nil {
		httpx.Error(c, 500, 50000, "删除回收站项目失败")
		return
	}
	httpx.NoContent(c)
}
func (a *API) deleteRecycleRecord(c *gin.Context, id uuid.UUID) error {
	var itemType string
	var itemID int64
	if err := a.DB.QueryRow(c, `SELECT item_type,item_id FROM recycle_bin WHERE id=$1`, id).Scan(&itemType, &itemID); err != nil {
		return fmt.Errorf("回收站项目不存在")
	}
	table := map[string]string{"article": "articles", "image": "images"}[itemType]
	if table == "" {
		return fmt.Errorf("回收站项目类型无效")
	}
	if _, err := a.DB.Exec(c, "DELETE FROM "+table+" WHERE id=$1", itemID); err != nil {
		return fmt.Errorf("资源仍被引用，无法永久删除: %w", err)
	}
	if _, err := a.DB.Exec(c, `DELETE FROM recycle_bin WHERE id=$1`, id); err != nil {
		return fmt.Errorf("删除回收站项目失败: %w", err)
	}
	return nil
}

func (a *API) siteSettings(c *gin.Context) {
	var item struct {
		SiteName  string `json:"siteName"`
		SiteTitle string `json:"siteTitle"`
		LogoURL   string `json:"logoUrl"`
	}
	err := a.DB.QueryRow(c, `SELECT site_name,site_title,logo_url FROM site_settings WHERE id=1`).Scan(&item.SiteName, &item.SiteTitle, &item.LogoURL)
	if err != nil {
		httpx.Error(c, 500, 50000, "获取站点设置失败")
		return
	}
	httpx.OK(c, 200, item)
}
func (a *API) publicSiteSettings(c *gin.Context) { a.siteSettings(c) }
func (a *API) updateSiteSettings(c *gin.Context) {
	if !a.requireReauth(c) {
		return
	}
	var request struct {
		SiteName  *string `json:"siteName"`
		SiteTitle *string `json:"siteTitle"`
		LogoURL   *string `json:"logoUrl"`
	}
	if c.ShouldBindJSON(&request) != nil {
		httpx.Error(c, 400, 40000, "请求参数错误")
		return
	}
	if _, err := a.DB.Exec(c, `UPDATE site_settings SET site_name=COALESCE($1,site_name),site_title=COALESCE($2,site_title),logo_url=COALESCE($3,logo_url),updated_by=$4,updated_at=now() WHERE id=1`, request.SiteName, request.SiteTitle, request.LogoURL, c.MustGet("lightdocs.user_id")); err != nil {
		httpx.Error(c, 422, 42200, "更新站点设置失败")
		return
	}
	a.siteSettings(c)
}
func (a *API) securitySettings(c *gin.Context) {
	var item gin.H
	var max, lock, days int
	var remember, logout, reauth bool
	err := a.DB.QueryRow(c, `SELECT max_login_failures,lock_minutes,session_days,remember_login,logout_other_devices_on_password_change,require_password_reauth_on_save FROM security_settings WHERE id=1`).Scan(&max, &lock, &days, &remember, &logout, &reauth)
	if err != nil {
		httpx.Error(c, 500, 50000, "获取安全设置失败")
		return
	}
	item = gin.H{"maxLoginFailures": max, "lockMinutes": lock, "sessionDays": days, "rememberLogin": remember, "logoutOtherDevicesOnPasswordChange": logout, "requirePasswordReauthOnSave": reauth}
	httpx.OK(c, 200, item)
}
func (a *API) updateSecuritySettings(c *gin.Context) {
	if !a.requireReauth(c) {
		return
	}
	var request struct {
		Max      int  `json:"maxLoginFailures"`
		Lock     int  `json:"lockMinutes"`
		Days     int  `json:"sessionDays"`
		Remember bool `json:"rememberLogin"`
		Logout   bool `json:"logoutOtherDevicesOnPasswordChange"`
		Reauth   bool `json:"requirePasswordReauthOnSave"`
	}
	if c.ShouldBindJSON(&request) != nil {
		httpx.Error(c, 400, 40000, "请求参数错误")
		return
	}
	if _, err := a.DB.Exec(c, `UPDATE security_settings SET max_login_failures=$1,lock_minutes=$2,session_days=$3,remember_login=$4,logout_other_devices_on_password_change=$5,require_password_reauth_on_save=$6,updated_by=$7,updated_at=now() WHERE id=1`, request.Max, request.Lock, request.Days, request.Remember, request.Logout, request.Reauth, c.MustGet("lightdocs.user_id")); err != nil {
		httpx.Error(c, 422, 42200, "更新安全设置失败")
		return
	}
	a.securitySettings(c)
}
func (a *API) adminSettings(c *gin.Context) {
	var id int64
	var username string
	if err := a.DB.QueryRow(c, `SELECT id,username FROM users WHERE id=$1`, c.MustGet("lightdocs.user_id")).Scan(&id, &username); err != nil {
		httpx.Error(c, 404, 40400, "管理员不存在")
		return
	}
	// Never expose a password, hash suffix, or any other password-derived value.
	httpx.OK(c, 200, gin.H{"id": id, "username": username})
}
func (a *API) updateAdminSettings(c *gin.Context) {
	if !a.requireReauth(c) {
		return
	}
	var request struct {
		Username string `json:"username"`
		Current  string `json:"currentPassword"`
		New      string `json:"newPassword"`
		Confirm  string `json:"confirmPassword"`
	}
	if c.ShouldBindJSON(&request) != nil || request.New == "" || request.New != request.Confirm {
		httpx.Error(c, 422, 42200, "密码参数错误")
		return
	}
	var hash string
	if err := a.DB.QueryRow(c, `SELECT password_hash FROM users WHERE id=$1`, c.MustGet("lightdocs.user_id")).Scan(&hash); err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(request.Current)) != nil {
		httpx.Error(c, 401, 40101, "当前密码不正确")
		return
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(request.New), bcrypt.DefaultCost)
	if err != nil {
		httpx.Error(c, 500, 50000, "密码更新失败")
		return
	}
	if request.Username == "" {
		request.Username = "admin"
	}
	var logoutOtherDevices bool
	if err = a.DB.QueryRow(c, `SELECT logout_other_devices_on_password_change FROM security_settings WHERE id=1`).Scan(&logoutOtherDevices); err != nil {
		httpx.Error(c, 500, 50000, "读取会话安全设置失败")
		return
	}
	currentTokenHash := authTokenHash(c)
	tx, err := a.DB.Begin(c)
	if err == nil {
		defer tx.Rollback(c)
	}
	if err == nil {
		_, err = tx.Exec(c, `UPDATE users SET username=$1,password_hash=$2,updated_at=now() WHERE id=$3`, request.Username, string(newHash), c.MustGet("lightdocs.user_id"))
	}
	if err == nil && logoutOtherDevices {
		_, err = tx.Exec(c, `UPDATE auth_sessions SET revoked_at=now() WHERE user_id=$1 AND token_hash<>$2 AND revoked_at IS NULL`, c.MustGet("lightdocs.user_id"), currentTokenHash)
	}
	if err == nil {
		err = tx.Commit(c)
	} else {
		_ = tx.Rollback(c)
	}
	if err != nil {
		httpx.Error(c, 409, 40901, "账号更新失败")
		return
	}
	a.adminSettings(c)
}

func authTokenHash(c *gin.Context) string {
	header := strings.TrimSpace(c.GetHeader("Authorization"))
	return auth.TokenHash(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
}

// requireReauth enforces the current security policy for sensitive settings.
// The supplied token is consumed on success and cannot be replayed.
func (a *API) requireReauth(c *gin.Context) bool {
	var required bool
	if err := a.DB.QueryRow(c, `SELECT require_password_reauth_on_save FROM security_settings WHERE id=1`).Scan(&required); err != nil {
		httpx.Error(c, 500, 50000, "读取重新验证设置失败")
		return false
	}
	if !required {
		return true
	}
	if a.Auth.ValidateReauth(c, c.GetHeader("X-Reauth-Token")) {
		return true
	}
	httpx.Error(c, 401, 40103, "请先重新验证密码")
	return false
}

func pagination(c *gin.Context) (int, int) {
	page := intQuery(c.Query("page"), 1)
	size := intQuery(c.Query("pageSize"), 10)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	if size > 100 {
		size = 100
	}
	return page, size
}
func intQuery(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
func idParam(c *gin.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil {
		httpx.Error(c, 400, 40000, "ID 无效")
		return 0, false
	}
	return id, true
}
func (a *API) uploadToDisk(fileHeader *multipart.FileHeader) (string, error) {
	if fileHeader.Size > a.MaxUploadBytes {
		return "", fmt.Errorf("文件过大")
	}
	if err := os.MkdirAll(a.UploadDirectory, 0755); err != nil {
		return "", err
	}
	key := uuid.NewString() + filepath.Ext(fileHeader.Filename)
	dst, err := os.Create(filepath.Join(a.UploadDirectory, key))
	if err != nil {
		return "", err
	}
	defer dst.Close()
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	_, err = io.Copy(dst, src)
	return key, err
}

var _ = context.Background
