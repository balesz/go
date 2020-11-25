package callable_test

import (
	"testing"

	"github.com/balesz/go/firebase/functions/callable"
)

func TestMain(t *testing.T) {
	ctx := callable.Context{
		Data: map[string]interface{}{
			"foo": "bar",
			"hello": map[string]interface{}{
				"world": true,
			},
		},
	}

	var data struct {
		Foo   string `json:"foo"`
		Hello struct {
			World bool `json:"world"`
		} `json:"hello"`
	}

	if err := ctx.GetData(&data); err != nil {
		t.Error(err)
	} else if data.Foo != "bar" {
		t.Error("data.Foo is not bar")
	} else if data.Hello.World != true {
		t.Error("data.Hello.World is not true")
	}
}
