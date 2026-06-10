package handler

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"template_sch/internal/config"
)

type HealthHandler struct {
	cfg       config.Config
	startedAt time.Time
}

func NewHealthHandler(cfg config.Config) HealthHandler {
	return HealthHandler{
		cfg:       cfg,
		startedAt: time.Now(),
	}
}

func (h HealthHandler) Check(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":      "ok",
		"app":         h.cfg.App.Name,
		"environment": h.cfg.App.Environment,
		"uptime":      time.Since(h.startedAt).String(),
	})
}
