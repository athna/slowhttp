package slowhttp

import (
	"context"
	"net/http"
)

type ErrorCode interface {
	Code() int
}

type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request) (context.Context, error)

func MakeHTTPHandler(fs ...HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		ctx := context.Background()
		for _, f := range fs {
			if ctx, err = f(ctx, w, r); err != nil {
				code := http.StatusInternalServerError
				if err, ok := err.(ErrorCode); ok {
					code = err.Code()
				}
				http.Error(w, err.Error(), code)
				return
			}
		}
	}
}
