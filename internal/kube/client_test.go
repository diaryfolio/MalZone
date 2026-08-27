package kube

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/diaryfolio/malzone/internal/model"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestAppendInteractionRechecksInteractiveStateBeforePatch(t *testing.T) {
	t.Parallel()
	patches := 0
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method == http.MethodPatch {
			patches++
		}
		body, _ := json.Marshal(model.Analysis{
			Metadata: model.ObjectMeta{Name: "analysis-one", ResourceVersion: "2"},
			Status:   model.AnalysisStatus{Phase: "Succeeded"},
		})
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(body)),
			Request:    request,
		}, nil
	})

	client := NewForTest("https://kubernetes.invalid", "malzone-system", &http.Client{Transport: transport})
	_, err := client.AppendInteraction(context.Background(), "analysis-one", model.InteractionAction{ID: "action-one", Type: "observe"})
	if !IsConflict(err) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if patches != 0 {
		t.Fatalf("unexpected patches: %d", patches)
	}
}
