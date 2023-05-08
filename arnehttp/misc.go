package arnehttp

import (
	"compress/gzip"
	"github.com/rs/zerolog/log"
	"io"
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

func GzipHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writer, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			panic(err)
		}
		defer func(writer *gzip.Writer) {
			err := writer.Close()
			if err != nil {
				log.Err(err).Msg("flushing")
			}
		}(writer)
		w2 := gzWriter{
			w, writer,
		}
		w.Header().Set("Content-Encoding", "gzip")
		h.ServeHTTP(w2, r)
	}
}

func GzipHandlerFunc(h http.HandlerFunc) http.HandlerFunc {
	return GzipHandler(h)
}

type gzWriter struct {
	http.ResponseWriter
	wrapped io.Writer
}

func (w gzWriter) Write(b []byte) (int, error) {
	return w.wrapped.Write(b)
}
