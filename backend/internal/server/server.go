package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/metrics"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/server/middleware"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/server/setup"
)

// Server wraps the HTTP server and lifecycle management for the API.
type Server struct {
	cfg    config.Config
	logger *zap.Logger
	engine *gin.Engine
	http   *http.Server
}

// New creates a configured HTTP server instance.
func New(cfg config.Config, logger *zap.Logger, deps setup.Dependencies, collector *metrics.Collector) *Server {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.RequestLogger(logger, middleware.RequestLoggerOptions{
		Enabled:         cfg.Log.Request.Enabled,
		Headers:         cfg.Log.Request.Headers,
		MaskHeaders:     cfg.Log.Request.MaskHeaders,
		QueryParams:     cfg.Log.Request.QueryParams,
		MaskQueryParams: cfg.Log.Request.MaskQueryParams,
	}))
	engine.Use(middleware.Recovery(logger))
	if cfg.Observability.Metrics.Enabled {
		engine.Use(middleware.Metrics(collector))
	}

	setup.Register(engine, cfg, deps, logger)
	if cfg.Observability.Metrics.Enabled && collector != nil {
		engine.GET(cfg.Observability.Metrics.Path, gin.WrapH(collector.Handler()))
	}

	httpSrv := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           engine,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.WriteTimeout,
	}

	return &Server{
		cfg:    cfg,
		logger: logger,
		engine: engine,
		http:   httpSrv,
	}
}

// Run starts the HTTP server and blocks until the context is cancelled.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.cfg.HTTP.ShutdownTimeout)
		defer cancel()
		return s.http.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// Engine exposes the underlying gin engine. Useful for tests.
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
