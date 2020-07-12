package health

import (
	"github.com/gofiber/fiber"
	"github.com/spf13/viper"
)

// Health model
type Health struct {
	Ready bool `json:"ready"`
}

// Handler is the HTTP handler for health requests
func Handler(config *viper.Viper, c *fiber.Ctx) {
	err := c.JSON(Health{Ready: true})
	if err != nil {
		c.Next(err)
	}
}
