package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/emrecetinkayadev/vpn-tridot/node-agent/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	agent, exporter, err := newAgent(cfg)
	if err != nil {
		log.Fatalf("init agent: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	var metricsSrv *http.Server
	var metricsDone chan struct{}
	if exporter != nil && cfg.Agent.MetricsAddress != "" {
		metricsSrv = &http.Server{Addr: cfg.Agent.MetricsAddress, Handler: exporter.Handler()}
		metricsDone = make(chan struct{})
		go func() {
			defer close(metricsDone)
			if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("metrics server error: %v", err)
			}
		}()
		go func() {
			<-ctx.Done()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := metricsSrv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("metrics server shutdown: %v", err)
			}
		}()
	}

	if err := agent.Run(ctx); err != nil {
		log.Fatalf("agent error: %v", err)
	}
	stop()
	if metricsDone != nil {
		<-metricsDone
	}
}
