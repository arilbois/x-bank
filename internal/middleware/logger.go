package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a structured request-logger middleware.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()
		latency := time.Since(start)
		full := path
		if raw != "" {
			full = path + "?" + raw
		}
		slog.Info("request",
			"method", c.Request.Method,
			"path", full,
			"status", c.Writer.Status(),
			"size", c.Writer.Size(),
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
