package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Scalingo/go-utils/logger"
	"github.com/curzolapierre/hook-manager/config"
	redisCtr "github.com/curzolapierre/hook-manager/redis"
	"github.com/curzolapierre/hook-manager/webserver"
	"github.com/sirupsen/logrus"
)

func logLevel() logrus.Level {
	switch os.Getenv("LOGGER_LEVEL") {
	case "panic":
		return logrus.PanicLevel
	case "fatal":
		return logrus.FatalLevel
	case "warn":
		return logrus.WarnLevel
	case "info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	default:
		return logrus.InfoLevel
	}
}

func initLogger() logrus.FieldLogger {
	logger := logrus.New()
	logger.SetLevel(logLevel())
	logger.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000",
		FullTimestamp:   true,
	}

	var fieldLogger logrus.FieldLogger = logger

	return fieldLogger
}

func main() {
	log := initLogger()
	ctx := logger.ToCtx(context.Background(), log)

	config, err := config.Lookup()
	if err != nil {
		log.WithError(err).Panic("Fail to load environment")
		return
	}

	_, err = redisCtr.Client(config)
	if err != nil {
		log.WithError(err).Panic("fail to init redis client")
		return
	}

	httpListenAddr := fmt.Sprintf("%s:%s", config.HttpHost, config.HttpPort)
	log.Infof("Starting the web server on %v", httpListenAddr)

	// Define routers
	router := webserver.NewRouter(ctx, config)

	go func() {
		err := http.ListenAndServe(httpListenAddr, router)
		if err != nil {
			log.WithError(err).Error("Fail to start web server")
			os.Exit(-1)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	for range signals {
		log.Info("Stopping the server")
		// for _, stopper := range stoppers {
		// 	stopper()
		// }
		os.Exit(0)
	}

}
