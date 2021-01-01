package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/efbar/minimal-service/helpers"
	"github.com/efbar/minimal-service/logging"
	tracer "github.com/efbar/minimal-service/tracer"
	"go.opentelemetry.io/otel/label"
)

// Data ...
type Data struct {
	l    logging.Logger
	envs map[string]string
}

// JSONResponse ...
type JSONResponse struct {
	Host       string            `json:"host"`
	StatusCode int               `json:"statuscode"`
	Headers    map[string]string `json:"headers"`
	Proto      string            `json:"protocol"`
	RequestURI string            `json:"requestURI"`
	ServedBy   string            `json:"servedBy"`
	Method     string            `json:"method"`
	Body       string            `json:"body,omitempty"`
}

// JSONPost ...
type JSONPost struct {
	Rebound  string `json:"rebound"`
	Endpoint string `json:"endpoint"`
}

// HandlerAnyHTTP ...
func HandlerAnyHTTP(l logging.Logger, envs map[string]string) *Data {
	return &Data{l, envs}
}

// HandlerBounceHTTP ...
func HandlerBounceHTTP(l logging.Logger, envs map[string]string) *Data {
	return &Data{l, envs}
}

// ServeHTTP ...
func (h *Data) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.l.Info(r.Method, r.URL.String(), r.RemoteAddr)
	st := time.Now()

	s, _ := strconv.Atoi(h.envs["DISCARD_QUOTA"])
	if helpers.RandBool(s, &h.l) {
		h.l.Info("Request discarded")
		return
	}

	if r.Method == http.MethodGet {
		h.simpleServe(rw, r, &st)
	} else if r.Method == http.MethodPost && r.RequestURI == "/bounce" {
		if err := h.reboundServe(rw, r, &st); err != nil {
			h.l.Info(err.Error())
		}
	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// simpleServe ...
func (h *Data) simpleServe(rw http.ResponseWriter, r *http.Request, st *time.Time) {

	contentType := r.Header.Get("Content-type")
	if contentType == "text/plain" {
		rw.Header().Set("Content-Type", "text/plain")
		if err := h.shapingPlain(rw, r, st); err != nil {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		rw.Header().Set("Content-Type", "application/json")

		delayEnv := h.envs["DELAY_MAX"]
		if len(delayEnv) != 0 && delayEnv != "0" {
			if err := h.Delayer(delayEnv); err != nil {
				h.l.Info(err.Error())
			}
		}

		enableTracing := h.envs["TRACING"]
		jaegerURL := h.envs["JAEGER_URL"]
		if enableTracing == "1" {
			h.execTracing(jaegerURL, "minimal-service")
		}

		js, err := h.shapingJSON(r, st)
		if err != nil {
			h.l.Info("error shaping json", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err = js.EncodeJSON(rw); err != nil {
			http.Error(rw, "Bad Request", http.StatusBadRequest)
			return
		}
	}

}

// reboundServe ...
func (h *Data) reboundServe(rw http.ResponseWriter, r *http.Request, st *time.Time) error {

	rw.Header().Set("Content-Type", "application/json")

	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	delayEnv := h.envs["DELAY_MAX"]
	if len(delayEnv) != 0 && delayEnv != "0" {
		if err := h.Delayer(delayEnv); err != nil {
			h.l.Info(err.Error())
			return err
		}
	}

	jsonRecived := &JSONPost{}

	if err := DecodeJSON(body, jsonRecived, rw); err != nil {
		h.l.Info("error decode json", err.Error())
		return err
	}

	if jsonRecived.Rebound == "true" {
		h.l.Info("jsonRecived.Rebound", jsonRecived.Rebound)
		if err := h.rawConnect(jsonRecived.Endpoint); err != nil {
			http.Error(rw, "Bad Request", http.StatusBadGateway)
			return err
		}

		resp, err := http.Get(jsonRecived.Endpoint)
		if err != nil {
			http.Error(rw, "Bad Gateway", http.StatusBadGateway)
			return err
		}
		h.l.Info(resp.Status, jsonRecived.Endpoint)

		js, err := h.shapingJSON(r, st)
		if err != nil {
			h.l.Info("error shaping json", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return err
		}

		if err = js.EncodeJSON(rw); err != nil {
			http.Error(rw, "Bad Request", http.StatusBadRequest)
			return err
		}

		enableTracing := h.envs["TRACING"]
		jaegerURL := h.envs["JAEGER_URL"]
		if enableTracing == "1" {
			if err := h.execTracing(jaegerURL, "minimal-service"); err != nil {
				return err
			}
		}

	}
	return nil

}

func (h *Data) rawConnect(endpoint string) error {
	timeout := time.Second

	url, err := url.ParseRequestURI(endpoint)
	if err != nil {
		h.l.Info("Wrong url,", err.Error())
		return err
	}
	h.l.Info("Correct url:", url.String())

	host := url.Hostname()
	port := url.Port()
	h.l.Info("Splitted url:", host, port)

	if host != "" {
		s, err := net.ResolveIPAddr("ip", host)
		if err != nil {
			h.l.Info("Resolve Error:", err.Error())
			return err
		}
		h.l.Info("Resolved url:", url.Scheme, s.String())

		if err == nil {
			if port == "" {
				if url.Scheme == "http" {
					port = "80"
				} else {
					port = "443"
				}
			}

			h.l.Info("Before dial:", host, port)
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
			if err != nil {
				h.l.Info("Connection Error:", err.Error())
				return err
			}
			if conn != nil {
				h.l.Info("Open:", net.JoinHostPort(host, port))
				defer conn.Close()
			}

			return err
		}
		return err
	}

	return err
}

// shapingJSON ...
func (h *Data) shapingJSON(r *http.Request, st *time.Time) (*JSONResponse, error) {

	body, _ := ioutil.ReadAll(r.Body)

	host, err := helpers.GetHostname()

	ft := time.Now()
	delta := ft.UnixNano() - st.UnixNano()

	var serverTiming = map[string]string{
		"Request-time":  st.UTC().String(),
		"Response-time": ft.UTC().String(),
		"Duration":      fmt.Sprint(float64(delta) / float64(time.Millisecond)),
	}
	headers := CollectHeaders(r, serverTiming)
	js := &JSONResponse{
		Host:       r.Host,
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       string(body),
		Proto:      string(r.Proto),
		RequestURI: string(r.RequestURI),
		ServedBy:   host,
		Method:     string(r.Method),
	}

	return js, err

}

// shapingPlain ...
func (h *Data) shapingPlain(rw http.ResponseWriter, r *http.Request, st *time.Time) error {

	host, err := helpers.GetHostname()
	fmt.Fprintf(rw, "Request served by %s\n\n", host)

	fmt.Fprintf(rw, "%s %s %s\n", r.Method, r.URL, r.Proto)
	fmt.Fprintf(rw, "Host: %s\n", r.Host)

	ft := time.Now()
	delta := ft.UnixNano() - st.UnixNano()
	var serverTiming = map[string]string{
		"RequestTime":  st.UTC().String(),
		"ResponseTime": ft.UTC().String(),
		"Duration":     fmt.Sprint(float64(delta) / float64(time.Millisecond)),
	}
	headers := CollectHeaders(r, serverTiming)
	for key, value := range headers {
		fmt.Fprintf(rw, "%s: %s\n", key, value)
	}

	fmt.Fprintln(rw, "")
	io.Copy(rw, r.Body)

	return err
}

func (h *Data) execTracing(url string, service string) error {
	t := &tracer.TraceObject{}
	tags := &[]label.KeyValue{
		label.String("exporter", "jaeger"),
		label.Float64("float", 312.23),
		label.Int64("int", 123),
	}
	err := t.Opentracer(url, service, *tags)
	return err
}
