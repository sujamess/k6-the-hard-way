package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "time/tzdata"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sujamess/k6-the-hard-way/pkgs/httpmiddleware"
	"golang.org/x/exp/slog"
)

var (
	ErrNoDBInContext      = errors.New("mysql: no db in context")
	ErrNoDBAndTxInContext = errors.New("mysql: no db and tx in context")
)

type mysql struct {
	host            string
	port            string
	user            string
	password        string
	db              string
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
	maxOpenConns    int
	maxIdleConns    int
}

func New(options ...func(*mysql)) *sql.DB {
	m := &mysql{
		maxOpenConns:    150,
		maxIdleConns:    100,
		connMaxLifetime: 1 * time.Minute,
	}
	for _, o := range options {
		o(m)
	}

	slog.Info("mysql: connecting to MySQL", slog.String("host", m.host), slog.String("port", m.port))

	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", m.user, m.password, m.host, m.port, m.db)
	db, err := sql.Open("mysql", dataSource+"?charset=utf8")
	if err != nil {
		panic(err)
	}
	if err := db.Ping(); err != nil {
		panic(err)
	}

	db.SetConnMaxIdleTime(m.connMaxLifetime)
	db.SetMaxOpenConns(m.maxOpenConns)
	db.SetMaxIdleConns(m.maxIdleConns)
	db.SetConnMaxIdleTime(m.connMaxIdleTime)

	slog.Info("mysql: connected to MySQL")
	return db
}

func WithHost(host string) func(*mysql) {
	return func(m *mysql) {
		m.host = host
	}
}

func WithPort(port string) func(*mysql) {
	return func(m *mysql) {
		m.port = port
	}
}

func WithUser(user string) func(*mysql) {
	return func(m *mysql) {
		m.user = user
	}
}

func WithPassword(password string) func(*mysql) {
	return func(m *mysql) {
		m.password = password
	}
}

func WithDB(db string) func(*mysql) {
	return func(m *mysql) {
		m.db = db
	}
}

func Conn(ctx context.Context) (*sql.Conn, error) {
	if db, ok := ctx.Value(httpmiddleware.SQLCtxKey{}).(*sql.DB); ok {
		return db.Conn(ctx)
	}
	return nil, ErrNoDBInContext
}

func ConnOrTx(ctx context.Context) (*sql.Conn, *sql.Tx, error) {
	if tx, ok := ctx.Value(httpmiddleware.SQLTxCtxKey{}).(*sql.Tx); ok {
		return nil, tx, nil
	} else if db, ok := ctx.Value(httpmiddleware.SQLCtxKey{}).(*sql.DB); ok {
		conn, err := db.Conn(ctx)
		if err != nil {
			return nil, nil, err
		}
		return conn, nil, nil
	}
	return nil, nil, ErrNoDBAndTxInContext
}
