package arnehttp

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
	"time"
)

type wrappedW struct {
	http.ResponseWriter
	r       *http.Request
	start   time.Time
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

type LoggingOpt func(w *wrappedW, e *zerolog.Event) *zerolog.Event

func WithDuration() LoggingOpt {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		return e.Dur("duration", time.Now().Sub(w.start))
	}
}

func WithPath() LoggingOpt {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		return e.Str("path", w.r.URL.Path)
	}
}

func WithCode() LoggingOpt {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		return e.Int("code", w.code)
	}
}

func WithLength() LoggingOpt {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		return e.Int("len", w.written)
	}
}

func IgnorePath(path string) LoggingOpt {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		if strings.HasPrefix(w.r.URL.Path, path) {
			return e.Discard()
		}
		return e
	}
}

func LoggingHandler(h http.Handler, opts ...LoggingOpt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wr := wrappedW{ResponseWriter: w, code: 200, start: time.Now(), r: r}
		h.ServeHTTP(&wr, r)
		evt := log.Debug()
		for _, opt := range opts {
			evt = opt(&wr, evt)
		}
		evt.Msg("access")
	}
}
