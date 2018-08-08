package scripting

import (
	"github.com/satori/go.uuid"
)

func isUUID(any interface{}) bool {
	switch v := any.(type) {
	case string:
		_, err := uuid.FromString(v)
		return err == nil
	default:
		return false
	}
}
