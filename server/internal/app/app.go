package app

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/aerscs/theca-public/internal/config"
	"github.com/aerscs/theca-public/internal/model"
	"github.com/aerscs/theca-public/internal/repository"
	"github.com/aerscs/theca-public/internal/server"
	"github.com/aerscs/theca-public/internal/server/handlers"
	"github.com/aerscs/theca-public/internal/server/middleware"
	"github.com/aerscs/theca-public/internal/service"
	"github.com/aerscs/theca-public/internal/storage/database"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/aerscs/theca-public/docs"
)

type Application struct {
	cfg            *config.Config
	log            *slog.Logger
	server         *server.Server
	authMiddleware middleware.AuthMiddleware
	db             database.Database
}

func New(ctx context.Context, cfg *config.Config, log *slog.Logger) *Application {
	server := server.New(cfg, log)

	if cfg.IsLocalRun {
		server.Router().Use(gin.Logger())
	}

	db, err := database.ConnectDatabase(ctx, cfg)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Bookmark{}); err != nil {
		log.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}
	if err := db.CreateIndexes(); err != nil {
		log.Error("failed to create indexes", "error", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}

	cache := repository.NewRedisRepository(redisClient, log)

	repo := repository.NewRepository(db.GetDB(), log)

	service := service.NewService(repo, cache, log, cfg)

	handlers := handlers.NewHandler(service, log)

	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTAccessSecret, cfg.JWTRefreshSecret)

	initHandlers(server, handlers, authMiddleware)
	initSwaggerHandlers(server)

	app := &Application{
		cfg:            cfg,
		log:            log,
		server:         server,
		authMiddleware: authMiddleware,
		db:             db,
	}

	return app
}

func initHandlers(server *server.Server, handlers *handlers.Handler, authMiddleware middleware.AuthMiddleware) {
	v1 := server.Router().Group("/v1")
	v1.GET("/health", handlers.HealthCheck)
	v1.POST("/register", handlers.Register)
	v1.POST("/login", handlers.Login)
	v1.POST("/send-email-verification-code", handlers.SendEmailVerificationCode)
	v1.PATCH("/verify-email", handlers.VerifyEmail)
	v1.GET("/refresh-tokens", handlers.RefreshTokens)
	v1.POST("/request-password-reset", handlers.RequestPasswordReset)
	v1.PATCH("/reset-password", handlers.ResetPassword)

	secV1 := v1.Group("/api", authMiddleware.JWTMiddleware())
	secV1.DELETE("/logout", handlers.Logout)
	secV1.GET("/user/me", handlers.GetSelfUser)
	secV1.GET("/user/:id", handlers.GetUser)

	bookmarks := secV1.Group("/bookmarks")
	bookmarks.POST("", handlers.AddBookmark)
	bookmarks.GET("", handlers.GetBookmarks)
	bookmarks.GET("/:id", handlers.GetBookmarkByID)
	bookmarks.PATCH("/:id", handlers.UpdateBookmark)
	bookmarks.DELETE("/:id", handlers.DeleteBookmark)
	bookmarks.PUT("/import", handlers.ImportBookmarks)
	bookmarks.GET("/export", handlers.ExportBookmarks)

	v2 := server.Router().Group("/v2")
	secV2 := v2.Group("/api", authMiddleware.JWTMiddleware())
	bookmarksV2 := secV2.Group("/bookmarks")
	bookmarksV2.POST("/import", handlers.ImportBookmarksV2)
	bookmarksV2.GET("/export", handlers.ExportBookmarksV2)
}

func initSwaggerHandlers(server *server.Server) {
	server.SwaggerRouter().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
}

func (a *Application) Run() {
	const op = "app.Run"
	a.server.Start()
	log := a.log.With(slog.String("op", op))
	log.Info("application started",
		slog.String("timestamp", time.Now().Format(time.RFC3339)),
		slog.String("name", a.cfg.AppName),
	)
}

func (a *Application) Stop() {
	const op = "app.Stop"
	log := a.log.With(slog.String("op", op))
	log.Info("shutting down application...")

	a.server.Stop()

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			log.Error("error closing database connection", "error", err)
		} else {
			log.Info("database connection closed")
		}
	}

	a.log.Info("application stopped")
}
