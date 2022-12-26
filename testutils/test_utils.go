package testutils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	contentTypeHeader   = "Content-Type"
	contentTypeJSON     = "application/json"
	authorizationHeader = "Authorization"
)

type HTTPRequest struct {
	Method   string
	Endpoint string
	Body     string
	Token    string
}

func (r *HTTPRequest) MakeHTTPRequest(respBody any) (statusCode int, err error) {
	request, err := http.NewRequest(r.Method, r.Endpoint, strings.NewReader(r.Body))
	if err != nil {
		return
	}

	request.Header.Set(contentTypeHeader, contentTypeJSON)
	if r.Token != "" {
		request.Header.Set(authorizationHeader, r.Token)
	}

	resp, err := (&http.Client{}).Do(request)
	if err != nil {
		return
	}

	statusCode = resp.StatusCode
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return statusCode, nil
	}

	if err = json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return
	}

	if err = resp.Body.Close(); err != nil {
		return
	}

	return statusCode, nil
}

func (r *HTTPRequest) MakeTestHTTPRequest(HandlerFunc gin.HandlerFunc, respBody any, params ...gin.Param) (statusCode int, err error) {
	request, err := http.NewRequest(r.Method, r.Endpoint, strings.NewReader(r.Body))
	if err != nil {
		return
	}

	resp := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(resp)
	c.Request = request
	c.Params = params

	HandlerFunc(c)

	statusCode = resp.Code
	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		return statusCode, nil
	}

	if err = json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return
	}

	return statusCode, nil
}
