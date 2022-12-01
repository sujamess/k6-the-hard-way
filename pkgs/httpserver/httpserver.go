package httpserver

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/exp/slog"
)

type HttpServer interface {
	ListenAndServe()
	GracefulShutdown()
	Router() *mux.Router
}

type httpServer struct {
	port        int
	r           *mux.Router
	srv         *http.Server
	middlewares []mux.MiddlewareFunc
}

func WithMiddleware(middlewares []mux.MiddlewareFunc) func(*httpServer) {
	return func(hs *httpServer) {
		hs.middlewares = middlewares
	}
}

func New(options ...func(*httpServer)) HttpServer {
	hs := &httpServer{}
	for _, o := range options {
		o(hs)
	}
	r := mux.NewRouter()
	r.Use(hs.middlewares...)
	hs.r = r
	return hs
}

func WithPort(port int) func(*httpServer) {
	return func(server *httpServer) {
		server.port = port
	}
}

func (hs *httpServer) ListenAndServe() {
	hs.srv = &http.Server{
		Addr:         "0.0.0.0:" + strconv.Itoa(hs.port),
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 30,
		IdleTimeout:  time.Second * 60,
		Handler:      hs.r,
	}

	go func() {
		slog.Info("httpserver: listening and serve", slog.Int("port", hs.port))
		if err := hs.srv.ListenAndServe(); err != nil {
			slog.Error("httpserver: failed to listen and serve", err)
		}
	}()
}

func (hs *httpServer) GracefulShutdown() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	hs.srv.Shutdown(ctx)
	slog.Warn("httpserver: gracefully shutting down")
	os.Exit(0)
	slog.Warn("httpserver: gracefully shut down")
}

func (hs *httpServer) Router() *mux.Router {
	return hs.r
}
