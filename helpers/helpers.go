package helpers

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/efbar/minimal-service/logging"
)

// ListEnvs ...
var ListEnvs map[string]string

// EnableDebug ...
var EnableDebug bool

func init() {
	ListEnvs = ReadEnv()
	rand.Seed(time.Now().UnixNano())
}

// GetHostname ... simply get hostname
func GetHostname() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		fmt.Printf("Server hostname unknown: %s\n\n", err.Error())
	}

	return host, err
}

// ReadEnv ... collect important envs and set some defaults if needed
func ReadEnv() map[string]string {
	valuableEnv := []string{
		"HTTPS",
		"SERVICE_PORT",
		"DELAY_MAX",
		"TRACING",
		"JAEGER_URL",
		"DISCARD_QUOTA",
		"REJECT",
		"DEBUG",
		"CONNECT",
		"CONSUL_AGENT",
		"CONSUL_HTTP_TOKEN",
		"CONSUL_CACERT",
		"HOST_IP",
		"POD_IP",
		"POD_NAME",
		"POD_NAMESPACE",
	}

	pair := map[string]string{}
	for _, elem := range os.Environ() {
		keyval := strings.SplitN(elem, "=", 2)
		if contains(valuableEnv, keyval[0]) {
			pair[keyval[0]] = keyval[1]
		}
	}

	// set some defaults if env not present
	if len(pair["HTTPS"]) == 0 {
		pair["HTTPS"] = "false"
	}
	if len(pair["SERVICE_PORT"]) == 0 {
		pair["SERVICE_PORT"] = "9090"
	}
	if len(pair["DELAY_MAX"]) == 0 {
		pair["DELAY_MAX"] = "0"
	}
	if len(pair["TRACING"]) == 0 {
		pair["TRACING"] = "0"
	}
	if len(pair["JAEGER_URL"]) == 0 {
		pair["JAEGER_URL"] = "http://localhost:14268/api/traces"
	}
	if len(pair["DISCARD_QUOTA"]) == 0 {
		pair["DISCARD_QUOTA"] = "0"
	}
	if len(pair["REJECT"]) == 0 {
		pair["REJECT"] = "0"
	}
	if len(pair["DEBUG"]) == 0 {
		pair["DEBUG"] = "0"
	}
	if len(pair["CONSUL_AGENT"]) == 0 {
		pair["CONSUL_AGENT"] = "http://127.0.0.1:8500"
	}
	if len(pair["CONNECT"]) == 0 {
		pair["CONNECT"] = "0"
	}

	return pair
}

func contains(listS []string, s string) bool {
	for _, value := range listS {
		if value == s {
			return true
		}
	}

	return false
}

// RandBool ... random true/false generator based on quota percentage
func RandBool(i int, l *logging.Logger) bool {
	if i > 100 || i < 0 {
		i = 0
	}
	quota := float32(i) / float32(100)

	return rand.Float32() < quota
}
