package operator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/diaryfolio/malzone/internal/kube"
	"github.com/diaryfolio/malzone/internal/model"
)

type Client interface {
	ListAnalyses(context.Context) ([]model.Analysis, error)
	GetJob(context.Context, string) (kube.JobState, error)
	CreateJob(context.Context, model.Analysis, string) error
	DeleteJob(context.Context, string) error
	JobLogs(context.Context, string) (string, error)
	UpdateStatus(context.Context, model.Analysis) (model.Analysis, error)
}

type Controller struct {
	client      Client
	runnerImage string
	interval    time.Duration
	logger      *slog.Logger
}

func New(client Client, runnerImage string, interval time.Duration, logger *slog.Logger) *Controller {
	return &Controller{client: client, runnerImage: runnerImage, interval: interval, logger: logger}
}

func (c *Controller) Run(ctx context.Context) error {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		if err := c.reconcileAll(ctx); err != nil && !errors.Is(err, context.Canceled) {
			c.logger.Error("reconcile list failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (c *Controller) reconcileAll(ctx context.Context) error {
	analyses, err := c.client.ListAnalyses(ctx)
	if err != nil {
		return err
	}
	for _, analysis := range analyses {
		if err := c.Reconcile(ctx, analysis); err != nil {
			c.logger.Error("reconcile failed", "analysis", analysis.Metadata.Name, "error", err)
		}
	}
	return nil
}

func (c *Controller) Reconcile(ctx context.Context, analysis model.Analysis) error {
	if model.IsTerminal(analysis.Status.Phase) {
		return nil
	}
	jobName := kube.JobName(analysis.Metadata.Name)
	if analysis.Spec.CancelRequested {
		_, err := c.client.GetJob(ctx, jobName)
		if kube.IsNotFound(err) {
			return c.setTerminal(ctx, analysis, "Cancelled", "Cancellation completed; runner resources absent")
		}
		if err != nil {
			return err
		}
		if err := c.client.DeleteJob(ctx, jobName); err != nil {
			return err
		}
		analysis.Status.ObservedGeneration = analysis.Metadata.Generation
		analysis.Status.Phase = "Cleaning"
		analysis.Status.PendingOutcome = "Cancelled"
		analysis.Status.Message = "Cancellation requested; waiting for runner resources to disappear"
		_, err = c.client.UpdateStatus(ctx, analysis)
		return err
	}

	job, err := c.client.GetJob(ctx, jobName)
	if kube.IsNotFound(err) {
		if analysis.Status.Phase == "Collecting" || analysis.Status.Phase == "Cleaning" {
			outcome := analysis.Status.PendingOutcome
			if outcome == "" {
				outcome = "Failed"
			}
			return c.setTerminal(ctx, analysis, outcome, "Runner resources removed and cleanup verified")
		}
		if err := c.client.CreateJob(ctx, analysis, c.runnerImage); err != nil && !kube.IsConflict(err) {
			return err
		}
		analysis.Status = model.AnalysisStatus{
			ObservedGeneration: analysis.Metadata.Generation,
			Phase:              "Provisioning",
			Message:            "Harmless POC runner Job created",
			RunnerJob:          jobName,
		}
		_, err = c.client.UpdateStatus(ctx, analysis)
		return err
	}
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	switch {
	case job.Succeeded > 0:
		logs, logErr := c.client.JobLogs(ctx, jobName)
		if logErr != nil {
			return fmt.Errorf("collect runner logs: %w", logErr)
		}
		result, parseErr := parseResult(logs)
		if parseErr != nil {
			return fmt.Errorf("parse runner result: %w", parseErr)
		}
		analysis.Status.Phase = "Collecting"
		analysis.Status.PendingOutcome = "Succeeded"
		analysis.Status.Message = "Result captured; deleting runner resources"
		analysis.Status.Result = &result
		if analysis.Status.StartedAt == "" {
			analysis.Status.StartedAt = now
		}
		if _, err := c.client.UpdateStatus(ctx, analysis); err != nil {
			return err
		}
		return c.client.DeleteJob(ctx, jobName)
	case job.Failed > 0:
		analysis.Status.Phase = "Collecting"
		analysis.Status.PendingOutcome = "Failed"
		analysis.Status.Message = "Runner failed; deleting runner resources"
		if _, err := c.client.UpdateStatus(ctx, analysis); err != nil {
			return err
		}
		return c.client.DeleteJob(ctx, jobName)
	case job.Active > 0:
		analysis.Status.Phase = "Running"
		analysis.Status.Message = "Harmless canary runner is active"
		if analysis.Status.StartedAt == "" {
			analysis.Status.StartedAt = now
		}
	default:
		analysis.Status.Phase = "Provisioning"
		analysis.Status.Message = "Waiting for harmless canary runner"
	}
	analysis.Status.ObservedGeneration = analysis.Metadata.Generation
	analysis.Status.RunnerJob = jobName
	_, err = c.client.UpdateStatus(ctx, analysis)
	return err
}

func (c *Controller) setTerminal(ctx context.Context, analysis model.Analysis, phase, message string) error {
	analysis.Status.ObservedGeneration = analysis.Metadata.Generation
	analysis.Status.Phase = phase
	analysis.Status.PendingOutcome = ""
	analysis.Status.Message = message
	analysis.Status.CleanupVerified = true
	analysis.Status.CompletedAt = time.Now().UTC().Format(time.RFC3339)
	_, err := c.client.UpdateStatus(ctx, analysis)
	return err
}

func parseResult(logs string) (model.AnalysisResult, error) {
	for _, line := range strings.Split(logs, "\n") {
		var envelope struct {
			Type                      string `json:"type"`
			SHA256                    string `json:"sha256"`
			Summary                   string `json:"summary"`
			ServiceAccountTokenAbsent bool   `json:"serviceAccountTokenAbsent"`
			KubernetesAPIDenied       bool   `json:"kubernetesApiDenied"`
		}
		if json.Unmarshal([]byte(line), &envelope) == nil && envelope.Type == "result" {
			return model.AnalysisResult{
				SHA256:                    envelope.SHA256,
				Summary:                   envelope.Summary,
				ServiceAccountTokenAbsent: envelope.ServiceAccountTokenAbsent,
				KubernetesAPIDenied:       envelope.KubernetesAPIDenied,
			}, nil
		}
	}
	return model.AnalysisResult{}, errors.New("structured result line not found")
}
