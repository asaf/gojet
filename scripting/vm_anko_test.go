package scripting

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewAnkoVM(t *testing.T) {
	vm, err := NewAnkoVM()
	assert.Nil(t, err)
	assert.NotNil(t, vm)
	res, err := vm.Execute("1 == 1")
	assert.Nil(t, err)
	assert.Equal(t, true, res)
}
