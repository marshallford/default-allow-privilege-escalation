package main

import (
	"crypto/tls"
	"defaultallowpe/pkg/config"
	"defaultallowpe/pkg/webhook"
	"net"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"

	"github.com/cloudflare/certinel"
	"github.com/cloudflare/certinel/fswatcher"
)

func main() {
	log.SetOutput(os.Stdout)
	config, err := config.New()
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("config file found but another error was produced")
	}
	logLevel, err := log.ParseLevel(config.GetString("logging.level"))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("invalid log level")
	}
	log.SetLevel(logLevel)
	if config.GetBool("logging.json") {
		log.SetFormatter(&log.JSONFormatter{})
	}
	config.OnConfigChange(func(e fsnotify.Event) {
		log.Info("config file changed")
	})

	app := webhook.New(config)
	ln, err := net.Listen("tcp", ":"+config.GetString("server.port"))
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("tcp listener failed")
	}
	if config.GetBool("server.tls.enabled") {
		watcher, err := fswatcher.New(
			filepath.Join(config.GetString("server.tls.dir"), config.GetString("server.tls.certFile")),
			filepath.Join(config.GetString("server.tls.dir"), config.GetString("server.tls.keyFile")),
		)
		if err != nil {
			log.WithFields(log.Fields{
				"err": err,
			}).Fatal("unable to read server certificate")
		}
		sentinel := certinel.New(watcher, func(err error) {
			log.WithFields(log.Fields{
				"err": err,
			}).Warn("certinel was unable to reload the certificate")
		})
		sentinel.Watch()
		ln = tls.NewListener(ln, &tls.Config{GetCertificate: sentinel.GetCertificate, MinVersion: tls.VersionTLS12})
	}
	err = app.Listener(ln)
	log.WithFields(log.Fields{
		"err": err,
	}).Fatal("webhook server failed")
}
