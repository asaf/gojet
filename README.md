# GoJet

[![Build Status](https://travis-ci.com/asaf/gojet.svg?branch=master)](https://travis-ci.com/asaf/gojet)

GoJet is a CLI tool to automate testing of HTTP APIs, written in Golang.

While unit tests aims to test internal functions and written by developers,
_acceptance / integration tests_ aims to test high level API and can be written by QA.

GoJet can run as a part of _CICD_ pipeline as one would do with standard unit tests. 

We struggled finding a descriptive approach to write _integration tests_ for our RESTful API, the result is GoJet.


# Playbook

A playbook is a composition of stages where each stage represents an http test,

here is a simple playbook with single stage that perform _GET_ http request to get a blog post:

```yml
name: "simplest playbook"
stages:
- name: "get a post"
  request:
    url: "https://jsonplaceholder.typicode.com/posts/1"
```

to run a playbook simply run: `gojet playbook run --file <file>.yml` 

output example:

```bash
playing simplest playbook
stage get a post
[SUCCESS: 200 OK] status 
```

explanation:

- GET http request is submitted (as no custom _method_ was specified) to the specified url
- Default status assertion expects status code to be _200 OK_
