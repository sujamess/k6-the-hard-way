package httpmiddleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lrw, r)
		addr := r.RemoteAddr
		if i := strings.LastIndex(addr, ":"); i != -1 {
			addr = addr[:i]
		}
		slog.Info(
			fmt.Sprintf("%s - - [%s]", addr, time.Now().Format("02/Jan/2006:15:04:05 -0700")),
			slog.String("method", r.Method),
			slog.String("url", r.URL.Path),
			slog.String("protocol", r.Proto),
			slog.Int("statusCode", lrw.statusCode),
			slog.Duration("duration", time.Since(start)),
			slog.String("referer", r.Referer()),
			slog.String("userAgent", r.UserAgent()),
		)
	})
}
