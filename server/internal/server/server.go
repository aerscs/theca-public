package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/OxytocinGroup/theca-v3/internal/config"
	"github.com/OxytocinGroup/theca-v3/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg           *config.Config
	log           *slog.Logger
	publicServer  *http.Server
	publicRouter  *gin.Engine
	swaggerServer *http.Server
	swaggerRouter *gin.Engine
}

func New(cfg *config.Config, log *slog.Logger) *Server {
	gin.SetMode(gin.ReleaseMode)
	if cfg.IsLocalRun {
		gin.SetMode(gin.DebugMode)
	}

	publicRouter := gin.New()
	publicRouter.Use(gin.Recovery())
	publicRouter.Use(middleware.PublicCORS())
	publicServer := &http.Server{
		Addr:    cfg.PublicAddr,
		Handler: publicRouter,
	}

	swaggerRouter := gin.New()
	swaggerRouter.Use(gin.Recovery())
	swaggerRouter.Use(middleware.PublicCORS())
	swaggerServer := &http.Server{
		Addr:    cfg.SwaggerAddr,
		Handler: swaggerRouter,
	}

	return &Server{
		cfg:           cfg,
		log:           log,
		publicServer:  publicServer,
		publicRouter:  publicRouter,
		swaggerServer: swaggerServer,
		swaggerRouter: swaggerRouter,
	}
}

func (s *Server) Start() {
	go s.startServer("publicServer", s.publicServer)
	go s.startServer("swaggerServer", s.swaggerServer)
}

func (s *Server) Stop() {
	const op = "server.stop"
	log := s.log.With("op", op)
	log.Info("Stopping public server")

	shutdownTimeout := 5 * time.Second
	if s.cfg.ShutdownTimeout > 0 {
		shutdownTimeout = time.Duration(s.cfg.ShutdownTimeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	wg := sync.WaitGroup{}

	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := s.publicServer.Shutdown(ctx); err != nil {
			log.Error("Failed to shutdown public server", "error", err, "timeout", shutdownTimeout)
		}
	}()

	go func() {
		defer wg.Done()
		if err := s.swaggerServer.Shutdown(ctx); err != nil {
			log.Error("Failed to shutdown swagger server", "error", err, "timeout", shutdownTimeout)
		}
	}()

	wg.Wait()

	log.Info("servers stopped successfully")
}

func (s *Server) startServer(logName string, server *http.Server) {
	log := s.log.With(slog.String("op", logName+".Start"), slog.String("host", server.Addr))
	log.Info("starting server", slog.String("addr", server.Addr))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("failed to start server", "error", err)
	}
}

func (s *Server) Router() *gin.Engine {
	return s.publicRouter
}

func (s *Server) SwaggerRouter() *gin.Engine {
	return s.swaggerRouter
}
