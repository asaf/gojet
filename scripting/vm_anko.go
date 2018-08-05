package scripting

import (
	"github.com/mattn/anko/vm"
)

// AnkoVM is an implememntation of the VM interface using Anko scripting engine
type AnkoVM struct {
	vm *vm.Env
}

// Define conforms VM.Execute
func (a *AnkoVM) Define(name string, method interface{}) error {
	return a.vm.Define(name, method)
}

// Execute conforms VM.Execute
func (a *AnkoVM) Execute(exp Exp) (interface{}, error) {
	return a.vm.Execute(string(exp))
}

// NewAnkoVM creates an instance of AnkoVM
func NewAnkoVM() (VM, error) {
	vm := &AnkoVM{vm.NewEnv()}

	if err := define(vm); err != nil {
		return nil, err
	}

	return vm, nil
}
