package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
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

	// create a logger object
	l := log.New(os.Stdout,
		"Logger: ",
		log.Ldate|log.Ltime)
	logger := &logging.Logger{
		Logger: l,
	}

	// fill envs
	envs := helpers.ListEnvs

	// some debug prints
	for key, val := range envs {
		logger.Debug(envs["DEBUG"], key+"="+val)
	}

	// set service port
	port := envs["SERVICE_PORT"]

	// create http requests handlers
	anyReq := handlers.HandlerAnyHTTP(*logger, envs)
	bounceReq := handlers.HandlerBounceHTTP(*logger, envs)
	healthReq := handlers.HandlerHealth(*logger, envs)

	// create server mux
	sm := http.NewServeMux()

	// assign handler to paths
	sm.Handle("/", anyReq)
	sm.Handle("/bounce", bounceReq)
	sm.Handle("/health", healthReq)

	// fill the new server config
	s := http.Server{
		Addr:         ":" + port,
		Handler:      sm,
		ErrorLog:     l,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// if consul connect enabled, connect to it
	var client *consul.Client
	if envs["CONNECT"] == "1" {
		client = connectToConsul(&s, envs, logger)
	}

	// run the http server
	go func() {
		logger.Info("Starting server on port " + port)
		var err error
		if envs["HTTPS"] == "true" {
			err = s.ListenAndServeTLS("certs/minimalservice.crt", "certs/minimalservice.key")
		} else {
			err = s.ListenAndServe()
		}
		if err != nil {
			logger.Error("Error from server,", err.Error())
			os.Exit(1)
		}
	}()

	// get terminal and syscall signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGKILL)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)
	signal.Notify(c, os.Kill)

	sig := <-c
	logger.Info("Got signal:", sig.String())

	// if we are here and consul connect is active, deregister the service from it
	if envs["CONNECT"] == "1" {
		if err := client.Agent().ServiceDeregister("minimal-service"); err != nil {
			logger.Error(err.Error())
		}
		logger.Debug("Consul service deregistration")
	}

	// gracefully shutdown if connections are active
	// wait max 3 seconds before shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s.Shutdown(ctx)
}

func connectToConsul(s *http.Server, envs map[string]string, logger *logging.Logger) *consul.Client {

	// fill some vars if we are in kube
	kubeNode := os.Getenv("HOST_IP")
	kubePod := os.Getenv("POD_NAME")
	var consul_server string
	if len(kubeNode) != 0 {
		envs["CONSUL_AGENT"] = kubeNode + ":8500"
		consul_server = kubePod
	} else if envs["CONSUL_AGENT"] != "" {
		consul_server = envs["CONSUL_AGENT"]
	} else {
		consul_server, _ = helpers.GetHostname()
	}

	// create client and service
	client, _ := consul.NewClient(&consul.Config{
		Address: consul_server,
		Scheme:  "http",
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
		},
	})
	svc, _ := connect.NewService("minimal-service", client)
	defer svc.Close()

	// set service details: serviceID, port, tags, address, meta tags
	serviceID := "minimal-service"
	port, _ := strconv.Atoi(envs["SERVICE_PORT"])
	tags := []string{"microservice", "http"}

	addresses, err := net.InterfaceAddrs()
	if err != nil {
		logger.Error(err.Error())
	}
	var ipAddr string
	for _, address := range addresses {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipAddr = ipnet.IP.String()
			}
		}
	}

	meta := map[string]string{}
	if len(kubeNode) != 0 {
		meta = map[string]string{
			"pod-name": os.Getenv("POD_NAME"),
			"version":  "v1",
		}
	} else {
		meta = map[string]string{
			"hostname": consul_server,
			"version":  "v1",
		}
	}

	var tlsSkip bool
	var endpoint string
	if envs["HTTPS"] == "true" {
		endpoint = "https://" + ipAddr + ":" + envs["SERVICE_PORT"] + "/health"
		tlsSkip = false
	} else {
		endpoint = "http://" + ipAddr + ":" + envs["SERVICE_PORT"] + "/health"
		tlsSkip = true
	}

	// fill service registration, set native connect, set service check
	service := &consul.AgentServiceRegistration{
		ID:      "_" + serviceID,
		Name:    serviceID,
		Port:    port,
		Address: ipAddr,
		Tags:    tags,
		Meta:    meta,
		Connect: &consul.AgentServiceConnect{
			Native: true,
		},
		Check: &consul.AgentServiceCheck{
			HTTP:                           endpoint,
			Interval:                       "5s",
			Timeout:                        "1s",
			DeregisterCriticalServiceAfter: "30s",
			TLSSkipVerify:                  tlsSkip,
		},
	}

	if err := client.Agent().ServiceRegister(service); err != nil {
		logger.Error(err.Error())
	}

	return client
}
