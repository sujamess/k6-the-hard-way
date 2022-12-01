package main

import (
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/gorilla/mux"
	"github.com/sujamess/k6-the-hard-way/pkgs/db/mysql"
	"github.com/sujamess/k6-the-hard-way/pkgs/httpmiddleware"
	"github.com/sujamess/k6-the-hard-way/pkgs/httpserver"
)

type config struct {
	Port        int `env:"PORT"`
	KafkaBroker struct {
		Host string `env:"KAFKA_BROKER_HOST"`
	}
	MySQL struct {
		Host     string `env:"MYSQL_HOST"`
		Port     string `env:"MYSQL_PORT"`
		User     string `env:"MYSQL_USER"`
		Password string `env:"MYSQL_PASSWORD"`
		Database string `env:"MYSQL_DATABASE"`
	}
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}

	mysql := mysql.New(
		mysql.WithHost(cfg.MySQL.Host),
		mysql.WithPort(cfg.MySQL.Port),
		mysql.WithUser(cfg.MySQL.User),
		mysql.WithPassword(cfg.MySQL.Password),
		mysql.WithDB(cfg.MySQL.Database),
	)

	h := NewHandler()

	srv := httpserver.New(
		httpserver.WithPort(cfg.Port),
		httpserver.WithMiddleware([]mux.MiddlewareFunc{
			httpmiddleware.SQL(mysql),
			httpmiddleware.Logging,
		}),
	)

	router := srv.Router()

	productRouter := router.PathPrefix("/products").Subrouter()
	productRouter.Methods(http.MethodPost).Path("/mock").HandlerFunc(h.MockProducts)
	productRouter.Methods(http.MethodGet).Path("/with-cache").HandlerFunc(h.ListProductsByIDsWithCache)
	productRouter.Methods(http.MethodGet).Path("").HandlerFunc(h.ListProductsByIDs)

	router.Methods(http.MethodGet).Path("/healthcheck").HandlerFunc(h.Healthcheck)

	srv.ListenAndServe()
	srv.GracefulShutdown()
}
