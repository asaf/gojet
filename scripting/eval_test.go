package scripting

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEval_Anko(t *testing.T) {
	vm, err := NewAnkoVM()
	vm.Define("foo", "world")
	assert.Nil(t, err)
	v, err := Eval(vm, "return 'hello ' + foo")
	assert.Nil(t, err)
	assert.Equal(t, "hello world", v)
}
