package health

import (
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

// Health model
type Health struct {
	Ready bool `json:"ready"`
}

// Routes manages Fiber routes for health pkg
func Routes(r fiber.Router, config *viper.Viper) {
	r.Get("/healthz", HandlerFunc(config))
}

// HandlerFunc returns a func that is a HTTP handler for health requests
func HandlerFunc(config *viper.Viper) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(Health{Ready: true})
	}
}
