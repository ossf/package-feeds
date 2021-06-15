package testutils

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

type HTTPHandlerFunc func(w http.ResponseWriter, r *http.Request)

func HTTPServerMock(handlerFuncs map[string]HTTPHandlerFunc) *httptest.Server {
	handler := http.NewServeMux()
	for endpoint, f := range handlerFuncs {
		handler.HandleFunc(endpoint, f)
	}
	srv := httptest.NewServer(handler)

	return srv
}

func UnexpectedWriteError(err error) string {
	return fmt.Sprintf("Unexpected error during mock http server write: %s", err.Error())
}

func NotFoundHandlerFunc(w http.ResponseWriter, r *http.Request) {
	http.NotFound(w, r)
}
