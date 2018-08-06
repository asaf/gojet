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
	"github.com/asaf/gojet/scripting"
)

//RunPlaybook runs a playbook and yields an Assertions per stage
// todo: split into more logical components
// todo: write unit tests using a local http server
func RunPlaybook(man *model.Playbook, vars model.Vars) (map[string]*model.Assertions, error) {
	c := http.DefaultClient

	as := map[string]*model.Assertions{}
	for _, st := range man.Stages {
		vm, err := scripting.NewAnkoVM()
		if err != nil {
			return nil, err
		}

		a := model.NewAssertions()
		as[st.Name] = a

		httpReq, err := createHttpRequestOfRequest(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create http request for stage [%s]", st)
		}

		resp := st.Response

		// (1) do request
		//
		httpResp, err := c.Do(httpReq)
		if err != nil {
			// stage failed hardly!
			a.AddOf(model.StatusAssertion, resp.Code, "503", "hard network failure")
			continue
		}

		// (2) assert status code
		//
		a.AddOf(model.StatusAssertion, int(resp.Code), httpResp.StatusCode, httpResp.Status)

		// (3) body (json) assertions
		//
		var jsonBody map[string]interface{}
		err = json.NewDecoder(httpResp.Body).Decode(&jsonBody)
		if err != nil {
			a.AddOf(model.BodyAssertion, "json body", "non json body", err.Error())
		}
		log.WithFields(log.Fields{"resp-body": jsonBody}).Debug("http response received")

		for k, v := range jsonBody {
			vm.Define(k, v)
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
				log.WithFields(log.Fields{"cause": err, "stage": st.Name, "path": path}).Error("error in path")
			} else {
				var expVal interface{}
				switch v := exp.(type) {
				case int:
					// todo: this is for the time being until yaml gets upgraded to latest ver that returns float64 instead of int
					expVal = float64(v)
				case scripting.Exp:
					vm.Define("val", actual)
					expRes, err := scripting.Eval(vm, v)
					if err != nil {
						// handle!
						// todo im prove error handling
						return nil, err
					}
					log.WithFields(log.Fields{"exp": exp, "value": expRes}).Debug("evaluated expected value expression")
					expVal = expRes
					actual = true
				default:
					expVal = exp
				}

				a.AddOf(model.BodyAssertion, expVal, actual, fmt.Sprintf("%v", actual))
			}
		}
	}

	return as, nil
}

// createHttpRequestOfRequest creates an http request for stage
func createHttpRequestOfRequest(stage *model.Stage) (*http.Request, error) {
	req := stage.Request

	body, err := json.Marshal(req.Json)
	if err != nil {
		return nil, errors.Wrap(err, "body cannot be a json")
	}

	httpReq, err := http.NewRequest(string(stage.Request.Method), stage.Request.Url, bytes.NewBuffer(body))
	q := httpReq.URL.Query()
	for k, v := range req.Query {
		q.Add(k, v)
	}
	httpReq.URL.RawQuery = q.Encode()

	h := http.Header{}
	for k, v := range req.Headers {
		h.Add(k, v)
	}

	httpReq.Header = h

	log.WithFields(log.Fields{"stage": stage.Name, "req-method": httpReq.Method, "req-header": httpReq.Header, "req-body": httpReq.Body, "req-url": httpReq.URL}).Debug("http request created")
	return httpReq, err
}
