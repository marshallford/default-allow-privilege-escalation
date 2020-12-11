package main

import (
	"crypto/tls"
	"defaultallowpe/pkg/config"
	"defaultallowpe/pkg/webhook"
	stdlog "log"
	"net"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"

	"github.com/cloudflare/certinel"
	"github.com/cloudflare/certinel/fswatcher"
)

func main() {
	logConfig := zap.NewProductionConfig()
	logConfig.Sampling = nil
	logger, err := logConfig.Build()
	if err != nil {
		stdlog.Panic("unable to construct logger")
	}

	defer func() {
		if err := logger.Sync(); err != nil {
			stdlog.Panic("unable to flush possible buffered log entries")
		}
	}()

	if err != nil {
		stdlog.Panic("error: unable to construct logger")
	}
	log := logger.Sugar()

	config, err := config.New()
	if err != nil {
		log.Fatalw("config file found but another error was produced",
			"err", err,
		)
	}
	level := zap.NewAtomicLevel()
	err = level.UnmarshalText(([]byte(config.GetString("logging.level"))))
	if err != nil {
		log.Fatalw("invalid log level",
			"err", err,
		)
	}
	logConfig.Level.SetLevel(level.Level())

	config.OnConfigChange(func(e fsnotify.Event) {
		log.Info("config file changed")
	})

	app := webhook.New(config)
	ln, err := net.Listen("tcp", ":"+config.GetString("server.port"))
	if err != nil {
		log.Fatalw("tcp listener failed",
			"err", err,
		)
	}
	if config.GetBool("server.tls.enabled") {
		watcher, err := fswatcher.New(
			filepath.Join(config.GetString("server.tls.dir"), config.GetString("server.tls.certFile")),
			filepath.Join(config.GetString("server.tls.dir"), config.GetString("server.tls.keyFile")),
		)
		if err != nil {
			log.Fatalw("unable to read server certificate",
				"err", err,
			)
		}
		sentinel := certinel.New(watcher, func(err error) {
			log.Warnw("certinel was unable to reload the certificate",
				"err", err,
			)
		})
		sentinel.Watch()
		ln = tls.NewListener(ln, &tls.Config{GetCertificate: sentinel.GetCertificate, MinVersion: tls.VersionTLS12})
	}
	err = app.Listener(ln)
	log.Fatalw("webhook server failed",
		"err", err,
	)
}
