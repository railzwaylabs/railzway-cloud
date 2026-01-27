package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/fx"

	railzwayoss "github.com/railzwaylabs/railzway-cloud/internal/adapter/billing/railzway_oss"
	nomadAdapter "github.com/railzwaylabs/railzway-cloud/internal/adapter/provisioning/nomad"
	postgresProvisioner "github.com/railzwaylabs/railzway-cloud/internal/adapter/provisioning/postgres"
	"github.com/railzwaylabs/railzway-cloud/internal/adapter/repository/postgres"
	"github.com/railzwaylabs/railzway-cloud/internal/api"
	"github.com/railzwaylabs/railzway-cloud/internal/auth"
	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/billing"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/provisioning"
	"github.com/railzwaylabs/railzway-cloud/internal/onboarding"
	"github.com/railzwaylabs/railzway-cloud/internal/organization"
	"github.com/railzwaylabs/railzway-cloud/internal/outbox"
	"github.com/railzwaylabs/railzway-cloud/internal/reconciler"
	"github.com/railzwaylabs/railzway-cloud/internal/usecase/deployment"
	"github.com/railzwaylabs/railzway-cloud/internal/user"
	"github.com/railzwaylabs/railzway-cloud/internal/version"
	"github.com/railzwaylabs/railzway-cloud/pkg/authclient"
	"github.com/railzwaylabs/railzway-cloud/pkg/db"
	zaplog "github.com/railzwaylabs/railzway-cloud/pkg/log"
	"github.com/railzwaylabs/railzway-cloud/pkg/nomad"
	"github.com/railzwaylabs/railzway-cloud/pkg/railzwayclient"
	"github.com/railzwaylabs/railzway-cloud/pkg/snowflake"
	"go.uber.org/zap"
)

func main() {
	app := fx.New(
		fx.Provide(
			// Config
			config.Load,

			// Infrastructure (Adapters)
			nomad.NewClient,
			railzwayclient.NewFromEnv,
			authclient.NewFromEnv,

			// Domain Adapters (Bind Interfaces)
			fx.Annotate(
				postgres.NewRepository,
				fx.As(new(instance.Repository)),
			),
			fx.Annotate(
				nomadAdapter.NewAdapter,
				fx.As(new(provisioning.Provisioner)),
			),
			fx.Annotate(
				newPostgresProvisioner,
				fx.As(new(provisioning.DatabaseProvisioner)),
			),
			fx.Annotate(
				railzwayoss.NewAdapter,
				fx.As(new(billing.Engine)),
				fx.As(new(billing.PriceResolver)),
			),

			// Database Config for tenant provisioning
			newDBConfig,
			newRuntimeConfig,

			// Use Cases
			deployment.NewDeployUseCase,
			deployment.NewLifecycleUseCase,
			deployment.NewUpgradeUseCase,
			deployment.NewRolloutUseCase,

			// Legacy / Other Services
			user.NewService,
			onboarding.NewService,
			organization.NewService,
			version.NewRegistry,
			outbox.NewProcessor,
			reconciler.NewInstanceReconciler,
			reconciler.NewLifecycleReconciler,

			// Auth & Session
			auth.NewSessionManager,

			// API
			api.NewRouter,
		),
		db.Module,        // Database Module
		snowflake.Module, // Snowflake ID Module
		zaplog.Module,    // Logger Module
		fx.Invoke(registerHooks),
	)

	app.Run()
}

func registerHooks(lc fx.Lifecycle, router *api.Router, processor *outbox.Processor, instanceReconciler *reconciler.InstanceReconciler, lifecycleReconciler *reconciler.LifecycleReconciler, client *railzwayclient.Client, logger *zap.Logger) {
	var processorCancel context.CancelFunc
	var reconcilerCancel context.CancelFunc
	var lifecycleCancel context.CancelFunc

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("Starting HTTP server", zap.String("port", "8080"))

			processorCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
			processorCancel = cancel
			go processor.Run(processorCtx)

			reconcilerCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
			reconcilerCancel = cancel
			go instanceReconciler.Run(reconcilerCtx)

			lifecycleCtx, cancel := context.WithCancel(context.WithoutCancel(ctx))
			lifecycleCancel = cancel
			go lifecycleReconciler.Run(lifecycleCtx)

			// Start server in goroutine
			go func() {
				if err := router.Run(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("Server failed to start", zap.Error(err))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down HTTP server gracefully...")

			if processorCancel != nil {
				processorCancel()
			}
			if reconcilerCancel != nil {
				reconcilerCancel()
			}
			if lifecycleCancel != nil {
				lifecycleCancel()
			}

			// Create shutdown context with timeout
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			// Gracefully shutdown the server
			if err := router.Shutdown(shutdownCtx); err != nil {
				logger.Error("Server forced to shutdown", zap.Error(err))
				return err
			}

			logger.Info("HTTP server stopped gracefully")
			return nil
		},
	})
}

// newDBConfig creates database configuration for tenant provisioning
func newDBConfig(cfg *config.Config) provisioning.DBConfig {
	return provisioning.DBConfig{
		Host: cfg.ProvisionDBHost,
		Port: mustParseInt(cfg.ProvisionDBPort),
	}
}

func newRuntimeConfig(cfg *config.Config) deployment.RuntimeConfig {
	return deployment.RuntimeConfig{
		RateLimitRedisAddr:     cfg.ProvisionRateLimitRedisAddr,
		RateLimitRedisPassword: cfg.ProvisionRateLimitRedisPassword,
		RateLimitRedisDB:       cfg.ProvisionRateLimitRedisDB,
	}
}

// newPostgresProvisioner creates PostgreSQL database provisioner
func newPostgresProvisioner(cfg *config.Config) *postgresProvisioner.Adapter {
	// Build admin connection string from config
	adminConnString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.ProvisionDBUser,
		cfg.ProvisionDBPassword,
		cfg.ProvisionDBHost,
		cfg.ProvisionDBPort,
		cfg.ProvisionDBName,
		cfg.ProvisionDBSSLMode,
	)
	return postgresProvisioner.NewAdapter(adminConnString)
}

func mustParseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Sprintf("invalid port: %s", s))
	}
	return val
}
