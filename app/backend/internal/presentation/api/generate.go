// Package api holds the Go types and the chi server interface generated from
// api/openapi.yaml. The controllers implement the interface (ADR-0011, ADR-0017).
// Run `make generate` after changing the document; never edit openapi.gen.go by hand.
package api

//go:generate go tool oapi-codegen -config oapi-codegen.yaml ../../../api/openapi.yaml
