package helpers

import (
	"fmt"
	"os"
	"strings"
)

// ListEnvs ...
var ListEnvs map[string]string

func init() {
	ListEnvs = ReadEnv()
}

// GetHostname ...
func GetHostname() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		fmt.Printf("Server hostname unknown: %s\n\n", err.Error())
	}
	return host, err
}

// ReadEnv ...
func ReadEnv() map[string]string {
	valuableEnv := []string{
		"SERVICE_PORT",
		"DELAY_MAX",
		"TRACING",
		"JAEGER_URL",
	}
	pair := map[string]string{}
	for _, elem := range os.Environ() {
		keyval := strings.SplitN(elem, "=", 2)
		if contains(valuableEnv, keyval[0]) {
			pair[keyval[0]] = keyval[1]
		}
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
