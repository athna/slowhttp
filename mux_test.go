package slowhttp

import (
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

var addrs = []string{
	"/abc/ef/xxx",
	"/abc/e",
	"/abc/ex/ccc",
	"/abc/ef/xc",
}

func run(t *testing.T, s *matchState, i int) *matchState {
	s = mergeState(s, newState(addrs[i], func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(addrs[i]))
	}))
	for j := range addrs {
		res := match(s, addrs[j])
		if j > i {
			assert.Nil(t, res, s.String()+" "+addrs[j])
		} else {
			w := httptest.NewRecorder()
			res.Handler(w, nil)
			assert.Equal(t, addrs[j], w.Body.String(), s.String()+" "+addrs[j])
		}
	}
	return s
}

func TestMux(t *testing.T) {
	var s *matchState
	for i := range addrs {
		s = run(t, s, i)
	}

	res := match(s, "/abc//ex///ccc///")
	w := httptest.NewRecorder()
	res.Handler(w, nil)
	assert.Equal(t, "/abc/ex/ccc", w.Body.String(), s.String())

	s = mergeState(s, newState("/abc/:test/xxx", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("/abc/:test/xxx"))
	}))
	t.Log(s)
	res = match(s, "/abc/eeeee/xxx")
	w = httptest.NewRecorder()
	res.Handler(w, nil)
	ctx := res.Parse(context.Background(), "/abc/eeeee/xxx")
	assert.Equal(t, "eeeee", ctx.Value("test"), s.String())
	assert.Equal(t, "/abc/:test/xxx", w.Body.String(), s.String())
}
