package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Daniel-Seifert/go-echo/internal/util"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

//nolint:gochecknoglobals,gocritic
var (
	app = kingpin.New("go-echo", "Simple go echo service")

	// Logging
	logFormat = app.Flag("log-format", "Log-Format for Kelon. Must be one of [TEXT, JSON]").Default("TEXT").Envar("LOG_FORMAT").Enum("TEXT", "JSON")
	port      = app.Flag("port", "Specify port go-echo listens on.").Default("8888").Envar("PORT").Uint16()
	address   = app.Flag("address", "Specify address go-echo listens on.").Default("0.0.0.0").Envar("ADDRESS").IP()
)

func main() {
	// Setup arg parser
	app.HelpFlag.Short('h')
	kingpin.MustParse(app.Parse(os.Args[1:]))
	log.SetOutput(os.Stdout)

	// Set log format
	switch *logFormat {
	case "JSON":
		log.SetFormatter(util.UTCFormatter{Formatter: &log.JSONFormatter{}})
	default:
		log.SetFormatter(util.UTCFormatter{Formatter: &log.TextFormatter{FullTimestamp: true}})
	}

	// Configure server
	router := mux.NewRouter()
	router.PathPrefix("/").Methods("POST").HandlerFunc(echo)
	server := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("%s:%d", address.String(), *port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start Server
	go func() {
		log.Infof("Starting go-echo v0.1.0 server at: http://%s:%d", address.String(), *port)
		if err := server.ListenAndServe(); err != nil {
			log.Warn(err)
		}
	}()

	// Await termination
	blockTillInterrupt(os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	gracefullShutdown(server, 10*time.Second)
	log.Info("Good bye!")
}

func gracefullShutdown(server *http.Server, timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	onShutdown := make(chan struct{})
	defer cancel()

	server.RegisterOnShutdown(func() {
		onShutdown <- struct{}{}
	})
	server.SetKeepAlivesEnabled(false)
	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Error while shutting down server")
	}

	select {
	case <-onShutdown:
		log.Info("Server shutdown completed")
	case <-ctx.Done():
		log.Warn("Server failed to shutdown before timeout!")
	}
}

func blockTillInterrupt(signals ...os.Signal) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, signals...)

	// Block until we receive our signal.
	<-interruptChan
	log.Info("Caught Sigterm")
}

func echo(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	var buffer bytes.Buffer
	_ = request.Write(&buffer)
	_, _ = writer.Write(buffer.Bytes())
}
