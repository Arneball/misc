package arnehttp

import (
	"compress/gzip"
	"io"
	"net/http"
)

func GzipHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w2 := gzWriter{
			w, gzip.NewWriter(w),
		}
		h.ServeHTTP(w2, r)
	}
}

type gzWriter struct {
	http.ResponseWriter
	wrapped io.Writer
}

func (w gzWriter) Write(b []byte) (int, error) {
	return w.wrapped.Write(b)
}
