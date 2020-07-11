package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

func New() *viper.Viper {
	c := map[string]interface{}{
		"configPath": ".",
		"server": map[string]interface{}{
			"port": 8443,
			"tls": map[string]interface{}{
				"enabled":  false,
				"dir":      "/run/secrets/tls",
				"certFile": "tls.crt",
				"keyFile":  "tls.key",
			},
		},
		"app": map[string]interface{}{
			"default": false,
		},
	}
	v := viper.New()
	for key, value := range c {
		v.SetDefault(key, value)
	}
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	v.AddConfigPath(v.GetString("configPath"))
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalln("config file was found but another error was produced")
		}
	}
	return v
}
