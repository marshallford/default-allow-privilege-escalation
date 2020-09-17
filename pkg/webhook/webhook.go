package webhook

import (
	"defaultallowpe/pkg/health"
	"defaultallowpe/pkg/mutate"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/spf13/viper"
)

// New creates a webhook fiber app
func New(config *viper.Viper) *fiber.App {
	app := fiber.New(fiber.Config{
		StrictRouting: true,
	})
	api := app.Group("/api", cors.New())
	v1 := api.Group("/v1")

	health.Routes(v1, config)
	mutate.Routes(v1, config)

	// API 404 handler
	api.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(map[string]interface{}{
			"code":   fiber.StatusNotFound,
			"status": fiber.ErrNotFound.Message,
		})
	})

	// App 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).SendString(fiber.ErrNotFound.Message)
	})

	return app
}
