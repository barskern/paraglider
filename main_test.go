package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// Test GET /api/igc
//
// Should result in:
// {
//   "uptime": <uptime>
//   "info": "Service for IGC tracks."
//   "version": "v1"
// }
func TestMetaHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/api", nil)
	res := httptest.NewRecorder()

	MetaHandler(res, req)

	var data map[string]interface{}
	if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Errorf("failed when trying to decode body as json: %s", res.Body)
	}
	if data["uptime"] == nil {
		t.Errorf("recived response: %s\ncontents of \"uptime\" does not exist", res.Body)
	}
	if data["info"] != "Service for IGC tracks." {
		t.Errorf("recived response: %s\nfailed when checking contents of \"info\", got: %s", res.Body, data["info"])
	}
	if data["version"] != "v1" {
		t.Errorf("recived response: %s\nfailed when checking contents of \"version\", got: %s", res.Body, data["version"])
	}
}
