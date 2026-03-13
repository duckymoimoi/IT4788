package response

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims la payload luu trong JWT token.
// Moi token chua user_id va role de middleware xac thuc
// ma khong can query database moi request.
type JWTClaims struct {
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"` // "patient", "admin", "coordinator", "staff"
	jwt.RegisteredClaims
}

// Cac loi khi parse token
var (
	ErrTokenInvalidSign = errors.New("token signature invalid")
	ErrTokenExpiredAt   = errors.New("token expired")
	ErrTokenMalformed   = errors.New("token malformed")
)

// getJWTSecret lay secret key tu bien moi truong.
// Neu khong set, dung gia tri mac dinh cho moi truong dev.
// KHONG BAO GIO dung gia tri mac dinh trong production.
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "hospital-dev-secret-key-2024"
	}
	return []byte(secret)
}

// getTokenExpiry lay thoi gian het han token.
// Mac dinh 7 ngay, du dai cho mobile app.
func getTokenExpiry() time.Duration {
	return 7 * 24 * time.Hour
}

// GenerateToken tao JWT token cho user sau khi dang nhap thanh cong.
// Token chua user_id va role, het han sau 7 ngay.
//
// role: voi patient thi role = "patient",
//
//	voi staff thi role = staff.Role (admin/coordinator/staff)
func GenerateToken(userID uint64, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(getTokenExpiry())),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ParseToken giai ma va xac thuc JWT token.
// Tra ve claims neu token hop le, tra loi cu the neu khong.
func ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Dam bao dung HS256, chong tan cong doi algorithm
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalidSign
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpiredAt
		}
		return nil, ErrTokenMalformed
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalidSign
	}

	return claims, nil
}
