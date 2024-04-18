package main

import (
	"testing"

	. "github.com/stevegt/goadapt"
)

func TestHello(t *testing.T) {
	result := Hello()
	Tassert(t, result == "Hello, World!", "Expected 'Hello, World!' but got %s", result)
}
