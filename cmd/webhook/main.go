package main

import (
	"crypto/tls"
	"defaultallowpe/pkg/config"
	"defaultallowpe/pkg/webhook"
	"log"
	"path/filepath"

	"github.com/cloudflare/certinel"
	"github.com/cloudflare/certinel/fswatcher"
)

func main() {
	config := config.New()
	app := webhook.New(config)

	var tlsConfigs []*tls.Config
	if config.GetBool("server.tls.enabled") {
		watcher, err := fswatcher.New(
			filepath.Join(config.GetString("server.tls.dir"), config.GetString("server.tls.certFile")),
			filepath.Join(config.GetString("server.tls.dir"), config.GetString("server.tls.keyFile")),
		)
		if err != nil {
			log.Fatalf("unable to read server certificate. err='%s'", err)
		}
		sentinel := certinel.New(watcher, func(err error) {
			log.Printf("certinel was unable to reload the certificate. err='%s'", err)
		})
		sentinel.Watch()
		tlsConfigs = append(tlsConfigs, &tls.Config{GetCertificate: sentinel.GetCertificate})
	}
	err := app.Listen(config.GetInt("server.port"), tlsConfigs...)
	if err != nil {
		log.Fatalln(err)
	}
}
