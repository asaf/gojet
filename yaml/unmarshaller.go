package yaml

import (
	"reflect"
	"fmt"
	"github.com/sanathkr/go-yaml"
	"github.com/asaf/gojet/scripting"
)

type ExpUnmarshaler struct{}

func (t *ExpUnmarshaler) UnmarshalYAMLTag(tag string, fieldValue reflect.Value) reflect.Value {
	exp := scripting.Exp(fmt.Sprintf("%s", fieldValue))
	return reflect.ValueOf(exp)
}

func Unmarshal(in []byte, out interface{}) error {
	yaml.RegisterTagUnmarshaler("!exp", &ExpUnmarshaler{})
	return yaml.Unmarshal(in, out)
}
