package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	response "hospital/pkg"
)

// Context keys de luu thong tin user sau khi xac thuc.
// Handler dung c.GetUint64("user_id") va c.GetString("role") de lay.
const (
	CtxKeyUserID = "user_id"
	CtxKeyRole   = "role"
)

// Auth middleware xac thuc JWT token cho cac route can dang nhap.
//
// Flow:
//  1. Doc header Authorization: Bearer <token>
//  2. Parse va verify token
//  3. Lay user_id va role tu claims
//  4. Set vao gin.Context de handler phia sau su dung
//
// Neu token khong hop le, tra loi loi va dung request (Abort).
//
// Su dung:
//
//	private := router.Group("/api")
//	private.Use(middleware.Auth())
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lay Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ErrNotAuthenticated(c)
			c.Abort()
			return
		}

		// Kiem tra format "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.ErrTokenInvalid(c)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse va verify token
		claims, err := response.ParseToken(tokenString)
		if err != nil {
			switch err {
			case response.ErrTokenExpiredAt:
				response.ErrTokenExpired(c)
			case response.ErrTokenInvalidSign:
				response.ErrTokenInvalid(c)
			default:
				response.ErrTokenInvalid(c)
			}
			c.Abort()
			return
		}

		// Set thong tin user vao context cho handler su dung
		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyRole, claims.Role)

		c.Next()
	}
}

// RequireStaff middleware kiem tra user co phai nhan vien khong.
// Chay SAU middleware Auth(), nen luon co role trong context.
// Cho phep role: staff, coordinator, admin.
//
// Su dung:
//
//	staffRoutes := router.Group("/api/staff")
//	staffRoutes.Use(middleware.Auth(), middleware.RequireStaff())
func RequireStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(CtxKeyRole)
		switch role {
		case "admin", "coordinator", "staff":
			c.Next()
		default:
			response.ErrPermissionDenied(c)
			c.Abort()
		}
	}
}

// RequireAdmin middleware kiem tra user co phai admin khong.
// Chay SAU middleware Auth().
// Chi cho phep role: admin.
//
// Su dung:
//
//	adminRoutes := router.Group("/api/admin")
//	adminRoutes.Use(middleware.Auth(), middleware.RequireAdmin())
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString(CtxKeyRole)
		if role != "admin" {
			response.ErrAdminRequired(c)
			c.Abort()
			return
		}
		c.Next()
	}
}

// GetUserID lay user_id tu context. Helper de handler viet ngan gon.
// Phai dung sau middleware Auth().
func GetUserID(c *gin.Context) uint64 {
	val, exists := c.Get(CtxKeyUserID)
	if !exists {
		return 0
	}
	userID, ok := val.(uint64)
	if !ok {
		return 0
	}
	return userID
}

// GetRole lay role tu context. Helper de handler viet ngan gon.
// Phai dung sau middleware Auth().
func GetRole(c *gin.Context) string {
	return c.GetString(CtxKeyRole)
}
