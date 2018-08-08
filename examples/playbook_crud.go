package main

import (
	"net/http"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"strconv"
	log "github.com/sirupsen/logrus"
)

var (
	authenticated = false
	authUser      = "admin"
	authPass      = "secret"
	token         = "1234"
	orgPosts      = []Post{
		{
			Id:    1,
			Title: "hello",
			Content: PostContent{
				Kind: "text",
				Body: "world",
			},
		},
		{
			Id:    2,
			Title: "welcome",
			Content: PostContent{
				Kind: "markdown",
				Body: "*home*",
			},
		},
	}
	idSeq = len(orgPosts)
	posts []Post
)

type Login struct {
	Token string `json:"token"`
}

type PostContent struct {
	Kind string `json:"kind"`
	Body string `json:"body"`
}

type Post struct {
	Id      int         `json:"id"`
	Title   string      `json:"title"`
	Content PostContent `json:"content"`
}

// curl -X POST 'http://localhost:9090/api/.reset'
func resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		notImplemented(w)
		return
	}

	log.WithFields(reqFields(r)).Info("reset handler invoked")
	posts = nil
	for _, p := range orgPosts {
		posts = append(posts, p)
	}
	idSeq = len(posts)
	write(w, map[string]bool{"reset": true})
}

// 200: curl http://localhost:9090/api/.login --data 'username=admin&password=secret'
// 401: curl http://localhost:9090/api/.login --data 'username=admin&password=wrong'
func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.WithFields(reqFields(r)).Info("login handler invoked")
	if r.Method != "POST" {
		notImplemented(w)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username != authUser || password != authPass {
		writeAuthFailure(w)
		return
	}

	write(w, &Login{Token: "1234"})
}

func multiHandler(w http.ResponseWriter, r *http.Request) {
	log.WithFields(reqFields(r)).Info("collection handler invoked")
	if !isAuth(r) {
		writeAuthFailure(w)
		return
	}

	switch r.Method {
	case "GET":
		// curl -H 'Token: 1234' http://localhost:9090/api/posts | jq
		write(w, posts)
	case "POST":
		// curl -H 'Token: 1234' http://localhost:9090/api/posts -d '{"title": "use gojet", "content": {"kind": "text", "body": "for awesome integration tests"}}' | jq
		var post *Post
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			writeError(w, err, http.StatusInternalServerError)
			return
		}
		log.WithFields(log.Fields{"resource": post}).Info("POST")

		if post.Content.Body == "" {
			writeError(w, fmt.Errorf("content.body is required"), http.StatusBadRequest)
			return
		}

		if post.Content.Kind == "" {
			writeError(w, fmt.Errorf("content.kind is required"), http.StatusBadRequest)
			return
		}

		idSeq++
		post.Id = idSeq
		posts = append(posts, *post)

		w.WriteHeader(http.StatusCreated)
		write(w, post)
	default:
		notImplemented(w)
	}
}

func singleHandler(w http.ResponseWriter, r *http.Request) {
	log.WithFields(reqFields(r)).Info("single handler invoked")
	if !isAuth(r) {
		writeAuthFailure(w)
		return
	}

	_, resId := extractColAndResourceId(r.URL)
	id, err := strconv.Atoi(resId)
	if err != nil {
		writeError(w, fmt.Errorf("id must be int"), http.StatusBadRequest)
		return
	}

	idx, p := getPostById(id);
	if p == nil {
		writeError(w, fmt.Errorf("post [%d] not found", id), http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		// 200: curl -H 'Token: 1234' http://localhost:9090/api/posts/1 |jq
		write(w, p)
	case "PUT":
		// curl -X PUT -H 'Token: 1234' http://localhost:9090/api/posts/1 -d '{"title": "use gojet", "content": {"kind": "text", "body": "for awesome integration tests"}}' | jq
		var post *Post
		err := json.NewDecoder(r.Body).Decode(&post)
		if err != nil {
			writeError(w, err, http.StatusInternalServerError)
			return
		}

		log.WithFields(log.Fields{"resource": post}).Info("PUT")

		if post.Content.Body == "" {
			writeError(w, fmt.Errorf("content.body is required"), http.StatusBadRequest)
			return
		}

		if post.Content.Kind == "" {
			writeError(w, fmt.Errorf("content.kind is required"), http.StatusBadRequest)
			return
		}

		post.Id = p.Id
		posts[idx] = *post
		write(w, post)
	case "DELETE":
		// curl -X DELETE -H 'Token: 1234' http://localhost:9090/api/posts/3
		w.WriteHeader(http.StatusNoContent)
		posts = append(posts[:idx], posts[idx+1:]...)
	default:
		notImplemented(w)
	}
}

func notImplemented(w http.ResponseWriter) {
	writeError(w, fmt.Errorf("not implemented"), http.StatusNotImplemented)
}

func writeAuthFailure(w http.ResponseWriter) {
	writeError(w, fmt.Errorf("not authenticated"), http.StatusUnauthorized)
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

func isAuth(request *http.Request) bool {
	return request.Header.Get("token") == token
}

func extractColAndResourceId(url *url.URL) (string, string) {
	urlParts := strings.Split(url.Path, "/")[2:]
	if len(urlParts) == 1 {
		return urlParts[0], ""
	} else if len(urlParts) == 2 {
		return urlParts[0], urlParts[1]
	} else {
		return "", ""
	}
}

func getPostById(id int) (int, *Post) {
	for i, p := range posts {
		if p.Id == id {
			return i, &p
		}
	}

	return -1, nil
}

func reqFields(r *http.Request) log.Fields {
	return log.Fields{"method": r.Method, "url": r.URL, "header": r.Header}
}

func main() {
	bind := ":9090"
	for _, p := range orgPosts {
		posts = append(posts, p)
	}
	http.HandleFunc("/api/.reset", resetHandler)
	http.HandleFunc("/api/.login", loginHandler)
	http.HandleFunc("/api/posts", multiHandler)
	http.HandleFunc("/api/posts/", singleHandler)
	if err := http.ListenAndServe(bind, nil); err != nil {
		panic(err)
	}
}
