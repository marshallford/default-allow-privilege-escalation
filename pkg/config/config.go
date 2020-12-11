package config

import (
	"strings"

	"github.com/spf13/viper"
)

// New creates a webhook config
func New() (*viper.Viper, error) {
	c := map[string]interface{}{
		"configPath": ".",
		"logging": map[string]interface{}{
			"level": "info",
		},
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
			return nil, err
		}
		return v, nil
	}
	v.WatchConfig()
	return v, nil
}
