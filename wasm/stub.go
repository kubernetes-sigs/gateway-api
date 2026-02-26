//go:build !js || !wasm

// Package main is built only for GOOS=js GOARCH=wasm; see main.go.
// This stub allows the package to load when not building for WebAssembly (e.g. go list ./..., linters).
package main
