package httpmiddleware

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/sujamess/k6-the-hard-way/pkgs/httpwriter"
)

type SQLCtxKey struct{}

func SQL(db *sql.DB) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), SQLCtxKey{}, db)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

type SQLTxCtxKey struct{}

func SQLTx(db *sql.DB) func(h http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tx, err := db.BeginTx(r.Context(), nil)
			if err != nil {
				httpwriter.Write(w, http.StatusInternalServerError, err)
				return
			}
			ctx := context.WithValue(r.Context(), SQLTxCtxKey{}, tx)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
