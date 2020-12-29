package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var logHandleReq io.Writer

func setupReqHTTPTest(t *testing.T) *Data {
	logHandleReq := os.Stdout
	return &Data{
		log.New(logHandleReq,
			"Test Logger: ",
			log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func TestReqHTTPResp(t *testing.T) {
	os.Setenv("DELAY_MAX", "2")
	req := httptest.NewRequest("GET", "/", nil)

	rr := httptest.NewRecorder()
	handler := setupReqHTTPTest(t)

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	//

	reqURI := "/test"
	reqAltGet := httptest.NewRequest("GET", reqURI, nil)

	rrAltGet := httptest.NewRecorder()
	handlerAltGet := setupReqHTTPTest(t)

	handlerAltGet.ServeHTTP(rrAltGet, reqAltGet)

	assert.Equal(t, http.StatusOK, rrAltGet.Code)

	var tmpl JSONResponse
	jsonRes := json.NewDecoder(strings.NewReader(rrAltGet.Body.String()))
	err := jsonRes.Decode(&tmpl)
	if err != nil {
		log.Fatalln(err)
	}
	assert.Equal(t, reqURI, tmpl.RequestURI)

	//

	reqWrongPOST := httptest.NewRequest("POST", "/", nil)

	rrWrongPOST := httptest.NewRecorder()
	handlerWrongPOST := setupReqHTTPTest(t)

	handlerWrongPOST.ServeHTTP(rrWrongPOST, reqWrongPOST)

	assert.Equal(t, http.StatusMethodNotAllowed, rrWrongPOST.Code)

	//
	reqWrongAltBody := "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it:443\"}"
	reqWrongAltPOST := httptest.NewRequest("POST", "/bounce", strings.NewReader(reqWrongAltBody))

	rrWrongAltPOST := httptest.NewRecorder()
	handlerWrongAltPOST := setupReqHTTPTest(t)

	handlerWrongAltPOST.ServeHTTP(rrWrongAltPOST, reqWrongAltPOST)

	assert.Equal(t, http.StatusBadGateway, rrWrongAltPOST.Code)
	assert.Equal(t, "Bad Gateway\n", rrWrongAltPOST.Body.String())

	////

	rightBody := "{\"rebound\":\"true\",\"endpoint\":\"http://www.google.it:80\"}"
	reqRightPOST := httptest.NewRequest("POST", "/bounce", strings.NewReader(rightBody))

	rrRightPOST := httptest.NewRecorder()
	handlerRightPOST := setupReqHTTPTest(t)

	handlerRightPOST.ServeHTTP(rrRightPOST, reqRightPOST)

	assert.Equal(t, http.StatusOK, rrRightPOST.Code)
	assert.NotContains(t, rrRightPOST.Body.String(), "body")

	////

	fakeDNS := "fakedns"
	DNSErrBody := "{\"rebound\":\"true\",\"endpoint\":\"http://" + fakeDNS + "\"}"
	reqDNSErrPOST := httptest.NewRequest("POST", "/bounce", strings.NewReader(DNSErrBody))

	rrDNSErrPOST := httptest.NewRecorder()
	handlerDNSErrPOST := setupReqHTTPTest(t)

	handlerDNSErrPOST.ServeHTTP(rrDNSErrPOST, reqDNSErrPOST)

	assert.Equal(t, http.StatusBadGateway, rrDNSErrPOST.Code)
	assert.Contains(t, rrDNSErrPOST.Body.String(), "no such host")

	////

	RefusedErrBody := "{\"rebound\":\"true\",\"endpoint\":\"http://127.0.0.1:7777\"}"
	reqRefusedErrPOST := httptest.NewRequest("POST", "/bounce", strings.NewReader(RefusedErrBody))

	rrRefusedErrPOST := httptest.NewRecorder()
	handlerRefusedErrPOST := setupReqHTTPTest(t)

	handlerRefusedErrPOST.ServeHTTP(rrRefusedErrPOST, reqRefusedErrPOST)

	assert.Equal(t, http.StatusBadGateway, rrRefusedErrPOST.Code)
	assert.Contains(t, rrRefusedErrPOST.Body.String(), "connection refused")

	////

	notAllowedBody := "notAllowedBody"
	reqnotAllowedErrPOST := httptest.NewRequest("POST", "/bounce", strings.NewReader(notAllowedBody))

	rrnotAllowedErrPOST := httptest.NewRecorder()
	handlernotAllowedErrPOST := setupReqHTTPTest(t)

	handlernotAllowedErrPOST.ServeHTTP(rrnotAllowedErrPOST, reqnotAllowedErrPOST)

	assert.Equal(t, http.StatusBadRequest, rrnotAllowedErrPOST.Code)
	assert.Equal(t, "Bad Request\n", rrnotAllowedErrPOST.Body.String())

}

///// THIS is worse then testify
// if status := rrWrongPOST.Code; status != http.StatusBadRequest {
// 	t.Errorf("handler returned wrong status code: got \"%v\" want \"%v\"",
// 		status, http.StatusBadRequest)
// }

// expected := `Bad Request`
// if rr.Body.String() != expected {
// 	t.Errorf("handler returned unexpected body: got \"%v\" want \"%v\"",
// 		rr.Body.String(), expected)
// }
