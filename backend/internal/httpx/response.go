package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID(c *gin.Context) string {
	if value := c.GetHeader("X-Request-ID"); value != "" {
		return value
	}
	return "req_" + uuid.NewString()
}

func OK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"code": 0, "message": "ok", "data": data, "requestId": RequestID(c)})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Error(c *gin.Context, status, code int, message string) {
	c.JSON(status, gin.H{"code": code, "message": message, "data": nil, "requestId": RequestID(c)})
}
