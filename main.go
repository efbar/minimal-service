package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/efbar/minimal-service/handlers"
	"github.com/efbar/minimal-service/helpers"
	"github.com/efbar/minimal-service/logging"
	consul "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/connect"
)

func main() {
	var client *consul.Client
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

	anyReq := handlers.HandlerAnyHTTP(*logger, envs)
	bounceReq := handlers.HandlerBounceHTTP(*logger, envs)
	healthReq := handlers.HandlerHealth(*logger, envs)

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

	if len(envs["CONNECT"]) == 0 {
		envs["CONNECT"] = "false"
	}
	connectActive, err := strconv.ParseBool(envs["CONNECT"])
	if err != nil {
		logger.Error("Error from consul,", err.Error())
	}

	if connectActive {
		client = connectToConsul(&s, envs, logger)
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
	signal.Notify(c, syscall.SIGKILL)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	signal.Notify(c, os.Kill)

	sig := <-c
	logger.Info("Got signal:", sig.String())

	if err := client.Agent().ServiceDeregister("minimal-service"); err != nil {
		logger.Error(err.Error())
	}
	if connectActive {
		logger.Debug("Consul service deregistration")
	}

	// wait max 3 seconds before shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.Shutdown(ctx)
}

func connectToConsul(s *http.Server, envs map[string]string, logger *logging.Logger) *consul.Client {

	if len(envs["CONSUL_SERVER"]) == 0 {
		envs["CONSUL_SERVER"] = "127.0.0.1:8500"
	}
	conf := &consul.Config{
		Address:    envs["CONSUL_SERVER"],
		Transport:  &http.Transport{},
		HttpClient: &http.Client{},
		HttpAuth:   &consul.HttpBasicAuth{},
		WaitTime:   0,
		TLSConfig:  consul.TLSConfig{},
	}
	client, _ := consul.NewClient(conf)
	svc, _ := connect.NewService("minimal-service", client)
	defer svc.Close()
	s.TLSConfig = svc.ServerTLSConfig()

	serviceID := "minimal-service"
	if len(envs["SERVICE_PORT"]) == 0 {
		envs["SERVICE_PORT"] = "9090"
	}
	port, _ := strconv.Atoi(envs["SERVICE_PORT"])
	tags := []string{"microservice", "http"}
	service := &consul.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceID,
		Port:    port,
		Address: serviceID,
		Tags:    tags,
		Check: &consul.AgentServiceCheck{
			HTTP:     "http://" + serviceID + ":" + envs["SERVICE_PORT"] + "/health",
			Interval: "5s",
			Timeout:  "1s",
		},
	}

	if err := client.Agent().ServiceRegister(service); err != nil {
		logger.Error(err.Error())
	}

	return client
}
