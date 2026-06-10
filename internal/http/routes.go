package http

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"

	"template_sch/internal/config"
	"template_sch/internal/http/dependency"
	"template_sch/internal/http/handler"
)

func RegisterRoutes(app *fiber.App, cfg config.Config, deps dependency.Dependencies, logger *slog.Logger) {
	healthHandler := handler.NewHealthHandler(cfg)

	app.Get("/health", healthHandler.Check)

	_ = deps
	_ = logger
	// generator:api-route
}
