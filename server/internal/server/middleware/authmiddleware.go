package middleware

import (
	"strings"
	"time"

	"github.com/aerscs/theca-public/internal/utils/errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware interface {
	JWTMiddleware() gin.HandlerFunc
}

type middleware struct {
	accessSecret  []byte
	refreshSecret []byte
}

func NewAuthMiddleware(accessSecret, refreshSecret []byte) AuthMiddleware {
	return &middleware{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
	}
}

func (mw *middleware) JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Unauthorized"))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Invalid auth header format"))
			c.Abort()
			return
		}

		tokenStr := parts[1]

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			} else if method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return mw.accessSecret, nil
		})

		if err != nil || !token.Valid {
			errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Invalid or expired token"))
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Token expired"))
					c.Abort()
					return
				}
			} else {
				errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Missing expiration claim"))
				c.Abort()
				return
			}

			if iat, ok := claims["iat"].(float64); ok {
				if time.Now().Unix() < int64(iat) {
					errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Token used before issued"))
					c.Abort()
					return
				}
			}

			if userIDFloat, ok := claims["userId"].(float64); ok {
				userID := uint(userIDFloat)
				if userID == 0 {
					errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Invalid user ID in token"))
					c.Abort()
					return
				}
				c.Set("userID", userID)
			} else {
				errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Missing user ID in token"))
				c.Abort()
				return
			}
		} else {
			errors.RespondWithError(c, errors.New(errors.CodeUnauthorized, "Invalid token claims"))
			c.Abort()
			return
		}

		c.Next()
	}
}
