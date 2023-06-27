package arnehttp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCrasher(t *testing.T) {
	h := LoggingHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}), WithReq, WithResp)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/bla", nil))
	if rr.Code != 200 {
		t.Error("Wanted 200")
	}
}
