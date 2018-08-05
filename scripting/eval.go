package scripting

import (
	"github.com/pkg/errors"
)

// Exp is a scripted expression to be evaluated on a vm
type Exp string

// Eval executes exp on vm, yielding some value and possible error
func Eval(vm VM, exp Exp) (interface{}, error) {
	res, err := vm.Execute(exp)
	if err != nil {
		return nil, errors.Wrap(err, "Exp evaluation error")
	}

	return res, nil
}
