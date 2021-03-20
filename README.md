# glogs-to-honeycomb

[![Project status](https://img.shields.io/github/release/tmc/glogs-to-honeycomb.svg?style=flat-square)](https://github.com/tmc/glogs-to-honeycomb/releases/latest)
[![Build Status](https://github.com/tmc/glogs-to-honeycomb/workflows/Test/badge.svg)](https://github.com/tmc/glogs-to-honeycomb/actions?query=workflow%3ATest)
[![Go Report Card](https://goreportcard.com/badge/tmc/glogs-to-honeycomb?cache=0)](https://goreportcard.com/report/tmc/glogs-to-honeycomb)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/tmc/glogs-to-honeycomb)

glogs-to-honeycomb is a [Go](https://golang.org/) program that delivers Istio sidecar logs to
Honeycomb.

## Installation

Presming  you have a working Go insallation, you can install `glogs-to-honeycomb` via:

```console
go install github.com/tmc/glogs-to-honeycomb
```

## Sample Terraform Resources

See [gcp_resources.tf](./gcp_resources.tf) for sample terraform resources.
