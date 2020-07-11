package webhook

import (
	"defaultallowpe/pkg/health"
	"defaultallowpe/pkg/mutate"

	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
	"github.com/spf13/viper"
)

func New(config *viper.Viper) *fiber.App {
	app := fiber.New(&fiber.Settings{
		StrictRouting: true,
	})
	api := app.Group("/api", cors.New())
	v1 := api.Group("/v1")

	v1.Get("/healthz", func(c *fiber.Ctx) {
		health.Handler(config, c)
	})

	v1.Post("/mutate", func(c *fiber.Ctx) {
		mutate.Handler(config, c)
	})

	// API 404 handler
	api.Use(func(c *fiber.Ctx) {
		c.Status(fiber.StatusNotFound).JSON(map[string]interface{}{
			"code":   fiber.StatusNotFound,
			"status": fiber.ErrNotFound.Message,
		})
	})

	// App 404 handler
	app.Use(func(c *fiber.Ctx) {
		c.Status(fiber.StatusNotFound).Send(fiber.ErrNotFound.Message)
	})

	return app
}
