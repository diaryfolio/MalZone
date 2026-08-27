package operator

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/diaryfolio/malzone/internal/kube"
	"github.com/diaryfolio/malzone/internal/model"
)

type fakeClient struct {
	job          kube.JobState
	jobErr       error
	created      bool
	deleted      bool
	logs         string
	status       model.AnalysisStatus
	statusWrites int
}

func (f *fakeClient) ListAnalyses(context.Context) ([]model.Analysis, error) { return nil, nil }
func (f *fakeClient) GetJob(context.Context, string) (kube.JobState, error) {
	return f.job, f.jobErr
}
func (f *fakeClient) CreateJob(context.Context, model.Analysis, string) error {
	f.created = true
	return nil
}
func (f *fakeClient) DeleteJob(context.Context, string) error {
	f.deleted = true
	return nil
}
func (f *fakeClient) JobLogs(context.Context, string) (string, error) { return f.logs, nil }
func (f *fakeClient) UpdateStatus(_ context.Context, analysis model.Analysis) (model.Analysis, error) {
	f.status = analysis.Status
	f.statusWrites++
	return analysis, nil
}

func testAnalysis() model.Analysis {
	return model.Analysis{
		Metadata: model.ObjectMeta{Name: "analysis-test", Generation: 1},
		Spec: model.AnalysisSpec{
			Sample:         model.SampleSpec{Kind: "canary", Content: "hello"},
			TimeoutSeconds: 1,
		},
	}
}

func testController(client Client) *Controller {
	return New(client, "malzone-poc:test", time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestReconcileCreatesRunner(t *testing.T) {
	t.Parallel()
	client := &fakeClient{jobErr: &kube.Error{StatusCode: http.StatusNotFound}}
	if err := testController(client).Reconcile(context.Background(), testAnalysis()); err != nil {
		t.Fatal(err)
	}
	if !client.created || client.status.Phase != "Provisioning" {
		t.Fatalf("created=%v status=%#v", client.created, client.status)
	}
}

func TestReconcileCollectsThenDeletesRunner(t *testing.T) {
	t.Parallel()
	client := &fakeClient{
		job: kube.JobState{Succeeded: 1},
		logs: `{"type":"runner.started"}
{"type":"result","sha256":"abc","summary":"safe","serviceAccountTokenAbsent":true,"kubernetesApiDenied":true}`,
	}
	if err := testController(client).Reconcile(context.Background(), testAnalysis()); err != nil {
		t.Fatal(err)
	}
	if !client.deleted || client.status.Phase != "Collecting" || client.status.Result == nil {
		t.Fatalf("deleted=%v status=%#v", client.deleted, client.status)
	}
	if !client.status.Result.ServiceAccountTokenAbsent || !client.status.Result.KubernetesAPIDenied {
		t.Fatalf("negative checks missing: %#v", client.status.Result)
	}
}

func TestReconcilePublishesTerminalOnlyAfterJobIsGone(t *testing.T) {
	t.Parallel()
	client := &fakeClient{jobErr: &kube.Error{StatusCode: http.StatusNotFound}}
	analysis := testAnalysis()
	analysis.Status.Phase = "Collecting"
	analysis.Status.PendingOutcome = "Succeeded"
	if err := testController(client).Reconcile(context.Background(), analysis); err != nil {
		t.Fatal(err)
	}
	if client.status.Phase != "Succeeded" || !client.status.CleanupVerified {
		t.Fatalf("status=%#v", client.status)
	}
}

func TestCancellationWaitsForDeletionBeforeTerminalStatus(t *testing.T) {
	t.Parallel()
	client := &fakeClient{job: kube.JobState{Active: 1}}
	analysis := testAnalysis()
	analysis.Spec.CancelRequested = true
	if err := testController(client).Reconcile(context.Background(), analysis); err != nil {
		t.Fatal(err)
	}
	if !client.deleted || client.status.Phase != "Cleaning" || client.status.CleanupVerified {
		t.Fatalf("deleted=%v status=%#v", client.deleted, client.status)
	}
	client.jobErr = &kube.Error{StatusCode: http.StatusNotFound}
	analysis.Status = client.status
	if err := testController(client).Reconcile(context.Background(), analysis); err != nil {
		t.Fatal(err)
	}
	if client.status.Phase != "Cancelled" || !client.status.CleanupVerified {
		t.Fatalf("status=%#v", client.status)
	}
}
