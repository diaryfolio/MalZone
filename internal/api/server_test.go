package api

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diaryfolio/malzone/internal/kube"
	"github.com/diaryfolio/malzone/internal/model"
)

type fakeStore struct {
	items []model.Analysis
}

func (f *fakeStore) ListAnalyses(context.Context) ([]model.Analysis, error) {
	return f.items, nil
}

func (f *fakeStore) GetAnalysis(_ context.Context, name string) (model.Analysis, error) {
	for _, item := range f.items {
		if item.Metadata.Name == name {
			return item, nil
		}
	}
	return model.Analysis{}, &kube.Error{StatusCode: http.StatusNotFound}
}

func (f *fakeStore) CreateAnalysis(_ context.Context, analysis model.Analysis) (model.Analysis, error) {
	f.items = append(f.items, analysis)
	return analysis, nil
}

func (f *fakeStore) RequestCancel(ctx context.Context, name string) (model.Analysis, error) {
	analysis, err := f.GetAnalysis(ctx, name)
	if err != nil {
		return model.Analysis{}, err
	}
	analysis.Spec.CancelRequested = true
	return analysis, nil
}

func TestCreateAnalysis(t *testing.T) {
	t.Parallel()
	store := &fakeStore{}
	handler := New(store, "malzone-system", slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodPost, "/api/v1alpha1/analyses",
		bytes.NewBufferString(`{"name":"analysis-one","sample":{"kind":"canary","content":"hello"},"timeoutSeconds":1}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if len(store.items) != 1 || store.items[0].Spec.Sample.Content != "hello" {
		t.Fatalf("unexpected created analysis: %#v", store.items)
	}
	if got := response.Header().Get("Location"); got != "/api/v1alpha1/analyses/analysis-one" {
		t.Fatalf("Location = %q", got)
	}
}

func TestCreateRejectsArbitraryCommand(t *testing.T) {
	t.Parallel()
	handler := New(&fakeStore{}, "malzone-system", slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodPost, "/api/v1alpha1/analyses",
		bytes.NewBufferString(`{"sample":{"kind":"command","content":"curl example.invalid"}}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestSecurityHeaders(t *testing.T) {
	t.Parallel()
	handler := New(&fakeStore{}, "malzone-system", slog.New(slog.NewTextHandler(io.Discard, nil)))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing nosniff header")
	}
}
