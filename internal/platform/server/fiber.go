package server

import (
	"github.com/gofiber/fiber/v2"

	"template_sch/internal/config"
)

func NewFiber(cfg config.Config) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
	})
}
