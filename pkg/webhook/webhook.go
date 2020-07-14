package webhook

import (
	"defaultallowpe/pkg/health"
	"defaultallowpe/pkg/mutate"

	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
	"github.com/spf13/viper"
)

// New creates a webhook fiber app
func New(config *viper.Viper) *fiber.App {
	app := fiber.New(&fiber.Settings{
		StrictRouting: true,
	})
	api := app.Group("/api", cors.New())
	v1 := api.Group("/v1")

	health.Routes(v1, config)
	mutate.Routes(v1, config)

	// API 404 handler
	api.Use(func(c *fiber.Ctx) {
		err := c.Status(fiber.StatusNotFound).JSON(map[string]interface{}{
			"code":   fiber.StatusNotFound,
			"status": fiber.ErrNotFound.Message,
		})
		if err != nil {
			c.Next(err)
		}
	})

	// App 404 handler
	app.Use(func(c *fiber.Ctx) {
		c.Status(fiber.StatusNotFound).Send(fiber.ErrNotFound.Message)
	})

	return app
}
