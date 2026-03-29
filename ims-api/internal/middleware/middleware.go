package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	bizErr "ims-api/pkg/errors"
	"ims-api/pkg/response"
	"ims-api/pkg/util"
)

// ==================== JWT ====================

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var jwtSecret []byte

// SetJWTSecret 设置密钥
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// GenerateToken 生成JWT
func GenerateToken(userID uint, username, role string, expireHours int) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
}

// ParseToken 解析JWT
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, bizErr.ErrTokenInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, bizErr.ErrTokenInvalid
	}
	return claims, nil
}

// JWT 鉴权中间件
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			response.Fail(c, bizErr.ErrUnauthorized)
			c.Abort()
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Fail(c, bizErr.ErrUnauthorized)
			c.Abort()
			return
		}
		claims, err := ParseToken(parts[1])
		if err != nil {
			response.Fail(c, bizErr.ErrTokenInvalid)
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// RequireRole 角色检查中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		roleStr, _ := role.(string)
		for _, r := range roles {
			if roleStr == r {
				c.Next()
				return
			}
		}
		response.Fail(c, bizErr.ErrForbidden)
		c.Abort()
	}
}

// CurrentUserID 获取当前用户ID
func CurrentUserID(c *gin.Context) uint {
	id, _ := c.Get("user_id")
	uid, _ := id.(uint)
	return uid
}

// ==================== Logger ====================

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		c.Next()
		zap.L().Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
			zap.String("trace_id", c.GetString("trace_id")),
		)
	}
}

// ==================== Recovery ====================

// Recovery panic恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				zap.L().Error("panic recovered", zap.Any("err", err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    10000,
					"message": "服务器内部错误",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// ==================== TraceID ====================

// TraceID 注入请求追踪ID
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := util.GenerateTraceID()
		c.Set("trace_id", id)
		c.Header("X-Trace-Id", id)
		c.Next()
	}
}

// ==================== CORS ====================

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Requested-With")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
