package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/efbar/minimal-service/handlers"
)

func main() {

	l := log.New(os.Stdout, "Logger: ", log.LstdFlags)

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "9090"
	}

	reqHTTP := handlers.HandlerHTTP(l)
	reqHealth := handlers.HandlerHealth(l)

	sm := http.NewServeMux()

	sm.Handle("/", reqHTTP)
	sm.Handle("/health", reqHealth)

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
		l.Println("Starting server on port " + port)

		err := s.ListenAndServe()
		if err != nil {
			l.Printf("Error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	// get sigterm or interupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	sig := <-c
	log.Println("Got signal:", sig)

	// wait max 3 seconds before shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.Shutdown(ctx)
}
