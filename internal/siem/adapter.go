package siem

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/diaryfolio/malzone/internal/model"
)

const ECSVersion = "9.5.0"

type AnalysisSource interface {
	ListAnalyses(context.Context) ([]model.Analysis, error)
}

type ECSEvent struct {
	Timestamp string           `json:"@timestamp"`
	ECS       ECSMetadata      `json:"ecs"`
	Event     EventMetadata    `json:"event"`
	Observer  ObserverMetadata `json:"observer"`
	File      *FileMetadata    `json:"file,omitempty"`
	MalZone   MalZoneMetadata  `json:"malzone"`
}

type ECSMetadata struct {
	Version string `json:"version"`
}

type EventMetadata struct {
	ID       string   `json:"id"`
	Kind     string   `json:"kind"`
	Category []string `json:"category"`
	Type     []string `json:"type"`
	Action   string   `json:"action"`
	Outcome  string   `json:"outcome"`
	Provider string   `json:"provider"`
}

type ObserverMetadata struct {
	Vendor  string `json:"vendor"`
	Product string `json:"product"`
	Type    string `json:"type"`
}

type FileMetadata struct {
	Hash FileHash `json:"hash"`
}

type FileHash struct {
	SHA256 string `json:"sha256"`
}

type MalZoneMetadata struct {
	SchemaVersion string                  `json:"schema_version"`
	Analysis      MalZoneAnalysisMetadata `json:"analysis"`
}

type MalZoneAnalysisMetadata struct {
	ID                 string `json:"id"`
	Namespace          string `json:"namespace"`
	Phase              string `json:"phase"`
	CleanupVerified    bool   `json:"cleanup_verified"`
	InteractionCount   int    `json:"interaction_count"`
	ObservationResults int    `json:"observation_result_count"`
}

type Adapter struct {
	source   AnalysisSource
	endpoint *url.URL
	http     *http.Client
	interval time.Duration
	logger   *slog.Logger
	sent     map[string]struct{}
}

func NewAdapter(source AnalysisSource, endpoint string, interval time.Duration, logger *slog.Logger) (*Adapter, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, errors.New("SIEM endpoint must be an absolute HTTP(S) URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || parsed.Path != "/events" {
		return nil, errors.New("SIEM endpoint must have path /events and no credentials, query, or fragment")
	}
	if interval <= 0 {
		return nil, errors.New("adapter interval must be positive")
	}
	return &Adapter{
		source: source, endpoint: parsed, interval: interval, logger: logger,
		http: &http.Client{Timeout: 5 * time.Second}, sent: make(map[string]struct{}),
	}, nil
}

func (a *Adapter) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()
	for {
		if err := a.export(ctx); err != nil && !errors.Is(err, context.Canceled) {
			a.logger.Error("SIEM POC export failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (a *Adapter) export(ctx context.Context) error {
	analyses, err := a.source.ListAnalyses(ctx)
	if err != nil {
		return fmt.Errorf("list analyses: %w", err)
	}
	for _, analysis := range analyses {
		if !model.IsTerminal(analysis.Status.Phase) || !analysis.Status.CleanupVerified || analysis.Status.CompletedAt == "" {
			continue
		}
		event := BuildECSEvent(analysis)
		if _, ok := a.sent[event.Event.ID]; ok {
			continue
		}
		if err := a.deliver(ctx, event); err != nil {
			return fmt.Errorf("deliver analysis %s: %w", analysis.Metadata.Name, err)
		}
		a.sent[event.Event.ID] = struct{}{}
		a.logger.Info("SIEM POC lifecycle metadata delivered", "analysis", analysis.Metadata.Name, "event_id", event.Event.ID)
	}
	return nil
}

func (a *Adapter) deliver(ctx context.Context, event ECSEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "malzone-poc-ecs-adapter/0.1")
	response, err := a.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 1024))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("sink returned HTTP %d", response.StatusCode)
	}
	return nil
}

func BuildECSEvent(analysis model.Analysis) ECSEvent {
	seed := strings.Join([]string{
		analysis.Metadata.Namespace, analysis.Metadata.Name, analysis.Metadata.UID,
		analysis.Status.Phase, analysis.Status.CompletedAt,
	}, "\x00")
	digest := sha256.Sum256([]byte(seed))
	outcome := "failure"
	if analysis.Status.Phase == "Succeeded" {
		outcome = "success"
	} else if analysis.Status.Phase == "Cancelled" {
		outcome = "unknown"
	}
	event := ECSEvent{
		Timestamp: analysis.Status.CompletedAt,
		ECS:       ECSMetadata{Version: ECSVersion},
		Event: EventMetadata{
			ID: hex.EncodeToString(digest[:]), Kind: "event", Category: []string{"malware"},
			Type: []string{"info"}, Action: "analysis-completed", Outcome: outcome, Provider: "malzone",
		},
		Observer: ObserverMetadata{Vendor: "MalZone", Product: "MalZone", Type: "sandbox"},
		MalZone: MalZoneMetadata{
			SchemaVersion: "malzone.ecs.poc/v1alpha1",
			Analysis: MalZoneAnalysisMetadata{
				ID: analysis.Metadata.Name, Namespace: analysis.Metadata.Namespace,
				Phase: analysis.Status.Phase, CleanupVerified: analysis.Status.CleanupVerified,
				InteractionCount:   len(analysis.Spec.Interactions),
				ObservationResults: len(analysis.Status.InteractionResults),
			},
		},
	}
	if analysis.Status.Result != nil && analysis.Status.Result.SHA256 != "" {
		event.File = &FileMetadata{Hash: FileHash{SHA256: analysis.Status.Result.SHA256}}
	}
	return event
}
