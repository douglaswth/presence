package main

import (
	"github.com/magefile/mage/sh"
)

var (
	Default = Build // nolint: deadcode

	packagesToMock = []string{"neighbors", "wrap"}
)

// Generate generates mock implementations of interfaces.
func Generate() (err error) { // nolint: deadcode
	for _, pkg := range packagesToMock {
		err = sh.Run("mockery", "--all", "--case=underscore", "--dir="+pkg, "--exported=false", "--output="+pkg+"/mocks")
		if err != nil {
			return
		}
	}
	return
}

// Build builds the binaries.
func Build() error { // nolint: deadcode
	return sh.RunV("go", "build", "./cmd/presence")
}

// Test runs the test suite.
func Test() error { // nolint: deadcode
	return sh.RunV("go", "test", "-cover", "-race", "./...")
}
