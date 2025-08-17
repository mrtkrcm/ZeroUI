//go:build tools
// +build tools

// Package tools manages development tool dependencies
package tools

import (
	// Development tools
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/vuln/cmd/govulncheck"
	_ "mvdan.cc/gofumpt"

	// Code quality and linting
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "honnef.co/go/tools/cmd/staticcheck"

	// Testing tools
	_ "github.com/golang/mock/mockgen"
	_ "github.com/stretchr/testify"
	_ "gotest.tools/gotestsum"

	// Performance and profiling
	_ "github.com/google/pprof"
	_ "golang.org/x/perf/cmd/benchstat"

	// Documentation tools
	_ "golang.org/x/tools/cmd/godoc"
	_ "github.com/swaggo/swag/cmd/swag"

	// Development utilities
	_ "github.com/air-verse/air"
)
