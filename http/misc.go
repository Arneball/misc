package http

import (
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type wrappedW struct {
	http.ResponseWriter
	written int
	code    int
}

func (w *wrappedW) Write(bytes []byte) (int, error) {
	n, err := w.ResponseWriter.Write(bytes)
	w.written += n
	return n, err
}

func (w *wrappedW) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func LoggingHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		before := time.Now()
		wr := wrappedW{ResponseWriter: w, code: 200}
		h.ServeHTTP(&wr, r)
		log.Debug().
			Str("path", r.URL.Path).
			Dur("duration", time.Now().Sub(before)).
			Int("code", wr.code).
			Int("len", wr.written).
			Msg("access")
	}
}
