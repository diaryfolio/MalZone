package siem

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diaryfolio/malzone/internal/model"
)

func terminalAnalysis() model.Analysis {
	return model.Analysis{
		Metadata: model.ObjectMeta{Name: "analysis-one", Namespace: "malzone-system", UID: "uid-one"},
		Spec: model.AnalysisSpec{
			Sample:       model.SampleSpec{Kind: "canary", Content: "must-not-export"},
			Interactions: []model.InteractionAction{{ID: "action-one", Type: "observe", Rationale: "private rationale"}},
		},
		Status: model.AnalysisStatus{
			Phase: "Succeeded", CompletedAt: "2026-08-27T20:00:00Z", CleanupVerified: true,
			Result:             &model.AnalysisResult{SHA256: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Summary: "must-not-export"},
			InteractionResults: []model.InteractionResult{{ID: "action-one", Status: "succeeded", Detail: "must-not-export"}},
		},
	}
}

func TestBuildECSEventIsDeterministicAndMetadataOnly(t *testing.T) {
	t.Parallel()
	analysis := terminalAnalysis()
	first := BuildECSEvent(analysis)
	second := BuildECSEvent(analysis)
	if first.Event.ID != second.Event.ID || first.Event.Outcome != "success" {
		t.Fatalf("unexpected event: %#v", first)
	}
	encoded, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"must-not-export", "private rationale"} {
		if bytes.Contains(encoded, []byte(forbidden)) {
			t.Fatalf("event disclosed %q: %s", forbidden, encoded)
		}
	}
}

func TestSinkDeduplicatesByEventID(t *testing.T) {
	t.Parallel()
	handler := NewSink(2)
	event := BuildECSEvent(terminalAnalysis())
	body, _ := json.Marshal(event)
	for range 2 {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body)))
		if response.Code != http.StatusAccepted {
			t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
		}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/events", nil))
	var listing struct {
		Items []ECSEvent `json:"items"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &listing); err != nil {
		t.Fatal(err)
	}
	if len(listing.Items) != 1 {
		t.Fatalf("items=%d", len(listing.Items))
	}
}

func TestAdapterRejectsCredentialedOrUnexpectedEndpoint(t *testing.T) {
	t.Parallel()
	for _, endpoint := range []string{"file:///tmp/events", "http://user:pass@example/events", "http://example/other", "http://example/events?token=secret"} {
		if _, err := NewAdapter(nil, endpoint, 1, nil); err == nil {
			t.Fatalf("expected endpoint %q to be rejected", endpoint)
		}
	}
}
