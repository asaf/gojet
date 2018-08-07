# GoJet

[![Build Status](https://travis-ci.com/asaf/gojet.svg?branch=master)](https://travis-ci.com/asaf/gojet)

GoJet is a CLI tool to automate testing of HTTP APIs, written in Golang.

While unit tests aims to test internal functions and written by developers,
_acceptance / integration tests_ aims to test high level API and can possible be written by automation / QA teams.

GoJet can run as a part of _CICD_ pipeline as one would do with standard unit tests. 

We struggled finding a descriptive approach to write _integration tests_ for our RESTful API that
suites our native stack, the result is GoJet.

# Quickstart

A playbook is a composition of stages where each stage represents an http test,

Here is a single stage playbook that performs an http _GET_ request for a blog post and asserts that the returned
status code is _200_:

```yml
name: "test blog REST API"
stages:
- name: "get a post 1"
  request:
    url: "https://jsonplaceholder.typicode.com/posts/1"
    method: GET
  response:
    code: 200 
```

_gojet_ is a single binary, distributed in the [release page](https://github.com/asaf/gojet/releases)


simply run a playbook by: `gojet playbook run --file <file>.yml` 

output example:

```bash
playing simplest playbook
stage get a post
[SUCCESS: 200 OK] status 
```
