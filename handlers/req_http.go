package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/efbar/minimal-service/utils"
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

// ToJSON ...
func (j *JSONResponse) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(j)
}

// HandlerHTTP ...
func HandlerHTTP(l *log.Logger) *Data {
	return &Data{l}
}

// ServeHTTP ...
func (h *Data) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	h.l.Printf("%s on %s from %s", r.Method, r.URL, r.RemoteAddr)
	st := time.Now()

	if r.Method == http.MethodGet {
		h.simpleServe(rw, r, &st)
	} else if r.Method == http.MethodPost {
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

		js, err := h.shapingJSON(r, st)

		js.ToJSON(rw)

		if err != nil {
			http.Error(rw, "Bad Request", http.StatusBadRequest)
			return
		}
	}

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

// reboundServe ...
func (h *Data) reboundServe(rw http.ResponseWriter, r *http.Request, st *time.Time) error {

	rw.Header().Set("Content-Type", "application/json")

	body, _ := ioutil.ReadAll(r.Body)

	decodedJSON := JSONPost{}

	err := json.Unmarshal(body, &decodedJSON)
	if err != nil {
		http.Error(rw, "Bad Request", http.StatusBadRequest)
		return err
	}

	if decodedJSON.Rebound == "true" {

		err := h.rawConnect(decodedJSON.Endpoint)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadGateway)
			return err
		}

		resp, err := http.Get(decodedJSON.Endpoint)
		if err != nil {
			http.Error(rw, "Bad Gateway", http.StatusBadGateway)
			return err
		}
		defer r.Body.Close()
		h.l.Println(resp.Status, decodedJSON.Endpoint)

		js, _ := h.shapingJSON(r, st)
		js.ToJSON(rw)

	} else {
		http.Error(rw, "Bad Request", http.StatusBadRequest)
		return err
	}
	return err

}

// shapingJSON ...
func (h *Data) shapingJSON(r *http.Request, st *time.Time) (*JSONResponse, error) {

	body, _ := ioutil.ReadAll(r.Body)

	host, err := os.Hostname()
	if err != nil {
		h.l.Printf("Server hostname unknown: %s\n\n", err.Error())
	}

	ft := time.Now()
	delta := ft.UnixNano() - st.UnixNano()

	var serverTiming = map[string]string{
		"Request-time":  st.UTC().String(),
		"Response-time": ft.UTC().String(),
		"Duration":      fmt.Sprint(float64(delta) / float64(time.Millisecond)),
	}
	headers := utils.GetHeaders(r, serverTiming)
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

	return js, nil

}

// shapingPlain ...
func (h *Data) shapingPlain(rw http.ResponseWriter, r *http.Request, st *time.Time) error {

	host, err := os.Hostname()
	if err != nil {
		h.l.Printf("Server hostname unknown: %s\n\n", err.Error())
	}
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
	headers := utils.GetHeaders(r, serverTiming)
	for key, value := range headers {
		fmt.Fprintf(rw, "%s: %s\n", key, value)
	}

	fmt.Fprintln(rw, "")
	io.Copy(rw, r.Body)

	return err
}
