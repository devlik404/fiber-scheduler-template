package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"template_sch/internal/config"
	httpapp "template_sch/internal/http"
	httpdependency "template_sch/internal/http/dependency"
	"template_sch/internal/platform/database"
	applogger "template_sch/internal/platform/logger"
	"template_sch/internal/platform/server"
	"template_sch/internal/scheduler"
	"template_sch/internal/task"
)

func main() {
	bootstrapLogger := applogger.Bootstrap()

	cfg, err := config.Load()
	if err != nil {
		bootstrapLogger.Error("config load failed", "error", err)
		os.Exit(1)
	}

	logger := applogger.New(cfg)
	logger.Info("application booting", "addr", cfg.HTTPAddress(), "scheduler_enabled", cfg.Scheduler.Enabled)

	db, err := database.OpenConnections(context.Background(), cfg.Database, logger)
	if err != nil {
		logger.Error("database connection failed", "error", err)
		os.Exit(1)
	}
	if db != nil {
		defer func() {
			if err := database.Close(db); err != nil {
				logger.Error("database close failed", "error", err)
			}
		}()
	}

	app := server.NewFiber(cfg)
	httpapp.RegisterRoutes(app, cfg, httpdependency.Dependencies{DB: db}, logger)

	taskRegistry := task.NewRegistry(cfg.Task, task.Dependencies{DB: db}, logger)
	taskScheduler, err := scheduler.New(cfg.Scheduler, taskRegistry.Jobs(), logger)
	if err != nil {
		logger.Error("scheduler create failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if cfg.Scheduler.Enabled {
		taskScheduler.Start()
		defer taskScheduler.Stop()
	}

	go func() {
		if err := app.Listen(cfg.HTTPAddress()); err != nil {
			logger.Error("http server stopped unexpectedly", "error", err)
			stop()
		}
	}()

	logger.Info("application started", "addr", cfg.HTTPAddress())
	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Error("http server shutdown failed", "error", err)
	}

	logger.Info("application stopped")
}
