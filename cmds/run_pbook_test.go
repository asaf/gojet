package cmds

import (
	"testing"
	"github.com/asaf/gojet/model"
	"net/http"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	)

func TestCreateHttpRequestOfRequest(t *testing.T) {
	st := &model.Stage{
		Request: &model.Request{
			Url:    "/foo",
			Method: http.MethodPost,
			Json: map[string]interface{}{
				"title": "hello",
				"body":  map[string]interface{}{
					"content": "world",
				},
			},
			Query:   map[string]string{"filter": "foo=bar"},
			Headers: map[string]string{"content": "json"},
		},
	}
	hreq, err := createHttpRequestOfRequest(st)
	assert.Nil(t, err)
	assert.Equal(t, "/foo", hreq.URL.Path)
	assert.Equal(t, "POST", hreq.Method)
	// assert query
	//
	assert.Len(t, hreq.URL.Query(), 1)
	assert.Equal(t, "foo=bar", hreq.URL.Query().Get("filter"))
	// assert header
	//
	assert.Len(t, hreq.Header, 1)
	assert.Equal(t, "json", hreq.Header.Get("content"))
	// assert json body
	bodyBytes, err := ioutil.ReadAll(hreq.Body)
	assert.Nil(t, err)
	assert.Equal(t, `{"body":{"content":"world"},"title":"hello"}`, string(bodyBytes))
}
