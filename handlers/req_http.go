package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Data ...
type Data struct {
	l *log.Logger
}

// JSONResponses ...
type JSONResponses []*JSONResponse

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
func HandlerAnyHTTP(l *log.Logger) *Data {
	return &Data{l}
}

// HandlerBounceHTTP ...
func HandlerBounceHTTP(l *log.Logger) *Data {
	return &Data{l}
}

// ServeHTTP ...
func (h *Data) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.l.Printf("%s on %s from %s", r.Method, r.URL, r.RemoteAddr)
	st := time.Now()

	if r.Method == http.MethodGet {
		h.simpleServe(rw, r, &st)
	} else if r.Method == http.MethodPost && r.RequestURI == "/bounce" {
		err := h.reboundServe(rw, r, &st)
		if err != nil {
			h.l.Println(err)
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
		err := h.shapingPlain(rw, r, st)
		if err != nil {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		rw.Header().Set("Content-Type", "application/json")

		err := h.Delayer()
		if err != nil {
			h.l.Println(err)
		}

		js, err := h.shapingJSON(r, st)

		err = js.EncodeJSON(rw)
		if err != nil {
			http.Error(rw, "Bad Request", http.StatusBadRequest)
			return
		}
	}

}

// reboundServe ...
func (h *Data) reboundServe(rw http.ResponseWriter, r *http.Request, st *time.Time) error {

	rw.Header().Set("Content-Type", "application/json")

	body, _ := ioutil.ReadAll(r.Body)

	err := h.Delayer()
	if err != nil {
		h.l.Println(err)
	}

	jsonRecived := &JSONPost{}

	err = DecodeJSON(body, jsonRecived, rw)

	if jsonRecived.Rebound == "true" {

		err := h.rawConnect(jsonRecived.Endpoint)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadGateway)
			return err
		}

		resp, err := http.Get(jsonRecived.Endpoint)
		if err != nil {
			http.Error(rw, "Bad Gateway", http.StatusBadGateway)
			return err
		}
		defer r.Body.Close()
		h.l.Println(resp.Status, jsonRecived.Endpoint)

		js, err := h.shapingJSON(r, st)

		js.EncodeJSON(rw)

	}

	return err

}

func (h *Data) rawConnect(endpoint string) error {
	timeout := time.Second

	url, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return err
	}
	h.l.Println("Parsed:", url)

	host := url.Hostname()
	port := url.Port()
	h.l.Println("Splitted:", host, port)

	if host != "" {
		s, err := net.ResolveIPAddr("ip", host)
		if err != nil {
			fmt.Println("Resolve Error:", err)
			return err
		}
		h.l.Println("Resolved:", url.Scheme, s)

		if err == nil {
			if port == "" {
				if url.Scheme == "http" {
					port = "80"
				} else {
					port = "443"
				}
			}

			h.l.Println("Before dial:", host, port)
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
			if err != nil {
				h.l.Println("Connection Error:", err)
				return err
			}
			if conn != nil {
				h.l.Println("Open:", net.JoinHostPort(host, port))
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

	host, err := h.GetHostname()

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

	host, err := h.GetHostname()
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
