package testutils

import (
	"net/http"
	"net/http/httptest"
)

type HttpHandlerFunc func(w http.ResponseWriter, r *http.Request)

func HttpServerMock(handlerFuncs map[string]HttpHandlerFunc) *httptest.Server {
	handler := http.NewServeMux()
	for endpoint, f := range handlerFuncs {
		handler.HandleFunc(endpoint, f)
	}
	srv := httptest.NewServer(handler)

	return srv
}
