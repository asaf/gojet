package model

import (
	"github.com/sirupsen/logrus"
	"github.com/asaf/gojet/placeholders"
	"github.com/pkg/errors"
	"os"
	"strings"
	)

// Vars is a map of flat key value pairs
type Vars map[string]interface{}

func (vars Vars) AddOf(key string, value interface{}) {
	logrus.WithFields(logrus.Fields{"key": key, "value": value}).Debug("adding var")
	vars[key] = value
}

func (vars Vars) Resolve(withEnvVars bool) error {
	envs := map[string]interface{}{}
	if withEnvVars {
		for _, e := range os.Environ() {
			pair := strings.Split(e, "=")
			envs[pair[0]] = pair[1]
		}
	}

	for k, v := range vars {
		switch v.(type) {
		case string:
			resolved, err := placeholders.Resolve(v.(string), vars, envs)
			if err != nil {
				return errors.Wrapf(err, "failed to resolve var [%s:%s]", k, v)
			}
			if k != resolved {
				vars[k] = resolved
			}
		}
	}
	return nil
}
