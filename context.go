package slowhttp

import (
	"context"
	"net/http"
	"reflect"
	"runtime"
)

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	switch {
	case e.Type == nil:
		return "slowhttp: Unmarshal(nil)"
	case e.Type.Kind() != reflect.Ptr:
		return "slowhttp: Unmarshal(non-pointer " + e.Type.String() + ")"
	default:
		return "slowhttp: Unmarshal(nil " + e.Type.String() + ")"
	}
}

func (e *InvalidUnmarshalError) Code() int {
	return http.StatusInternalServerError
}

func GetContext(ctx context.Context, v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v)}
	}
	rv = rv.Elem()

	for i := 0; i < rv.NumField(); i++ {
		key := rv.Type().Field(i).Tag.Get("ctx")
		if key == "" || key == "-" {
			continue
		}
		rv.Field(i).Set(reflect.ValueOf(ctx.Value(key)))
	}

	return nil
}
