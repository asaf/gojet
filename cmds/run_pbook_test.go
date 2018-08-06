package cmds

import (
	"testing"
	"github.com/asaf/gojet/model"
	"net/http"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http/httptest"
	"encoding/json"
	"github.com/asaf/gojet/scripting"
	"github.com/sirupsen/logrus"
)

var post = map[string]interface{}{
	"id":    1,
	"title": "hello",
	"content": map[string]interface{}{
		"type": "markdown",
	},
}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestCreateHttpRequestOfRequest(t *testing.T) {
	st := &model.Stage{
		Request: &model.Request{
			Url:     "/foo",
			Method:  http.MethodPost,
			Json:    post,
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
	assert.Equal(t, `{"content":{"type":"markdown"},"id":1,"title":"hello"}`, string(bodyBytes))
}

func TestResolveStagePlaceholders(t *testing.T) {
	st := &model.Stage{
		Request: &model.Request{
			Url:    "https://{server}/posts/1",
			Method: http.MethodPost,
			Json:   post,
			Query: map[string]string{
				"filter": "{q}",
				"foo":    "bar",
			},
			Headers: map[string]string{
				"content": "{ctype}",
			},
		},
	}

	vars := model.Vars{"server": "localhost", "q": "foo=bar", "ctype": "json"}
	err := resolveStagePlaceholders(st, vars)
	assert.Nil(t, err)
	assert.Equal(t, "https://localhost/posts/1", st.Request.Url)
	assert.Equal(t, "foo=bar", st.Request.Query["filter"], "placeholder should be resolved")
	assert.Equal(t, "bar", st.Request.Query["foo"], "static val should stay as is")
	assert.Equal(t, "json", st.Request.Headers["content"], "placeholder should be resolved")
}

func TestEvalPath(t *testing.T) {
	obj := map[string]interface{}{"foo": "bar"}
	res, err := findPath(obj, "foo")
	assert.Nil(t, err)
	assert.Equal(t, "bar", res)
	res, err = findPath(obj, "$.foo")
	assert.Equal(t, "bar", res)
}

func TestEvalPossibleExp(t *testing.T) {
	vm, err := scripting.NewAnkoVM()
	assert.Nil(t, err)
	valInt := 1
	res, isExp, err := evalPossibleExp(vm, valInt)
	assert.Nil(t, err)
	assert.False(t, isExp)
	assert.Equal(t, float64(valInt), res)

	valStr := "foo"
	res, isExp, err = evalPossibleExp(vm, valStr)
	assert.Nil(t, err)
	assert.False(t, isExp)
	assert.Equal(t, valStr, res)

	vm.Define("val", 9)
	valExp := scripting.Exp(`val == 9`)
	res, isExp, err = evalPossibleExp(vm, valExp)
	assert.Nil(t, err)
	assert.True(t, isExp)
	assert.True(t, res.(bool))
}

// TestRunPlaybook_Simplest tests the simplest playbook with single stage
func TestRunPlaybook_Simplest(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	stName := "get"

	pbook := &model.Playbook{Name: "simplest"}
	pbook.Stages = append(pbook.Stages, &model.Stage{
		Name: stName,
		Request: &model.Request{
			Url: s.URL,
		},
	})

	as, err := RunPlaybook(pbook, model.Vars{})
	assert.Nil(t, err)
	assert.Len(t, as, 1, "assertions per stage")
	as1 := as[stName].Assertions
	assert.NotNil(t, as1)
	assert.Len(t, as1, 1, "only status should be checked")
	assert.True(t, as1[0].True())
}

// TestRunPlaybook_AssertBody tests body assertions
func TestRunPlaybook_AssertBody(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		write(w, post)
	}))

	stName := "get"

	pbook := &model.Playbook{Name: "assert body pbook"}
	pbook.Stages = append(pbook.Stages, &model.Stage{
		Name: stName,
		Request: &model.Request{
			Url: s.URL,
		},
		Response: &model.Response{
			Body: map[string]interface {
			}{
				"title":        "hello",
				"content.type": scripting.Exp("v == 'markdown'"),
			},
		},
	})

	as, err := RunPlaybook(pbook, model.Vars{})
	assert.Nil(t, err)
	assert.Len(t, as, 1, "assertions per stage")
	as1 := as[stName].Assertions
	assert.NotNil(t, as1)
	assert.Len(t, as1, 3, "status and 3 body assertions")
	assert.True(t, as1[0].True())
	assert.True(t, as1[1].True())
	assert.True(t, as1[2].True())
}

// TestRunPlaybook_Save test saving vars
func TestRunPlaybook_Save(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			write(w, post)
		}
	}))

	pbook := &model.Playbook{}
	pbook.Stages = append(pbook.Stages, &model.Stage{
		Name: "s1",
		Request: &model.Request{
			Url:    s.URL,
			Method: model.POST,
		},
		Response: &model.Response{
			Save: &model.SaveResp{
				Body: map[string]string{"b": "content.type"},
			},
		},
	})

	vars := model.Vars{}
	_, err := RunPlaybook(pbook, vars)
	assert.Nil(t, err)
	assert.Len(t, vars, 1)
	assert.Equal(t, "markdown", vars["b"])
}

//TestRunPlaybook is more of a complete scenario that tests multi stages with shared vars, placeholders, etc.
func TestRunPlaybook(t *testing.T) {
	hits := 0
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		switch r.Method {
		case "POST":
			w.WriteHeader(http.StatusCreated)
			write(w, post)
		case "GET":
			assert.Len(t, r.URL.Query(), 1)
			assert.Equal(t, "1", r.URL.Query().Get("postId"))
		}
	}))

	pbook := &model.Playbook{Name: "full pbook"}
	pbook.Stages = append(pbook.Stages, &model.Stage{
		Name: "create a post",
		Request: &model.Request{
			Url:    s.URL,
			Method: model.POST,
		},
		Response: &model.Response{
			Code: http.StatusCreated,
			Save: &model.SaveResp{
				Body: map[string]string{"post_id": "id"},
			},
		},
	})

	pbook.Stages = append(pbook.Stages, &model.Stage{
		Name: "get post by id",
		Request: &model.Request{
			Url:   s.URL,
			Query: map[string]string{"postId": "{post_id}"},
		},
	})

	_, err := RunPlaybook(pbook, model.Vars{})
	assert.Nil(t, err)
	assert.Equal(t, 2, hits)

}

func TestRunPlaybook_NoReq(t *testing.T) {
	pbook := &model.Playbook{}
	pbook.Stages = append(pbook.Stages, &model.Stage{
		Name: "no request",
	})

	_, err := RunPlaybook(pbook, model.Vars{})
	assert.Error(t, err, "request is required")
}

func writeError(w http.ResponseWriter, err error, code int) {
	errObj := make(map[string]interface{})
	errObj["message"] = err.Error()
	bytes, _ := json.Marshal(errObj)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(bytes)
}

func write(w http.ResponseWriter, body interface{}) {
	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}
	w.Write(b)
}
