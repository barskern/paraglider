package igcserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

// Convenience function to create a simple igc-file hosting server which hosts
// two files, one valid 'test.igc' and an invalid 'invalid.igc'
func makeIgcFileServer() *httptest.Server {
	return httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.RequestURI == "/test.igc" {
				f, err := os.Open("../assets/test.igc")
				if err != nil {
					fmt.Printf("error when trying to read 'test.igc': %s", err)
				}
				_, err = io.Copy(w, f)
				if err != nil {
					fmt.Printf("error when trying to write file contents to response: %s", err)
				}
				fmt.Println("wrote valid igc content to response")
			} else if r.RequestURI == "/invalid.igc" {
				invalidIGC := "asljdkfjaøsljfølwer jfølvjasdløkv aøljsgødl v"
				w.Write([]byte(invalidIGC))
				fmt.Println("wrote invalid igc content to response")
			} else {
				http.Error(w, "not found", http.StatusNotFound)
				fmt.Println("wrote not found to response")
			}
		}),
	)
}

// Convenience function to create testdata to insert into the database
func makeWebhooksTestData() []WebhookInfo {
	return []WebhookInfo{
		{
			NewWebhookID([]byte("asd")),
			"http://unique.com",
			1,
			time.Now(),
		},
		{
			NewWebhookID([]byte("dsa")),
			"http://unique2.com",
			2,
			time.Now(),
		},
	}
}

// Convenience function to create testdata to insert into the database
func makeIGCTestData(serverURL string) []TrackMeta {
	return []TrackMeta{
		{
			NewTrackID([]byte("asd")),
			time.Now(),
			time.Now(),
			"Aladin Special",
			"Magical Carpet",
			"MGI2",
			1200,
			serverURL + "/aladin.igc",
		},
		{
			NewTrackID([]byte("dsa")),
			time.Now(),
			time.Now(),
			"John Normal",
			"Boeng 777",
			"BG7",
			10,
			serverURL + "/boeng.igc",
		},
	}
}

func makeTestServers() (server Server, igcFileServer *httptest.Server) {
	// Setup a simple igc-file hosting server
	igcFileServer = makeIgcFileServer()
	igcFileServer.Start()

	// Setup in-memory track metas
	trackMetasMap := NewTrackMetasMap()

	// Setup a dummy ticker
	ticker := NewTickerDummy(2)

	// Setup a in-memory webhooks
	webhooks := NewWebhooksMap()

	// Initialize main API server
	server = NewServer(igcFileServer.Client(), &trackMetasMap, &ticker, &webhooks)
	return
}

// Test GET /
func TestIgcServerGetMetaValid(t *testing.T) {
	// We don't need any extra deps to test metadata
	server := NewServer(nil, nil, nil, nil)

	req := httptest.NewRequest("GET", "/", nil)
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	var data map[string]interface{}
	if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Errorf("received response body: '%s'", res.Body)
		t.Fatalf("failed when trying to decode body as json")
	}
	for _, field := range []string{"uptime", "info", "version"} {
		if data[field] == nil {
			t.Errorf("'%s' was not found in fields ", field)
		}
	}
}

// Test bad POST /track
func TestIgcServerPostTrackBad(t *testing.T) {
	server, fileserver := makeTestServers()
	defer fileserver.Close()

	for _, body := range []string{
		fmt.Sprintf("{\"url\":\"%s\"}", fileserver.URL+"/invalid.igc"),
		fmt.Sprintf("{\"url\":\"%s\"}", fileserver.URL+"/missing.igc"),
		fmt.Sprintf("{\"url\":\"%s\"}", "asfd££@@1££¡@3invalidasULR"),
		"{\"url\":null}",
		"{\"url\":\"  a:b:c:d@a:b:©:1\"}",
		fmt.Sprintf("{\"l\":\"%s\"}", fileserver.URL+"/aa.igc"),
		fmt.Sprintf("\"l\":\"%s\"}", fileserver.URL+"/bb.igc"),
		fmt.Sprintf("{\"l\":\"%s\", asdf asdf}", fileserver.URL+"/cc.igc"),
	} {
		req := httptest.NewRequest("POST", "/track", bytes.NewReader([]byte(body)))
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		code := res.Result().StatusCode
		if code != 400 {
			t.Fatalf("expected '%s' to return 400 (bad request), got '%d'", body, code)
		}
	}
}

// Test valid POST /track
func TestIgcServerPostTrackValid(t *testing.T) {
	server, fileserver := makeTestServers()
	defer fileserver.Close()

	body := fmt.Sprintf("{\"url\":\"%s\"}", fileserver.URL+"/test.igc")
	req := httptest.NewRequest("POST", "/track", bytes.NewReader([]byte(body)))
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	var data map[string]TrackID
	if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Errorf("received response body: '%s'", res.Body)
		t.Fatalf("failed when trying to decode body as json")
	}

	req = httptest.NewRequest("GET", "/track", nil)
	res = httptest.NewRecorder()

	server.ServeHTTP(res, req)

	var respData []TrackID
	if err := json.Unmarshal(res.Body.Bytes(), &respData); err != nil {
		t.Errorf("received response body: '%s'", res.Body)
		t.Fatalf("failed when trying to decode body as json")
	}

	for _, gotID := range respData {
		if gotID == data["id"] {
			return
		}
	}
	t.Fatalf("id of inserted track ('%d') was not found in ids returned from `GET /track` ('%d')", data["id"], respData)
}

// Test valid POST /track
func TestIgcServerPostTrackValidDuplicate(t *testing.T) {
	server, fileserver := makeTestServers()
	defer fileserver.Close()

	body := fmt.Sprintf("{\"url\":\"%s\"}", fileserver.URL+"/test.igc")
	req := httptest.NewRequest("POST", "/track", bytes.NewReader([]byte(body)))
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	var data map[string]TrackID
	if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Errorf("received response body: '%s'", res.Body)
		t.Fatalf("failed when trying to decode body as json")
	}

	req = httptest.NewRequest("POST", "/track", bytes.NewReader([]byte(body)))
	res = httptest.NewRecorder()
	server.ServeHTTP(res, req)

	code := res.Result().StatusCode
	if code != 403 {
		t.Fatalf("expected attempt to register same file twice to result in 403, got '%d'", code)
	}

}

// Test GET /track
func TestIgcServerGetTrack(t *testing.T) {
	trackMetasMap := NewTrackMetasMap()
	server := NewServer(nil, &trackMetasMap, nil, nil)

	testTrackMetas := makeIGCTestData("localhost")
	ids := make([]TrackID, 0, len(testTrackMetas))
	for _, trackMeta := range testTrackMetas {
		err := server.tracks.Append(trackMeta)
		if err != nil {
			t.Errorf("unable to add metadata: %s", err)
			continue
		}
		ids = append(ids, trackMeta.ID)
	}

	req := httptest.NewRequest("GET", "/track", nil)
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	var data []TrackID
	if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
		t.Errorf("received response body: '%s'", res.Body)
		t.Fatalf("failed when trying to decode body as json")
	}

outer:
	for _, exptID := range ids {
		for _, gotID := range data {
			if TrackID(gotID) == exptID {
				continue outer
			}
		}
		t.Errorf("id of inserted track ('%d') was not found in ids returned from `GET /track` ('%d')", exptID, data)
	}
}

// Test valid GET /track/<id>
func TestIgcServerGetTrackByIdValid(t *testing.T) {
	trackMetasMap := NewTrackMetasMap()
	server := NewServer(nil, &trackMetasMap, nil, nil)

	testTrackMetas := makeIGCTestData("localhost")
	ids := make([]TrackID, 0, len(testTrackMetas))
	for _, trackMeta := range testTrackMetas {
		err := server.tracks.Append(trackMeta)
		if err != nil {
			t.Errorf("unable to add metadata: %s", err)
			continue
		}
		ids = append(ids, trackMeta.ID)
	}

	for i, id := range ids {
		uri := fmt.Sprintf("/track/%d", id)
		req := httptest.NewRequest("GET", uri, nil)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		var data TrackMeta
		if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
			t.Errorf("received response body: '%s'", res.Body)
			t.Fatalf("failed when trying to decode body as json")
		}
		expt, _ := json.MarshalIndent(testTrackMetas[i], "", "  ")
		got, _ := json.MarshalIndent(data, "", "  ")
		if !cmp.Equal(expt, got) {
			t.Errorf("returned track was not equal to inserted track:\n\nrequested id: %d\nexpected:\n%s\n\nreturned:\n%s", id, expt, got)
		}
	}
}

// Test bad GET /track/<id>
func TestIgcServerGetTrackByIdBad(t *testing.T) {
	trackMetasMap := NewTrackMetasMap()
	server := NewServer(nil, &trackMetasMap, nil, nil)

	for _, badID := range []struct {
		int
		string
	}{
		{400, "aaaabbbb"},
		{400, "aaaabbbb/asdfa"},
		{400, "bad"},
		{400, "aøaskdljflkasdjfløjsdaf"},
		{400, "12312o3123"},
		{400, "--asdf--"},
		{400, "a"},
		{404, "1232"},
		{404, "99999"},
	} {
		req := httptest.NewRequest("GET", "/track/"+badID.string, nil)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		code := res.Result().StatusCode
		if code != badID.int {
			t.Errorf("expected `GET /track/%s` to return '%d', got '%d'", badID.string, badID.int, code)
		}
	}
}

// Test valid GET /track/<id>/<field>
func TestIgcServerGetTrackFieldValid(t *testing.T) {
	trackMetasMap := NewTrackMetasMap()
	server := NewServer(nil, &trackMetasMap, nil, nil)

	testTrackMetas := makeIGCTestData("localhost")
	ids := make([]TrackID, 0, len(testTrackMetas))
	for _, trackMeta := range testTrackMetas {
		err := server.tracks.Append(trackMeta)
		if err != nil {
			t.Errorf("unable to add metadata: %s", err)
			continue
		}
		ids = append(ids, trackMeta.ID)
	}

	for i, id := range ids {
		// Encode and decode struct to make indexing easier
		exptJSON, _ := json.Marshal(testTrackMetas[i])
		var expt map[string]interface{}
		json.Unmarshal(exptJSON, &expt)

		for _, field := range []string{
			"pilot",
			"glider",
			"glider_id",
			"track_src_url",
		} {
			uri := fmt.Sprintf("/track/%d/%s", id, field)
			req := httptest.NewRequest("GET", uri, nil)
			res := httptest.NewRecorder()

			server.ServeHTTP(res, req)

			got, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("error when reading body: %s", err)
			}
			if string(got) != expt[field] {
				t.Errorf("unexpected field when `GET /track/%d/%s`, got '%s' but expected '%s'", id, field, got, expt[field])
			}

		}
		for _, field := range []string{
			"H_date",
			"track_length",
		} {
			uri := fmt.Sprintf("/track/%d/%s", id, field)
			req := httptest.NewRequest("GET", uri, nil)
			res := httptest.NewRecorder()

			server.ServeHTTP(res, req)

			got, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("error when reading body: %s", err)
			}
			if got == nil {
				t.Errorf("empty field when `GET /track/%d/%s`", id, field)
			}
			if string(got) == "" {
				t.Errorf("empty string when `GET /track/%d/%s`", id, field)
			}
		}
	}
}

// Test bad GET /track/<id>/<field>
func TestIgcServerGetTrackFieldBad(t *testing.T) {
	trackMetasMap := NewTrackMetasMap()
	server := NewServer(nil, &trackMetasMap, nil, nil)

	testTrackMetas := makeIGCTestData("localhost")
	ids := make([]TrackID, 0, len(testTrackMetas))
	for _, trackMeta := range testTrackMetas {
		err := server.tracks.Append(trackMeta)
		if err != nil {
			t.Errorf("unable to add metadata: %s", err)
			continue
		}
		ids = append(ids, trackMeta.ID)
	}

	var unknownID TrackID

	// It is a rare occurrence, but make sure that the unknown id NEVER can be
	// equal to a generated id
outer:
	for {
		unknownID = TrackID(rand.Int())
		for _, id := range ids {
			if unknownID == id {
				continue outer
			}
		}
		break
	}

	for _, data := range []struct {
		code  int
		field string
		id    TrackID
	}{
		{400, "asdlfkjaksl", ids[0]},
		{400, "aasdf90123", ids[0]},
		{400, "12312", ids[1]},
		{400, "--..s.a", ids[1]},
		{404, "asdf", unknownID},
	} {
		uri := fmt.Sprintf("/track/%d/%s", data.id, data.field)
		req := httptest.NewRequest("GET", uri, nil)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		code := res.Result().StatusCode
		if code != data.code {
			t.Fatalf("expected `GET /track/%d/%s` to return '%d', got '%d'", data.id, data.field, data.code, code)
		}
	}
}

// Test different rubbish urls -> 404
func TestIgcServerGetRubbish(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	rubbishURLs := []string{
		"/rubbish",
		"/asdfa",
		"/asdfas/asdfasd/asdfasdf/asdfasdf/",
		"/paragliding/asdfasf",
		"/paragliding/igcasd",
		"/paragliding/api/asdlfasdf",
		"/paragliding/api/rubbish",
		"/paragliding/api/0a90a9ds109123",
		"/paragliding/api/some-path",
		"/paragliding/api/webhook/asdf-path",
		"/paragliding/api/webhook/asdf-paasasdfasdfasdf",
		"/paragliding/api/webhook/new_track/asdfa/asdfa",
		"/paragliding/api/ticker/new_track/asdfa/asdfa",
		"/paragliding/api/ticker/asdfa/asdfa/asdfa",
		"/paragliding/api/ticker/latest/asdfa/asdfa",
		"/012312390123123/api/some-path",
		"/a213asd123/api/some-path",
	}

	for _, rubbishURL := range rubbishURLs {
		req := httptest.NewRequest("GET", rubbishURL, nil)
		res := httptest.NewRecorder()
		server.ServeHTTP(res, req)

		code := res.Result().StatusCode
		if code != 404 {
			t.Fatalf("expected `GET %s` to return a 404, got %d", rubbishURL, code)
		}
	}
}

// Test PUT -> 405 response
func TestIgcServerPutMethod(t *testing.T) {
	server := NewServer(nil, nil, nil, nil)

	req := httptest.NewRequest("PUT", "/", nil)
	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	code := res.Result().StatusCode
	if code != 405 {
		t.Fatalf("expected `PUT /` to return a 405, got '%d'", code)
	}
	allowedMethods := res.Result().Header.Get("Allow")
	if allowedMethods == "" {
		t.Fatalf("expected `PUT /` to return an `Allow` header containing the allowed methods, no methods returned or missing header")
	}
}

// Test bad GET /webhook/new_track/<id>
func TestGetWebhookByBadID(t *testing.T) {
	webhooksMap := NewWebhooksMap()
	server := NewServer(nil, nil, nil, &webhooksMap)

	for _, badID := range []struct {
		int
		string
	}{
		{400, "aaaabbbb"},
		{400, "bad"},
		{400, "aøaskdljflkasdjfløjsdaf"},
		{400, "12312o3123"},
		{400, "--asdf--"},
		{400, "a"},
		{404, "1232"},
		{404, "99999"},
	} {
		req := httptest.NewRequest("GET", "/webhook/new_track/"+badID.string, nil)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		code := res.Result().StatusCode
		if code != badID.int {
			t.Errorf("expected `GET /webhook/new_track/%s` to return '%d', got '%d'", badID.string, badID.int, code)
		}
	}
}

// Test valid GET /webhook/new_track/<id>
func TestGetWebhookByIdValid(t *testing.T) {
	webhooksMap := NewWebhooksMap()
	server := NewServer(nil, nil, nil, &webhooksMap)

	testData := makeWebhooksTestData()
	ids := make([]WebhookID, 0, len(testData))
	for _, webhook := range testData {
		err := server.webhooks.Append(webhook)
		if err != nil {
			t.Errorf("unable to add metadata: %s", err)
			continue
		}
		ids = append(ids, webhook.ID)
	}

	for i, id := range ids {
		uri := fmt.Sprintf("/webhook/new_track/%d", id)
		req := httptest.NewRequest("GET", uri, nil)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		var data WebhookInfo
		if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
			t.Errorf("received response body: '%s'", res.Body)
			t.Fatalf("failed when trying to decode body as json")
		}
		expt, _ := json.MarshalIndent(testData[i], "", "  ")
		got, _ := json.MarshalIndent(data, "", "  ")
		if !cmp.Equal(expt, got) {
			t.Errorf("returned track was not equal to inserted track:\n\nrequested id: %d\nexpected:\n%s\n\nreturned:\n%s", id, expt, got)
		}
	}
}

// Test valid POST /webhook/new_track/
func TestRegWebhook(t *testing.T) {
	webhooksMap := NewWebhooksMap()
	server := NewServer(nil, nil, nil, &webhooksMap)

	testData := makeWebhooksTestData()
	ids := make([]WebhookID, len(testData))
	b := new(bytes.Buffer)
	for i, webhook := range testData {
		json.NewEncoder(b).Encode(&webhook)
		req := httptest.NewRequest("POST", "/webhook/new_track", b)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		if res.Result().StatusCode != 200 {
			t.Fatalf("unable to add webhook through POST request")
		}
		defer res.Result().Body.Close()
		var id int
		var err error
		if id, err = strconv.Atoi(string(res.Body.Bytes())); err != nil {
			t.Fatal("unable to decode response as integer")
		}
		ids[i] = WebhookID(id)
	}

	for i, id := range ids {
		uri := fmt.Sprintf("/webhook/new_track/%d", id)
		req := httptest.NewRequest("GET", uri, nil)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)

		var data WebhookInfo
		if err := json.Unmarshal(res.Body.Bytes(), &data); err != nil {
			t.Errorf("received response body: '%s'", res.Body)
			t.Fatalf("failed when trying to decode body as json")
		}
		expt, _ := json.MarshalIndent(testData[i], "", "  ")
		got, _ := json.MarshalIndent(data, "", "  ")
		if !cmp.Equal(expt, got) {
			t.Errorf("returned track was not equal to inserted track:\n\nrequested id: %d\nexpected:\n%s\n\nreturned:\n%s", id, expt, got)
		}
	}
}

// Test invalid POST /webhook/new_track/
func TestRegWebhookBad(t *testing.T) {
	webhooksMap := NewWebhooksMap()
	server := NewServer(nil, nil, nil, &webhooksMap)

	var data = []struct {
		int
		string
	}{
		{400, "{\"sdfsfs\":\"sadf\"}"},
		{400, "{\"s11sfs\":\"sadf\"}"},
		{400, "\"s11sfs\":\"sadf\"}"},
		{400, "{\"sdfsfs\":\12123}"},
		{400, "{\"webhookURL\":12123}"},
		{400, "{\"webhookURL\":aabb}"},
		{400, "{\"webhookURl\":\"abff}"},
	}

	b := new(bytes.Buffer)
	for _, dat := range data {
		b.Reset()
		b.WriteString(dat.string)
		req := httptest.NewRequest("POST", "/webhook/new_track", b)
		res := httptest.NewRecorder()

		server.ServeHTTP(res, req)
		code := res.Result().StatusCode
		if code != dat.int {
			t.Fatalf("expected `POST /webhook/new_track/` to give '%d' but got '%d'", dat.int, code)
		}
	}
}
