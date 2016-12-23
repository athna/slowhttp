package slowhttp

import (
	"fmt"
	"golang.org/x/net/context"
	"net/http"
	"strings"
)

type matchResult struct {
	Path    string
	Handler http.HandlerFunc
}

func (r *matchResult) Parse(ctx context.Context, path string) context.Context {
	p := r.Path
	for i := moveUntil(p, 0, ':'); i < len(p); i = moveUntil(p, i, ':') {

		j := moveUntil(p, i, '/')
		key := p[i+1 : j]
		p = p[j:]

		j = moveUntil(path, i, '/')
		val := path[i:j]
		path = path[j:]

		ctx = context.WithValue(ctx, key, val)
	}
	return ctx
}

type matchState struct {
	Value string
	Next  [256]*matchState

	Result *matchResult
}

func newState(val string, h http.HandlerFunc) *matchState {
	var nv string
	for i, j := 0, 0; i < len(val); i = moveUntil(val, j+1, '/') {
		if j = moveUntil(val, i, ':'); j == len(val) {
			j--
		}
		nv += val[i : j+1]
	}
	return &matchState{
		Value: nv,
		Result: &matchResult{
			Path:    val,
			Handler: h,
		},
	}
}

func (s *matchState) fst() byte {
	return s.Value[0]
}

func (s *matchState) String() string {
	var next string
	for _, s := range s.Next {
		if s != nil {
			next += s.String() + ", "
		}
	}
	next = strings.TrimSuffix(next, ", ")
	return fmt.Sprintf("(Value: %s Next: [%s])", s.Value, next)
}

func mergeState(s1, s2 *matchState) *matchState {
	if s1 == nil {
		return s2
	}
	if s2 == nil {
		return s1
	}
	for i := 0; i < len(s1.Value) && i < len(s2.Value); i++ {
		if s1.Value[i] == s2.Value[i] {
			continue
		}
		s := newState(s1.Value[:i], nil)

		s1.Value = s1.Value[i:]
		s2.Value = s2.Value[i:]

		s.Next[s1.fst()] = s1
		s.Next[s2.fst()] = s2
		return s
	}
	if len(s1.Value) < len(s2.Value) {
		s1, s2 = s2, s1
	}
	if len(s1.Value) != len(s2.Value) {
		s1.Value = s1.Value[len(s2.Value):]
		s2.Next[s1.fst()] = mergeState(s1, s2.Next[s1.fst()])
		return s2
	}
	for i := range s2.Next {
		s1.Next[i] = mergeState(s1.Next[i], s2.Next[i])
	}
	return s1
}

func regular(path string) string {
	return "/" + strings.Trim(path, "/")
}

func moveUntilNot(path string, idx int, delim byte) int {
	for idx < len(path) && path[idx] == delim {
		idx++
	}
	return idx
}

func moveUntil(path string, idx int, delim byte) int {
	for idx < len(path) && path[idx] != delim {
		idx++
	}
	return idx
}

func match(s *matchState, path string) *matchResult {
	path = regular(path)
	for s, idx := matchFrom(s, path, 0); s != nil; s, idx = matchFrom(s, path, idx) {
		if idx == len(path) {
			return s.Result
		}
	}
	return nil
}

func matchFrom(s *matchState, path string, idx int) (*matchState, int) {
	var fallback struct {
		s *matchState
		i int
	}
	for i := idx; s != nil; s = s.Next[path[i]] {
		for j := 0; j < len(s.Value); j++ {
			if s.Value[j] == ':' {
				i = moveUntil(path, i, '/')
				continue
			}
			if i == len(path) || path[i] != s.Value[j] {
				return fallback.s, fallback.i
			}
			if i++; path[i-1] == '/' {
				fallback.s, fallback.i = nil, 0
				i = moveUntilNot(path, i, '/')
			}
		}
		if i == len(path) {
			return s, i
		}
		if ss := s.Next[':']; ss != nil {
			fallback.s, fallback.i = ss, i
		}
	}
	return fallback.s, fallback.i
}
