package httpparam

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

var (
	ErrNotFoundPathParam  = errors.New("http: not found path param")
	ErrNotFoundQueryParam = errors.New("http: not found query parau")
)

func PathParam[T uint64 | string](r *http.Request, param string) (T, error) {
	var t T
	if v, exist := mux.Vars(r)[param]; exist {
		switch any(t).(type) {
		case uint64:
			u, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return t, err
			}
			t = any(u).(T)
		default:
			t = any(v).(T)
		}
		return t, nil
	} else {
		return t, ErrNotFoundPathParam
	}
}

func QueryParam[T any](r *http.Request, key string) (T, error) {
	var t T
	v := r.URL.Query().Get(key)
	if v == "" {
		return t, ErrNotFoundQueryParam
	}

	switch any(t).(type) {
	case uint64:
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return t, err
		}
		t = any(u).(T)
	default:
		t = any(v).(T)
	}
	return t, nil
}

func QueryParams[T any](r *http.Request, key string) ([]T, error) {
	ts := make([]T, 0)
	v := r.URL.Query().Get(key)
	if v == "" {
		return nil, ErrNotFoundQueryParam
	}
	strs := strings.Split(v, ",")
	var t T
	for _, str := range strs {
		switch any(t).(type) {
		case uint64:
			u, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				return nil, err
			}
			ts = append(ts, any(u).(T))
		default:
			ts = append(ts, any(str).(T))
		}
	}
	return ts, nil
}
