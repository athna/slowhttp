package slowhttp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Class struct {
	Name string
}

type User struct {
	UID  string
	Name string
}

func TestGetContext(t *testing.T) {
	var err error

	cls := Class{"A"}
	usr := User{"1", "fancl20"}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "class", cls)
	ctx = context.WithValue(ctx, "user", usr)
	ctx = context.WithValue(ctx, "foo", "foo")

	var arg struct {
		Foo   string
		Class Class `ctx:"class"`
		User  User  `ctx:"user"`
	}
	err = GetContext(ctx, &arg)
	assert.Nil(t, err)
	assert.Equal(t, cls, arg.Class)
	assert.Equal(t, usr, arg.User)
	assert.Equal(t, "", arg.Foo)

	err = GetContext(ctx, arg)
	assert.NotNil(t, err)
}
