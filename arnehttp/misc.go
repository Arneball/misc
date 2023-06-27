package arnehttp

import (
	"bytes"
	"encoding/json"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"
	"time"
)

type wrappedW struct {
	http.ResponseWriter
	r                     *http.Request
	start                 time.Time
	written               int
	code                  int
	recordReq, recordResp bool
	resp                  []byte
	req                   []byte
}

func (w *wrappedW) Write(bytes []byte) (int, error) {
	n, err := w.ResponseWriter.Write(bytes)
	w.written += n
	w.resp = append(w.resp, bytes...)
	return n, err
}

func (w *wrappedW) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

type LoggingOpt interface {
	Pimp(w *wrappedW, e *zerolog.Event) *zerolog.Event
}

type LoggingOptFunc func(w *wrappedW, e *zerolog.Event) *zerolog.Event

type specialLoggingShit struct {
	req bool
	f   func(w *wrappedW, e *zerolog.Event) *zerolog.Event
}

func (s specialLoggingShit) Pimp(w *wrappedW, e *zerolog.Event) *zerolog.Event {
	return s.f(w, e)
}

func (l LoggingOptFunc) Pimp(w *wrappedW, e *zerolog.Event) *zerolog.Event {
	return l(w, e)
}

func WithDuration() LoggingOptFunc {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		return e.Dur("duration", time.Now().Sub(w.start))
	}
}

func WithPath() LoggingOptFunc {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		return e.Str("path", w.r.URL.Path)
	}
}

func WithCode() LoggingOptFunc {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		return e.Int("code", w.code)
	}
}

func WithLength() LoggingOptFunc {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		return e.Int("len", w.written)
	}
}

func IgnorePath(path string) LoggingOptFunc {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		if strings.HasPrefix(w.r.URL.Path, path) {
			return e.Discard()
		}
		return e
	}
}

func WithParams() LoggingOptFunc {
	return func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		query := w.r.URL.Query()
		d := zerolog.Dict()
		for k, v := range query {
			d = d.Strs(k, v)
		}
		return e.Dict("query", d)
	}
}

var WithResp LoggingOpt = specialLoggingShit{
	req: false,
	f: func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		if w.recordResp {
			return e.EmbedObject(maybeJson{key: "resp", payload: w.resp})
		}
		return e
	},
}

var WithReq LoggingOpt = specialLoggingShit{
	req: true,
	f: func(w *wrappedW, e *zerolog.Event) *zerolog.Event {
		if w.recordReq {
			return e.EmbedObject(maybeJson{key: "req", payload: w.req})
		}
		return e
	},
}

func LoggingHandler(h http.Handler, opts ...LoggingOpt) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var recordResp, recordReq bool
		for _, opt := range opts {
			switch casted := opt.(type) {
			case specialLoggingShit:
				recordReq = recordReq || casted.req
				recordResp = recordResp || !casted.req
			}
		}

		wr := wrappedW{
			ResponseWriter: w,
			code:           200,
			start:          time.Now(),
			r:              r,
			recordResp:     recordResp,
			recordReq:      recordReq,
		}
		if recordReq {
			b, err := io.ReadAll(r.Body)
			if err != nil {
				log.Err(err).Msg("io.ReadAll")
				panic(err)
			}
			if err := r.Body.Close(); err != nil {
				log.Err(err).Msg("r.Close")
				panic(err)
			}
			wr.req = b
			r.Body = io.NopCloser(bytes.NewReader(b))
		}
		h.ServeHTTP(&wr, r)
		evt := log.Debug()
		for _, opt := range opts {
			evt = opt.Pimp(&wr, evt)
		}
		evt.Msg("access")
	}
}

type maybeJson struct {
	key     string
	payload []byte
}

func (m maybeJson) MarshalZerologObject(e *zerolog.Event) {
	var b bytes.Buffer
	var respJson []byte
	if err := json.Compact(&b, m.payload); err == nil {
		respJson = b.Bytes()
	} else {
		respJson = []byte(`"<not json>"`)
	}
	e.RawJSON(m.key, respJson)
}

type headerLogger struct {
	key    string
	header http.Header
}

func (h headerLogger) MarshalZerologObject(e *zerolog.Event) {
	sub := zerolog.Dict()
	for k, values := range h.header {
		sub.Strs(k, values)
	}
	e.Dict(h.key, sub)
}
