package binder

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var (
	ErrRequiredField = errors.New("validator: missing required field")
)

type BindAndValidate interface {
	Validate() error
}

func Bind[T BindAndValidate](r *http.Request, dest BindAndValidate) error {
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &dest); err != nil {
		return err
	}

	if err = dest.Validate(); err != nil {
		return err
	}

	return nil
}
