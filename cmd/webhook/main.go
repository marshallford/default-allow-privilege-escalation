package main

import (
	"crypto/tls"
	"defaultallowpe/pkg/config"
	"defaultallowpe/pkg/webhook"
	"log"
	"path/filepath"
)

func main() {
	config := config.New()
	app := webhook.New(config)

	var tlsConfigs []*tls.Config
	if config.GetBool("server.tls.enabled") {
		cer, err := tls.LoadX509KeyPair(
			filepath.Join(config.GetString("server.tls.dir"), config.GetString("server.tls.certFile")),
			filepath.Join(config.GetString("server.tls.dir"), config.GetString("server.tls.keyFile")),
		)
		if err != nil {
			log.Fatalln(err)
		}
		tlsConfigs = append(tlsConfigs, &tls.Config{Certificates: []tls.Certificate{cer}})
	}
	app.Listen(config.GetInt("server.port"), tlsConfigs...)
}
