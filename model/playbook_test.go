package model

import (
	"testing"
	"io/ioutil"
	"github.com/stretchr/testify/assert"
		"github.com/asaf/gojet/yaml"
	"github.com/asaf/gojet/scripting"
)

func TestPlaybook_Unmarshal(t *testing.T) {
	f, err := ioutil.ReadFile("./playbook_test.yml")
	assert.Nil(t, err)
	var pb *Playbook
	err = yaml.Unmarshal(f, &pb)
	assert.Nil(t, err)
	// top level assertions
	//
	assert.Equal(t, "Sample Playbook", pb.Name)
	assert.Len(t, pb.Stages, 2)

	// stage level assertions
	//
	// stage 1
	//
	//
	s1 := pb.Stages[0]
	assert.Equal(t, "s1", s1.Name, "stages order should be preserved")

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
	assert.Equal(t, int(1), body["id"])

	// stage 2
	//
	//
	s2 := pb.Stages[1]
	assert.Equal(t, "s2", s2.Name, "stages order should be preserved")
	req = s2.Request
	assert.Equal(t, "https://server/posts", req.Url)
	assert.Equal(t, POST, req.Method)
	assert.Len(t, req.Headers, 1)
	assert.Equal(t, "application/json", req.Headers["content-type"])

	// response level assertions
	resp = s2.Response
	body = resp.Body
	assert.Equal(t, scripting.Exp("total > 0"), body["total"])
}