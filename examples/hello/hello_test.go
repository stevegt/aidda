package main

import (
	"testing"

	. "github.com/stevegt/goadapt"
)

func TestHelloWorld(t *testing.T) {
	out := hello()
	Tassert(t, out == "Hello, World!", "Expected 'Hello, World!', got '%s'", out)
}
