package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/superrman1290/lightdocs/backend/internal/httpx"
)

const userKey = "lightdocs.user_id"

type Service struct {
	DB                  *pgxpool.Pool
	AccessTokenTTL      time.Duration
	RememberSessionDays int
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func (s *Service) Login(c *gin.Context, username, password string, remember bool) {
	ctx := c.Request.Context()
	var id int64
	var hash, role, status string
	var failed int
	var lockedUntil *time.Time
	err := s.DB.QueryRow(ctx, `
		SELECT id, password_hash, role, status, failed_login_count, locked_until
		FROM users WHERE username = $1`, username).Scan(&id, &hash, &role, &status, &failed, &lockedUntil)
	if err != nil || status != "active" {
		httpx.Error(c, http.StatusUnauthorized, 40101, "账号或密码不正确")
		return
	}

	now := time.Now()
	if lockedUntil != nil && lockedUntil.After(now) {
		httpx.Error(c, http.StatusLocked, 40102, "账号已锁定，请稍后重试")
		return
	}

	// A lock is temporary. Once it has elapsed, start a fresh failure window.
	if lockedUntil != nil {
		if _, err = s.DB.Exec(ctx, `UPDATE users SET failed_login_count = 0, locked_until = NULL, updated_at = now() WHERE id = $1`, id); err != nil {
			httpx.Error(c, http.StatusInternalServerError, 50000, "登录失败")
			return
		}
		failed = 0
		lockedUntil = nil
	}

	var maxFailures, lockMinutes int
	if err = s.DB.QueryRow(ctx, `SELECT max_login_failures, lock_minutes FROM security_settings WHERE id = 1`).Scan(&maxFailures, &lockMinutes); err != nil {
		httpx.Error(c, http.StatusInternalServerError, 50000, "读取登录安全设置失败")
		return
	}
	if maxFailures < 1 {
		maxFailures = 5
	}
	if lockMinutes < 1 {
		lockMinutes = 15
	}

	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		var nextFailed int
		var nextLockedUntil *time.Time
		err = s.DB.QueryRow(ctx, `
			UPDATE users
			SET failed_login_count = failed_login_count + 1,
			    locked_until = CASE
				  WHEN failed_login_count + 1 >= $1 THEN now() + make_interval(mins => $2)
				  ELSE NULL
			    END,
			    updated_at = now()
			WHERE id = $3
			RETURNING failed_login_count, locked_until`, maxFailures, lockMinutes, id).Scan(&nextFailed, &nextLockedUntil)
		if err != nil {
			httpx.Error(c, http.StatusInternalServerError, 50000, "登录失败")
			return
		}
		if nextFailed >= maxFailures && nextLockedUntil != nil {
			httpx.Error(c, http.StatusLocked, 40102, "账号已锁定，请稍后重试")
			return
		}
		httpx.Error(c, http.StatusUnauthorized, 40101, "账号或密码不正确")
		return
	}

	ttl := s.AccessTokenTTL
	if remember {
		ttl = time.Duration(s.RememberSessionDays) * 24 * time.Hour
	}
	token, err := randomToken()
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, 50000, "无法创建登录会话")
		return
	}
	expires := time.Now().Add(ttl)
	_, err = s.DB.Exec(ctx, `
		INSERT INTO auth_sessions (id, user_id, token_hash, expires_at, ip, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)`, uuid.New(), id, hashToken(token), expires, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, 50000, "无法创建登录会话")
		return
	}
	_, _ = s.DB.Exec(ctx, `UPDATE users SET failed_login_count = 0, locked_until = NULL, last_login_at = now(), updated_at = now() WHERE id = $1`, id)
	httpx.OK(c, http.StatusOK, gin.H{"accessToken": token, "tokenType": "Bearer", "expiresAt": expires.UTC(), "user": User{ID: id, Username: username, Role: role}})
}

func (s *Service) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if !strings.HasPrefix(header, "Bearer ") {
			httpx.Error(c, http.StatusUnauthorized, 40100, "未登录或令牌无效")
			c.Abort()
			return
		}
		var userID int64
		err := s.DB.QueryRow(c.Request.Context(), `
			SELECT user_id FROM auth_sessions
			WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > now()`, hashToken(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))).Scan(&userID)
		if err != nil {
			httpx.Error(c, http.StatusUnauthorized, 40100, "未登录或令牌无效")
			c.Abort()
			return
		}
		c.Set(userKey, userID)
		c.Next()
	}
}

func (s *Service) Me(c *gin.Context) {
	var user User
	if err := s.DB.QueryRow(c.Request.Context(), `SELECT id, username, role FROM users WHERE id = $1`, c.MustGet(userKey)).Scan(&user.ID, &user.Username, &user.Role); err != nil {
		httpx.Error(c, http.StatusUnauthorized, 40100, "用户不存在")
		return
	}
	httpx.OK(c, http.StatusOK, user)
}

func (s *Service) Logout(c *gin.Context) {
	header := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	_, _ = s.DB.Exec(c.Request.Context(), `UPDATE auth_sessions SET revoked_at = now() WHERE token_hash = $1`, hashToken(header))
	httpx.NoContent(c)
}

func (s *Service) RevokeAll(c *gin.Context) {
	userID := c.MustGet(userKey)
	_, _ = s.DB.Exec(c.Request.Context(), `UPDATE auth_sessions SET revoked_at = now() WHERE user_id = $1 AND revoked_at IS NULL`, userID)
	httpx.OK(c, http.StatusOK, gin.H{"forceReLogin": true})
}

func (s *Service) Reauth(c *gin.Context, currentPassword string) {
	var hash string
	if err := s.DB.QueryRow(c.Request.Context(), `SELECT password_hash FROM users WHERE id = $1`, c.MustGet(userKey)).Scan(&hash); err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(currentPassword)) != nil {
		httpx.Error(c, http.StatusUnauthorized, 40101, "当前密码不正确")
		return
	}
	token, err := randomToken()
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, 50000, "无法创建验证凭证")
		return
	}
	expires := time.Now().Add(5 * time.Minute).UTC()
	tx, err := s.DB.Begin(c.Request.Context())
	if err == nil {
		_, err = tx.Exec(c.Request.Context(), `DELETE FROM reauth_tokens WHERE user_id = $1 AND expires_at <= now()`, c.MustGet(userKey))
	}
	if err == nil {
		_, err = tx.Exec(c.Request.Context(), `INSERT INTO reauth_tokens (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`, uuid.New(), c.MustGet(userKey), hashToken(token), expires)
	}
	if err == nil {
		err = tx.Commit(c.Request.Context())
	} else {
		_ = tx.Rollback(c.Request.Context())
	}
	if err != nil {
		httpx.Error(c, http.StatusInternalServerError, 50000, "无法保存验证凭证")
		return
	}
	httpx.OK(c, http.StatusOK, gin.H{"reauthToken": token, "expiresAt": expires})
}

// ValidateReauth consumes a short-lived re-authentication token exactly once.
func (s *Service) ValidateReauth(c *gin.Context, token string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}
	var id uuid.UUID
	err := s.DB.QueryRow(c.Request.Context(), `
		DELETE FROM reauth_tokens
		WHERE user_id = $1 AND token_hash = $2 AND expires_at > now()
		RETURNING id`, c.MustGet(userKey), hashToken(token)).Scan(&id)
	return err == nil
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "ld_at_" + hex.EncodeToString(buf), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// TokenHash returns the database representation of a bearer token. It is
// exposed to the API package only for operations that must preserve the
// current session while revoking other sessions.
func TokenHash(token string) string {
	return hashToken(token)
}
