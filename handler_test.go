package slowhttp

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ErrBadReq struct{}

func (e *ErrBadReq) Error() string {
	return "bad request"
}
func (e *ErrBadReq) Code() int {
	return 400
}

func TestMakeHTTPHandler(t *testing.T) {
	cls := Class{"A"}
	usr := User{"1", "fancl20"}
	w := httptest.NewRecorder()

	getClass := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
		return context.WithValue(ctx, "class", cls), nil
	}
	getUser := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
		return context.WithValue(ctx, "user", usr), nil
	}
	response := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
		var arg struct {
			Class Class `ctx:"class"`
			User  User  `ctx:"user"`
		}
		GetContext(ctx, &arg)
		_, err := w.Write([]byte(fmt.Sprintf("%+v %+v", arg.Class, arg.User)))
		return ctx, err
	}
	respErr := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
		return ctx, &ErrBadReq{}
	}

	handler := MakeHTTPHandler(getClass, getUser, response)
	handler(w, nil)
	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "{Name:A} {UID:1 Name:fancl20}", w.Body.String())

	w = httptest.NewRecorder()
	handler = MakeHTTPHandler(getClass, respErr, getUser, response)
	handler(w, nil)
	assert.Equal(t, 400, w.Code)
	assert.Equal(t, "bad request\n", w.Body.String())
}
