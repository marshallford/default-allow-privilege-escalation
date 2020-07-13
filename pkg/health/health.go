package health

import (
	"github.com/gofiber/fiber"
	"github.com/spf13/viper"
)

// Health model
type Health struct {
	Ready bool `json:"ready"`
}

// HandlerFunc returns a func that is a HTTP handler for health requests
func HandlerFunc(config *viper.Viper) fiber.Handler {
	return func(c *fiber.Ctx) {
		err := c.JSON(Health{Ready: true})
		if err != nil {
			c.Next(err)
		}
	}
}
