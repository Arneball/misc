package arnehttp

import (
	"compress/gzip"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

type gzWriter struct {
	http.ResponseWriter
	wrapped io.Writer
}

func (w gzWriter) Write(b []byte) (int, error) {
	return w.wrapped.Write(b)
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
