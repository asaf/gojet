package scripting

import "fmt"

// VM is an interface of a virtual machine evaluates expressions
type VM interface {
	// Define binds a new method to the vm
	Define(name string, method interface{}) error
	// Execute evaluates exp and returns the yielded value
	Execute(exp Exp) (interface{}, error)
}

// define registers some default methods in vm
func define(vm VM) error {
	err := vm.Define("echo", fmt.Print)
	if err != nil {
		return err
	}

	return nil
}
