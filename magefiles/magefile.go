package main

import (
	"runtime"
	"strconv"

	"github.com/magefile/mage/sh"
)

var (
	Default = Build
)

// Generate generates mock implementations of interfaces.
func Generate() (err error) {
	return sh.RunV("go", "tool", "cmg", "gen", "-testify", "./...")
}

// Build builds the binaries.
func Build() error {
	return sh.RunV("go", "build", "./cmd/presence")
}

// Lint runs the lint suite.
func Lint() error {
	return sh.RunV("golangci-lint", "run", "./...")
}

// Test runs the test suite.
func Test() error {
	return sh.RunV("go", "test", "-cover", "-race", "./...")
}

// Snapshot runs the release snapshot.
func Snapshot() error {
	nc := runtime.NumCPU()
	return sh.RunV("goreleaser", "release", "--clean", "--parallelism", strconv.Itoa(nc), "--snapshot")
}
