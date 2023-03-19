package main

import (
	"github.com/magefile/mage/sh"
)

var (
	Default = Build // nolint: deadcode
)

// Generate generates mock implementations of interfaces.
func Generate() (err error) { // nolint: deadcode
	return sh.RunV("cmg", "gen", "./...")
}

// Build builds the binaries.
func Build() error { // nolint: deadcode
	return sh.RunV("go", "build", "./cmd/presence")
}

// Test runs the test suite.
func Test() error { // nolint: deadcode
	return sh.RunV("go", "test", "-cover", "-race", "./...")
}
