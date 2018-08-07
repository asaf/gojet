package model

// httpMethod is a type of all possible HTTP methods
type httpMethod string

const (
	GET     httpMethod = "GET"
	POST    httpMethod = "POST"
	PUT     httpMethod = "PUT"
	PATCH   httpMethod = "PATCH"
	DELETE  httpMethod = "DELETE"
	OPTIONS httpMethod = "OPTIONS"
	HEAD    httpMethod = "HEAD"
)

// Playbook is a named stages composition
type Playbook struct {
	Name   string   `json:"name"`
	Vars   Vars     `json:"vars"`
	Stages []*Stage `json:"stages"`
}

// Stage is a test to be executed within a playbook
type Stage struct {
	Name     string    `json:"name"`
	Request  *Request  `json:"request"`
	Response *Response `json:"response"`
}

// Request describes an http request to be executed as apart of a stage
type Request struct {
	Url     string                 `json:"url"`
	Method  httpMethod             `json:"method"`
	Json    map[string]interface{} `json:"json"`
	Query   map[string]string      `json:"query"`
	Headers map[string]string      `json:"headers"`
}

// Request describes an http response to be asserted as apart of a stage
type Response struct {
	Code int                    `json:"code"`
	Body map[string]interface{} `json:"body"`
	Save *SaveResp              `json:"save"`
}

type SaveResp struct {
	Body map[string]string `json:"body"`
}
