package cmds

import (
	"github.com/asaf/gojet/model"
	"net/http"
	"github.com/pkg/errors"
	"encoding/json"
	"github.com/oliveagle/jsonpath"
	"strings"
	"fmt"
	"bytes"
	log "github.com/sirupsen/logrus"
)

//RunPlaybook runs a playbook and yields an Assertions per stage
// todo: split into more logical components
// todo: write unit tests using a local http server
func RunPlaybook(man *model.Playbook, vars model.Vars) (map[string]*model.Assertions, error) {
	c := http.DefaultClient

	as := map[string]*model.Assertions{}
	for _, st := range man.Stages {
		a := model.NewAssertions()
		as[st.Name] = a

		httpReq, err := buildHttpReqForStage(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create http request for stage [%s]", st)
		}

		// (1) do request
		//
		httpResp, err := c.Do(httpReq)
		if err != nil {
			// stage failed hardly!
			a.AddOf(model.StatusAssertion, "", "503", "hard network failure")
			continue
		}

		resp := st.Response
		// (2) assert status code
		//
		a.AddOf(model.StatusAssertion, resp.Code, httpResp.StatusCode, httpResp.Status)

		// (3) body (json) assertions
		//
		var jsonBody interface{}
		err = json.NewDecoder(httpResp.Body).Decode(&jsonBody)
		if err != nil {
			a.AddOf(model.BodyAssertion, "json body", "non json body", err.Error())
		}

		// handle each body assertion
		for k, exp := range st.Response.Body {
			path := k
			if !strings.HasPrefix(k, "$.") {
				path = "$." + k
			}

			// todo: handle better
			actual, err := jsonpath.JsonPathLookup(jsonBody, path)
			if err != nil {
				fmt.Println("path error: ", err)
				a.AddOf(model.BodyAssertion, exp, actual, fmt.Sprintf("%v", actual))
			} else {
				a.AddOf(model.BodyAssertion, exp, actual, fmt.Sprintf("%v", actual))
			}
		}
	}

	return as, nil
}

// buildHttpReqForStage creates an http request for stage
func buildHttpReqForStage(stage *model.Stage) (*http.Request, error) {
	req := stage.Request

	body, err := json.Marshal(req.Json)
	if err != nil {
		return nil, errors.Wrap(err, "body cannot be a json")
	}

	httpReq, err := http.NewRequest(string(stage.Request.Method), stage.Request.Url, bytes.NewBuffer(body))

	h := http.Header{}
	for k, v := range req.Headers {
		h.Add(k, v)
	}

	httpReq.Header = h

	log.WithFields(log.Fields{"stage": stage.Name, "req": httpReq}).Debug("http request created")
	return httpReq, err
}
