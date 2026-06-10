package http

import (
	"github.com/gofiber/fiber/v2"

	"template_sch/internal/config"
	"template_sch/internal/http/handler"
)

func RegisterRoutes(app *fiber.App, cfg config.Config) {
	healthHandler := handler.NewHealthHandler(cfg)

	app.Get("/health", healthHandler.Check)
}
