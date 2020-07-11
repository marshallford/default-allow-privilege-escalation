package health

import (
	"github.com/gofiber/fiber"
	"github.com/spf13/viper"
)

type Health struct {
	Ready bool `json:"ready"`
}

func Handler(config *viper.Viper, c *fiber.Ctx) {
	c.JSON(Health{Ready: true})
}
