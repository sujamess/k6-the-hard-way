package httpwriter

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"sync"

	"golang.org/x/exp/slog"
)

var gzipWriter = sync.Pool{
	New: func() any {
		return gzip.NewWriter(nil)
	},
}

type (
	NilResponse   struct{}
	ErrorResponse struct {
		ErrorMsg string `json:"errorMessage"`
	}
)

type httpWriter struct {
	headers  map[string]string
	withGzip bool
}

func WithHeaders(headers map[string]string) func(*httpWriter) {
	return func(writer *httpWriter) {
		writer.headers = headers
	}
}

func WithGzip(b bool) func(*httpWriter) {
	return func(writer *httpWriter) {
		writer.withGzip = b
	}
}

func Write(w http.ResponseWriter, statusCode int, data interface{}, options ...func(*httpWriter)) error {
	writer := &httpWriter{withGzip: true}
	for _, o := range options {
		o(writer)
	}

	// // add custom headers
	for k, v := range writer.headers {
		w.Header().Add(k, v)
	}

	var gw *gzip.Writer
	if writer.withGzip {
		gw = gzipWriter.Get().(*gzip.Writer)
		gw.Reset(w)
		defer gzipWriter.Put(gw)
		defer func() {
			_ = gw.Close()
		}()
		w.Header().Add("Content-Encoding", "gzip")
	}
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	var res any
	switch statusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		if data == nil {
			res = NilResponse{}
		} else {
			res = data
		}
	default:
		slog.Error("message", data.(error))
		res = ErrorResponse{ErrorMsg: data.(error).Error()}
	}

	if writer.withGzip {
		return json.NewEncoder(gw).Encode(res)
	}
	return json.NewEncoder(w).Encode(res)
}
