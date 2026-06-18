package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/arilbois/x-bank/internal/services/auth"
)

// UserContextKey is the gin context key holding the authenticated claims.
const UserContextKey = "auth.user"

// RequireAuth verifies the Bearer token and stores the claims in the
// gin context.
func RequireAuth(svc *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}
		claims, err := svc.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.Set(UserContextKey, claims)
		c.Next()
	}
}

// RequireRole enforces that the authenticated user has the given role.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(UserContextKey)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no user in context"})
			return
		}
		claims, ok := v.(*auth.Claims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "bad user in context"})
			return
		}
		if claims.Role != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// ClaimsFromContext extracts the auth claims (or nil) from a gin context.
func ClaimsFromContext(c *gin.Context) *auth.Claims {
	v, ok := c.Get(UserContextKey)
	if !ok {
		return nil
	}
	cl, _ := v.(*auth.Claims)
	return cl
}

// FromContext is a context.Context-aware helper for downstream services.
func FromContext(ctx context.Context) *auth.Claims {
	cl, _ := ctx.Value(UserContextKey).(*auth.Claims)
	return cl
}
