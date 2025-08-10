//go:build tools
// +build tools

// Package tools manages development tool dependencies
package tools

import (
	// Development tools
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/vuln/cmd/govulncheck"

	// Testing tools
	_ "github.com/golang/mock/mockgen"
	_ "github.com/stretchr/testify"

	// Documentation tools
	_ "golang.org/x/tools/cmd/godoc"
)
