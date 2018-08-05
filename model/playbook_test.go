package model

import (
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
	"github.com/ghodss/yaml"
)

func TestPlaybook_Marshal(t *testing.T) {
	f, err := ioutil.ReadFile("./playbook_test.yml")
	assert.Nil(t, err)
	var pb *Playbook
	err = yaml.Unmarshal(f, &pb)
	assert.Nil(t, err)
	// top level assertions
	//
	assert.Equal(t, "Sample Playbook", pb.Name)
	assert.Len(t, pb.Stages, 2)
	s1 := pb.Stages[0]
	assert.Equal(t, "s1", s1.Name)
	s2 := pb.Stages[1]
	assert.Equal(t, "s2", s2.Name)

	// stage level assertions
	//
	assert.Equal(t, "s1", s1.Name)

	// request level assertions
	//
	req := s1.Request
	assert.Equal(t, "https://server/posts/1", req.Url)
	assert.Equal(t, GET, req.Method)
	assert.Len(t, req.Headers, 1)
	assert.Equal(t, "application/json", req.Headers["content-type"])

	// response level assertions
	//
	resp := s1.Response
	assert.Equal(t, 200, resp.Code)
	// response.body level
	body := resp.Body
	assert.Len(t, body, 1)
	assert.Equal(t, float64(1), body["id"])
}
