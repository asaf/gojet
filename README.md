# GoJet

GoJet is a CLI tool to automate testing of HTTP APIs, written in Golang.

While unit tests aims to test internal functions and written by developers,
_acceptance / integration tests_ aims to test high level API and can be written by QA.

GoJet can run as a part of _CICD_ pipeline as one would do with standard unit tests. 

We struggled finding a descriptive approach to write _integration tests_ for our RESTful API, the result is GoJet.
