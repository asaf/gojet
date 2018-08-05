package model

import (
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestVars(t *testing.T) {
	f, err := ioutil.ReadFile("./vars.yml")
	assert.Nil(t, err)
	var vars Vars
	err = yaml.Unmarshal(f, &vars)
	assert.Nil(t, err)

	assert.Len(t, vars, 2)
	assert.Equal(t, "foo", vars["token"])
	assert.Equal(t, "http://localhost", vars["api_url"])
}
