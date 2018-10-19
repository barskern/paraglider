package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

var igcServer = NewIgcServer()

// Test GET / of the igc-server
//
// Should result in JSON object with the following structure:
//
// ```json
// {
//   "uptime": <uptime>
//   "info": "Service for IGC tracks."
//   "version": "v1"
// }
// ```
func TestIgcServerGetMeta(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	igcServer.ServeHTTP(res, req)

	var data map[string]interface{}
	if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Errorf("received response body: '%s'", res.Body)
		t.Fatalf("failed when trying to decode body as json")
	}
	if data["uptime"] == nil {
		t.Errorf("\"uptime\" does not exist")
	}
	if data["info"] == nil {
		t.Errorf("\"info\" does not exist")
	}
	if data["version"] == nil {
		t.Errorf("\"version\" does not exist")
	}
}

// Test GET /rubbish of the igc-server
//
// Should result in a 404 response
//
func TestIgcServerGetRubbish(t *testing.T) {
	req := httptest.NewRequest("GET", "/rubbish", nil)
	res := httptest.NewRecorder()

	igcServer.ServeHTTP(res, req)

	code := res.Result().StatusCode
	if code != 404 {
		t.Fatalf("expected `GET /rubbish` to return a 404, got %d", code)
	}
}

// Test PUT / of the igc-server
//
// Should result in a 405 response
//
func TestIgcServerPutMethod(t *testing.T) {
	req := httptest.NewRequest("PUT", "/", nil)
	res := httptest.NewRecorder()

	igcServer.ServeHTTP(res, req)

	code := res.Result().StatusCode
	if code != 405 {
		t.Fatalf("expected `PUT /` to return a 405, got '%d'", code)
	}
}
