package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/auth"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/billing"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/config"
	applogger "github.com/emrecetinkayadev/vpn-tridot/backend/internal/logger"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/hash"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/platform/jwt"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/regions"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/server"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/server/handlers/auth"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/server/handlers/billing"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/server/handlers/nodes"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/server/handlers/regions"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/server/middleware"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/server/setup"
	"github.com/emrecetinkayadev/vpn-tridot/backend/internal/storage/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger, err := applogger.New(cfg.App.Env, cfg.Log.Level)
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	store, err := postgres.New(context.Background(), cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer store.Close()
	if err := store.Ping(context.Background()); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	hasher, err := hash.NewArgon2Hasher(cfg.Auth.ArgonMemory, cfg.Auth.ArgonIterations, cfg.Auth.ArgonSaltLength, cfg.Auth.ArgonKeyLength, cfg.Auth.ArgonParallelism)
	if err != nil {
		log.Fatalf("init hasher: %v", err)
	}

	jwtManager, err := jwt.NewManager(cfg.Auth.JWTSigningKey, cfg.App.Name, cfg.Auth.AccessTokenTTL)
	if err != nil {
		log.Fatalf("init jwt manager: %v", err)
	}

	authRepo := postgres.NewAuthRepository(store.Pool())
	authService := auth.NewService(authRepo, hasher, jwtManager, cfg.Auth)
	authHandler := authhandler.New(authService, logger)
	authMiddleware := middleware.Auth(jwtManager, logger)

	regionsRepo := postgres.NewRegionsRepository(store.Pool())
	regionsService := regions.NewService(regionsRepo, cfg)
	if err := regionsService.SeedDefaultRegions(context.Background()); err != nil {
		log.Fatalf("seed regions: %v", err)
	}
	regionHandler := regionshandler.New(regionsService, logger)
	nodeHandler := nodeshandler.New(regionsService, cfg.Node, logger)

	billingRepo := postgres.NewBillingRepository(store.Pool())
	providers := map[string]billing.PaymentProvider{}
	if stripeProvider := billing.NewStripeProvider(cfg.Billing.Stripe); stripeProvider != nil {
		providers[stripeProvider.Name()] = stripeProvider
	}
	if iyzicoProvider := billing.NewIyzicoProvider(cfg.Billing.Iyzico); iyzicoProvider != nil {
		providers[iyzicoProvider.Name()] = iyzicoProvider
	}
	billingService := billing.NewService(billingRepo, authRepo, providers, cfg.Billing)
	if err := billingService.SeedDefaultPlans(context.Background()); err != nil {
		log.Fatalf("seed plans: %v", err)
	}
	billingHandler := billinghandler.New(billingService, logger)

	deps := setup.Dependencies{
		AuthHandler:    authHandler,
		AuthMiddleware: authMiddleware,
		BillingHandler: billingHandler,
		RegionsHandler: regionHandler,
		NodesHandler:   nodeHandler,
	}

	srv := server.New(cfg, logger, deps)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := srv.Run(ctx); err != nil {
		logger.Error("server shutdown", zap.Error(err))
	}
}
