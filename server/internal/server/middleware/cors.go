package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func PublicCORS() gin.HandlerFunc {
	var origins = make([]string, 0, 30)
	origins = append(origins, []string{
		"https://theca.oxytocingroup.com",
	}...)

	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET,POST,PATCH,PUT,DELETE,OPTIONS"},
		AllowHeaders:     []string{"Accept", "Referer", "Origin", "DNT", "User-Agent", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           1 * time.Hour,
		AllowWildcard:    true,
	})
}
