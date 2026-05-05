package ingest

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

// fakeHandler implements the same routing as Handler but records calls instead
// of hitting BQ.
type fakeHandler struct {
	resultCalls int
	statusCalls int
	auditCalls  int
	failNext    bool
}

func (f *fakeHandler) InsertResult(_ interface{}, msg PubSubMessage, _ interface{}) error {
	f.resultCalls++
	return nil
}

// We test the server routing by wiring a real Server with a nil Handler and
// checking that route dispatch works via the subscription field. Since we can't
// inject a fake Handler (it's a concrete type), we test via a thin HTTP-level
// integration: build a PubSub push envelope, POST it, and check the status code.
// The handler will fail when it tries to call BQ (nil client), so we only test
// the 400 path (bad JSON) and the routing/method paths that don't reach BQ.

func newTestServer() (*Server, *httptest.Server) {
	log := logrus.New()
	log.SetOutput(bytes.NewBuffer(nil)) // silence

	// We can't construct a real Handler without a BQ project, so test the
	// Server's HTTP layer (method guard, bad JSON, unknown subscription) only.
	srv := &Server{
		handler:   nil, // not called in the paths we test
		log:       log,
		resultSub: "fleet-result-logs-sub",
		statusSub: "fleet-status-logs-sub",
		auditSub:  "fleet-audit-logs-sub",
	}
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)
	return srv, httptest.NewServer(mux)
}

func TestServer_MethodGuard(t *testing.T) {
	_, ts := newTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ingest")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("want 405, got %d", resp.StatusCode)
	}
}

func TestServer_BadJSON(t *testing.T) {
	_, ts := newTestServer()
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/ingest", "application/json", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("want 400, got %d", resp.StatusCode)
	}
}

func TestServer_UnknownSubscription(t *testing.T) {
	_, ts := newTestServer()
	defer ts.Close()

	msg := map[string]interface{}{
		"subscription": "projects/p/subscriptions/unknown-sub",
		"message": map[string]interface{}{
			"data":       base64.StdEncoding.EncodeToString([]byte(`{}`)),
			"messageId":  "1",
			"attributes": map[string]string{},
		},
	}
	body, _ := json.Marshal(msg)
	resp, err := http.Post(ts.URL+"/ingest", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	// Unknown subscription returns 200 to prevent PubSub retry loops.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200 for unknown sub, got %d", resp.StatusCode)
	}
}

func TestServer_Healthz(t *testing.T) {
	_, ts := newTestServer()
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("want 200, got %d", resp.StatusCode)
	}
}
