package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

	discarded, _ := strconv.Atoi(h.envs["DISCARD_QUOTA"])
	rejected, _ := strconv.Atoi(h.envs["REJECT"])
	if helpers.RandBool(discarded, &h.l) {
		h.l.Info("Request discarded")
		if rejected == 1 {
			if r.Header.Get("Content-type") == "application/json" {
				ErrorJSON(rw, "Internal Server Error", http.StatusInternalServerError)
			} else {
				http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			}
			h.l.Debug(h.envs["DEBUG"], "Status code 500 sent")
			respHeaders := make(map[string]string)
			respHeaders["Content-type"] = r.Header.Get("Content-type")
			respHeaders["User-Agent"] = r.Header.Get("User-Agent")
			respHeaders["FailCause"] = "request rejected"
			h.execTracing("minimal-service", http.StatusInternalServerError, http.StatusText(500), respHeaders)
		}
		return
	}

	if r.Method == http.MethodGet {
		h.simpleServe(rw, r, &st)
	} else if r.Method == http.MethodPost && r.RequestURI == "/bounce" {
		if err := h.reboundServe(rw, r, &st); err != nil {
			h.l.Error(err.Error())
		}
	} else {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		respHeaders := make(map[string]string)
		respHeaders["URI"] = r.RequestURI
		h.execTracing("minimal-service", http.StatusMethodNotAllowed, http.StatusText(405), respHeaders)
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
				h.l.Error(err.Error())
			}
		}

		js, err := h.shapingJSON(r, st)
		if err != nil {
			h.l.Error("error shaping json", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err = js.EncodeJSON(rw); err != nil {
			http.Error(rw, "Bad Request", http.StatusBadRequest)
			return
		}

		h.execTracing("minimal-service", http.StatusOK, http.StatusText(200), js.Headers)

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
			h.l.Error(err.Error())
			return err
		}
	}

	jsonRecived := &JSONPost{}

	if err := DecodeJSON(body, jsonRecived, rw); err != nil {
		h.l.Error("error decode json", err.Error())
		return err
	}

	if jsonRecived.Rebound == "true" {
		h.l.Debug(h.envs["DEBUG"], "jsonRecived.Rebound", jsonRecived.Rebound)
		if err := h.rawConnect(jsonRecived.Endpoint); err != nil {
			http.Error(rw, "Bad Gateway", http.StatusBadGateway)
			respHeaders := make(map[string]string)
			respHeaders["Content-type"] = r.Header.Get("Content-type")
			respHeaders["User-Agent"] = r.Header.Get("User-Agent")
			respHeaders["FailCause"] = "Bad Gateway"
			h.execTracing("minimal-service", http.StatusBadGateway, http.StatusText(502), respHeaders)
			return err
		}

		resp, err := http.Get(jsonRecived.Endpoint)
		if err != nil {
			http.Error(rw, "Bad Gateway", http.StatusBadGateway)
			return err
		}
		defer resp.Body.Close()
		h.l.Info(resp.Status, jsonRecived.Endpoint)

		r.Header = resp.Header
		r.Body = ioutil.NopCloser(strings.NewReader(resp.Status))

		js, err := h.shapingJSON(r, st)
		if err != nil {
			h.l.Error("error shaping json", err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return err
		}

		if err = js.EncodeJSON(rw); err != nil {
			http.Error(rw, "Bad Request", http.StatusBadRequest)
			return err
		}

		h.execTracing("minimal-service", http.StatusOK, http.StatusText(200), js.Headers)

	}
	return nil

}

func (h *Data) rawConnect(endpoint string) error {
	timeout := time.Second

	url, err := url.ParseRequestURI(endpoint)
	if err != nil {
		h.l.Error("Wrong url")
		return err
	}
	h.l.Debug(h.envs["DEBUG"], "Correct url:", url.String())

	host := url.Hostname()
	port := url.Port()
	h.l.Debug(h.envs["DEBUG"], "Splitted url:", host, port)

	s, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		h.l.Error("Resolve Error:", err.Error())
		return err
	}
	h.l.Debug(h.envs["DEBUG"], "Resolved url:", url.Scheme, s.String())

	if port == "" {
		if url.Scheme == "http" {
			port = "80"
		} else {
			port = "443"
		}
	}

	h.l.Debug(h.envs["DEBUG"], "Before dial:", host, port)
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		h.l.Error("Connection Error:", err.Error())
		return err
	}
	if conn != nil {
		h.l.Debug(h.envs["DEBUG"], "Open:", net.JoinHostPort(host, port))
		defer conn.Close()
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

	h.execTracing("minimal-service", http.StatusOK, http.StatusText(200), headers)

	return err
}

func (h *Data) execTracing(service string, code int, message string, headers map[string]string) {
	enableTracing := h.envs["TRACING"]
	if enableTracing == "1" {
		jaegerURL := h.envs["JAEGER_URL"]
		host, _ := helpers.GetHostname()

		t := &tracer.TraceObject{}
		tags := []label.KeyValue{
			label.String("Exporter", "opentracing-jaeger-plugin"),
			label.String("Hostname", host),
		}

		for key, value := range headers {
			tags = append(tags, label.String(key, value))
		}
		err := t.Opentracer(jaegerURL, service, code, message, tags, &h.l, h.envs)

		if err != nil {
			h.l.Error("TRACING, ", err.Error())
		}
	}
}
