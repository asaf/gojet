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
	"time"
	"github.com/asaf/gojet/placeholders"
)

//RunPlaybook runs a playbook and yields an Assertions per stage
// todo: split into more logical components
// todo: write unit tests using a local http server
func RunPlaybook(pbook *model.Playbook, vars model.Vars) (map[string]*model.Assertions, error) {
	log.WithFields(log.Fields{"name": pbook.Name}).Debug("playbook execution started")
	elapsed := time.Now()
	c := http.DefaultClient

	as := map[string]*model.Assertions{}
	for _, st := range pbook.Stages {
		log.WithFields(log.Fields{"name": st.Name}).Debug("executing stage started")
		stElapsed := time.Now()

		a := model.NewAssertions()
		as[st.Name] = a

		if st.Request == nil {
			return nil, fmt.Errorf("request in stage [%s] is required", st.Name)
		}

		// GET is default
		if st.Request.Method == "" {
			st.Request.Method = http.MethodGet
		}

		if st.Response == nil {
			st.Response = &model.Response{
				Code: http.StatusOK,
			}
		}

		if st.Response.Code == 0 {
			st.Response.Code = http.StatusOK
		}

		resp := st.Response

		// (1) resolve stage placeholder
		if err := resolveStagePlaceholders(st, vars); err != nil {
			return nil, errors.Wrap(err, "failed to resolve stage placeholders")
		}

		// (2) create http request
		httpReq, err := createHttpRequestOfRequest(st)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create http request for stage [%s]", st.Name)
		}

		// (3) do request
		//
		httpResp, err := c.Do(httpReq)
		if err != nil {
			// stage failed hardly!
			a.AddOf(model.StatusAssertion, resp.Code, "503", "hard network failure")
			continue
		}

		// (4) perform assertions
		//
		// assert status code
		a.AddOf(model.StatusAssertion, int(resp.Code), httpResp.StatusCode, httpResp.Status)

		// body (json) assertions
		//
		// expecting body to be a json
		var jsonBody map[string]interface{}
		err = json.NewDecoder(httpResp.Body).Decode(&jsonBody)
		if err != nil {
			//a.AddOf(model.BodyAssertion, "json body", "non json body", err.Error())
			if st.Request.Method != http.MethodGet {
				log.WithFields(log.Fields{"stage": st.Name}).Debug("body is not a json")
			}
		}
		log.WithFields(log.Fields{"resp-body": jsonBody}).Debug("http response received")

		vm, err := createVMForStage(st, jsonBody)
		if err != nil {
			return nil, errors.Wrapf(err, "error creating VM for stage [%s]", st.Name)
		}

		// handle each body assertion
		// body is constructed of a path (jsonpath) -> value / expression
		for path, valOrExp := range st.Response.Body {
			// resolvedPath is the value that path points to and is the actual value
			actualValue, err := findPath(jsonBody, path)
			if err != nil {
				return nil, errors.Wrap(err, "error in path")
				log.WithFields(log.Fields{"cause": err, "stage": st.Name, "path": path}).Error("error in path")
			}

			// can be nil?
			if actualValue != nil {
				// override current value in question
				vm.Define("v", actualValue)
				vm.Define("val", actualValue)
				// valOrExp is the expected value or an expression that should yield a bool (whether assertion is true or false)
				expectedVal, isExp, err := evalPossibleExp(vm, valOrExp)
				if err != nil {
					return nil, errors.Wrapf(err, "error evaluating exp [%s]", valOrExp)
				}
				if !isExp {
					// not an exp so value is primitive
					a.AddOf(model.BodyAssertion, actualValue, expectedVal, fmt.Sprintf("e: %v, a: %v", expectedVal, actualValue))
				} else {
					// an exp, so value is a true/false
					a.AddOf(model.BodyAssertion, true, expectedVal, fmt.Sprintf("e: %v, a: %v", expectedVal, actualValue))
				}
			}
		}

		// (5) save vars
		if resp.Save != nil {
			if resp.Save.Body != nil {
				// k is the var name to be saved where p is the path in body
				for k, p := range resp.Save.Body {
					v, err := findPath(jsonBody, p)
					if err != nil {
						return nil, errors.Wrap(err, "failed to resolve path to body to be saved")
					}
					vars.AddOf(k, v)
				}
			}
		}

		log.WithFields(log.Fields{"name": st.Name, "elapsed": time.Since(stElapsed)}).Debug("executing stage completed")
	}

	log.WithFields(log.Fields{"elapsed": time.Since(elapsed)}).Debug("playbook execution completed")
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
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
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

// findPath finds the path in obj and yields the path evaluation result
func findPath(obj map[string]interface{}, path string) (interface{}, error) {
	if !strings.HasPrefix(path, "$.") {
		path = "$." + path
	}

	result, err := jsonpath.JsonPathLookup(obj, path)
	if err != nil {
		return nil, errors.Wrap(err, "unresolved path")
	}

	return result, nil
}

// evalExp determines if valOrExp is an expression and if yes it evaluates it on vm, otherwise it just returns it as is
func evalPossibleExp(vm scripting.VM, valOrExp interface{}) (interface{}, bool, error) {
	var expVal interface{}
	switch v := valOrExp.(type) {
	case int:
		// todo: this is for the time being until yaml gets upgraded to latest ver that returns float64 instead of int
		expVal = float64(v)
		return expVal, false, nil
	case scripting.Exp:
		expRes, err := scripting.Eval(vm, v)
		if err != nil {
			return nil, true, errors.Wrapf(err, "failed to eval exp [%s]", v)
		}
		//log.WithFields(log.Fields{"exp": exp, "value": expRes}).Debug("evaluated expected value expression")
		return expRes, true, nil
	default:
		return valOrExp, false, nil
	}
}

// resolveStagePlaceholders resolves all placeholders of st by vars
func resolveStagePlaceholders(st *model.Stage, vars model.Vars) error {
	// resolve url
	//
	req := st.Request
	resolvedUrl, err := placeholders.Resolve(req.Url, vars)
	if err != nil {
		return err
	}
	req.Url = resolvedUrl.(string)

	// resolve query params
	q := req.Query
	for k, v := range q {
		val, err := placeholders.Resolve(v, vars)
		if err != nil {
			return fmt.Errorf("cannot resolve query param [%s:%s]", k, v)
		}
		q[k] = fmt.Sprintf("%v", val)
	}

	// resolve header
	h := req.Headers
	for k, v := range h {
		val, err := placeholders.Resolve(v, vars)
		if err != nil {
			return fmt.Errorf("cannot resolve header [%s:%s]", k, v)
		}
		h[k] = fmt.Sprintf("%v", val)
	}
	return nil
}

func createVMForStage(st *model.Stage, body map[string]interface{}) (scripting.VM, error) {
	vm, err := scripting.NewAnkoVM()
	if err != nil {
		return nil, err
	}

	for k, v := range body {
		vm.Define(k, v)
	}

	return vm, nil
}
