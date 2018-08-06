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

	err = vars.Resolve(true)
	assert.Nil(t, err)

	assert.Len(t, vars, 5)
	assert.Equal(t, "foo", vars["token"])
	assert.Equal(t, "http://localhost", vars["url"])
	assert.Equal(t, "http://localhost/api", vars["api_url"])
	assert.NotNil(t, vars["p"])
}
