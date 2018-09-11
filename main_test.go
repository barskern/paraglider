package main

import (
	"bytes"
	"net/http/httptest"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	// Expected return value from `helloWorld` handler
	expt := []byte("Hello world")

	// Create a fake ResponseRecorder
	res := httptest.NewRecorder()

	// Pass the fake recorder to the handler
	helloWorld(res, nil)

	// Allocate a bytebuffer to fit the return value
	body := make([]byte, len(expt))

	// Read the value of the buffer
	res.Result().Body.Read(body)

	t.Logf("Comparing \"%s\" to \"%s\"", expt, body)

	// Compare the expected and actual output
	if !bytes.Equal(body, expt) {
		t.Errorf("Expected \"%s\" but got \"%s\"", expt, body)
	}
}
