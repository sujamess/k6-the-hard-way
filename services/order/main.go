package main

import (
	"net/http"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/gorilla/mux"
	"github.com/sujamess/k6-the-hard-way/pkgs/broker"
	"github.com/sujamess/k6-the-hard-way/pkgs/db/mysql"
	"github.com/sujamess/k6-the-hard-way/pkgs/httpmiddleware"
	"github.com/sujamess/k6-the-hard-way/pkgs/httprequester"
	"github.com/sujamess/k6-the-hard-way/pkgs/httpserver"
	"golang.org/x/exp/slog"
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
	Service struct {
		Cart struct {
			Host string `env:"CART_HOST"`
		}
		Product struct {
			Host string `env:"PRODUCT_HOST"`
		}
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

	producer := broker.NewProducer(cfg.KafkaBroker.Host)
	defer func() {
		if err := producer.Close(); err != nil {
			slog.Error("broker: failed to close a producer", err)
		}
	}()

	cartService := NewCartService(httprequester.New(httprequester.WithBaseURL(cfg.Service.Cart.Host)), producer)
	productService := NewProductService(httprequester.New(httprequester.WithBaseURL(cfg.Service.Product.Host)))
	consumer := broker.NewConsumer(cfg.KafkaBroker.Host, broker.CreateOrderTopic+".order.consumer.group", NewConsumer(mysql, cartService, productService).CreateOrder)
	defer func() {
		if err := consumer.Close(); err != nil {
			slog.Error("broker: failed to close a consumer", err)
		}
	}()

	go func() {
		consumer.Consume(strings.Join([]string{broker.CreateOrderTopic}, ","))
	}()

	h := NewHandler(cartService, productService)

	srv := httpserver.New(
		httpserver.WithPort(cfg.Port),
		httpserver.WithMiddleware([]mux.MiddlewareFunc{
			httpmiddleware.SQL(mysql),
			httpmiddleware.SQLTx(mysql),
			httpmiddleware.Logging,
		}),
	)

	r := srv.Router().PathPrefix("/orders").Subrouter()
	r.Methods(http.MethodPost).Path("").HandlerFunc(h.CreateOrder)

	srv.ListenAndServe()
	srv.GracefulShutdown()
}
