package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// Test metadata handler
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
func TestMetaHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api", nil)
	res := httptest.NewRecorder()

	MetaHandler(res, req)

	var data map[string]interface{}
	if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Errorf("recieved response body: '%s'", res.Body)
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
