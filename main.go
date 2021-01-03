package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/efbar/minimal-service/handlers"
	"github.com/efbar/minimal-service/helpers"
	"github.com/efbar/minimal-service/logging"
)

func main() {

	l := log.New(os.Stdout,
		"Logger: ",
		log.Ldate|log.Ltime)
	logger := &logging.Logger{
		Logger: l,
	}
	envs := helpers.ListEnvs

	port := envs["SERVICE_PORT"]
	if port == "" {
		port = "9090"
	}

	// anyReq := handlers.HandlerAnyHTTP(l, envs)
	anyReq := handlers.HandlerAnyHTTP(*logger, envs)
	bounceReq := handlers.HandlerBounceHTTP(*logger, envs)
	healthReq := handlers.HandlerHealth(*logger)

	sm := http.NewServeMux()

	sm.Handle("/", anyReq)
	sm.Handle("/bounce", bounceReq)
	sm.Handle("/health", healthReq)

	// create a new server
	s := http.Server{
		Addr:         ":" + port,
		Handler:      sm,
		ErrorLog:     l,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// run it
	go func() {
		logger.Info("Starting server on port " + port)

		err := s.ListenAndServe()
		if err != nil {
			logger.Error("Error from server,", err.Error())
			os.Exit(1)
		}
	}()

	// get sigterm or interupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	sig := <-c
	logger.Info("Got signal:", sig.String())

	// wait max 3 seconds before shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.Shutdown(ctx)
}
