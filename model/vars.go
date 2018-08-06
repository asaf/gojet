package model

import "github.com/sirupsen/logrus"

// Vars is a map of flat key value pairs
type Vars map[string]interface{}

func (v Vars) AddOf(key string, value interface{}) {
	logrus.WithFields(logrus.Fields{"key": key, "value": value}).Debug("adding var")
	v[key] = value
}