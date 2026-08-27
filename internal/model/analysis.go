package model

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	APIVersion = "malzone.io/v1alpha1"
	Kind       = "Analysis"
)

var dnsLabel = regexp.MustCompile(`^[a-z0-9](?:[-a-z0-9]*[a-z0-9])?$`)

type ObjectMeta struct {
	Name            string `json:"name,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	UID             string `json:"uid,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
	Generation      int64  `json:"generation,omitempty"`
}

type Analysis struct {
	APIVersion string         `json:"apiVersion"`
	Kind       string         `json:"kind"`
	Metadata   ObjectMeta     `json:"metadata"`
	Spec       AnalysisSpec   `json:"spec"`
	Status     AnalysisStatus `json:"status,omitempty"`
}

type AnalysisList struct {
	Items []Analysis `json:"items"`
}

type AnalysisSpec struct {
	Sample          SampleSpec          `json:"sample"`
	TimeoutSeconds  int                 `json:"timeoutSeconds"`
	CancelRequested bool                `json:"cancelRequested,omitempty"`
	Interactions    []InteractionAction `json:"interactions,omitempty"`
}

type SampleSpec struct {
	Kind    string `json:"kind"`
	Content string `json:"content"`
}

type AnalysisStatus struct {
	ObservedGeneration int64               `json:"observedGeneration,omitempty"`
	Phase              string              `json:"phase,omitempty"`
	PendingOutcome     string              `json:"pendingOutcome,omitempty"`
	Message            string              `json:"message,omitempty"`
	RunnerJob          string              `json:"runnerJob,omitempty"`
	StartedAt          string              `json:"startedAt,omitempty"`
	CompletedAt        string              `json:"completedAt,omitempty"`
	CleanupVerified    bool                `json:"cleanupVerified,omitempty"`
	Result             *AnalysisResult     `json:"result,omitempty"`
	InteractionResults []InteractionResult `json:"interactionResults,omitempty"`
}

type InteractionAction struct {
	ID                  string `json:"id"`
	Type                string `json:"type"`
	Rationale           string `json:"rationale"`
	ExpectedObservation string `json:"expectedObservation"`
	CreatedAt           string `json:"createdAt"`
}

type InteractionResult struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	ObservedAt string `json:"observedAt"`
	Detail     string `json:"detail"`
}

type AnalysisResult struct {
	SHA256                    string `json:"sha256,omitempty"`
	Summary                   string `json:"summary,omitempty"`
	ServiceAccountTokenAbsent bool   `json:"serviceAccountTokenAbsent"`
	KubernetesAPIDenied       bool   `json:"kubernetesApiDenied"`
}

type CreateAnalysisRequest struct {
	Name           string     `json:"name,omitempty"`
	Sample         SampleSpec `json:"sample"`
	TimeoutSeconds int        `json:"timeoutSeconds,omitempty"`
}

type CreateInteractionRequest struct {
	Type                string `json:"type"`
	Rationale           string `json:"rationale"`
	ExpectedObservation string `json:"expectedObservation"`
}

func NewAnalysis(namespace string, request CreateAnalysisRequest) Analysis {
	timeout := request.TimeoutSeconds
	if timeout == 0 {
		timeout = 3
	}
	return Analysis{
		APIVersion: APIVersion,
		Kind:       Kind,
		Metadata: ObjectMeta{
			Name:      request.Name,
			Namespace: namespace,
		},
		Spec: AnalysisSpec{
			Sample:         request.Sample,
			TimeoutSeconds: timeout,
		},
	}
}

func ValidateCreate(request CreateAnalysisRequest) error {
	if request.Name != "" {
		if len(request.Name) > 63 || !dnsLabel.MatchString(request.Name) {
			return errors.New("name must be a lowercase DNS label of at most 63 characters")
		}
	}
	if request.Sample.Kind != "canary" {
		return errors.New("POC accepts only sample.kind=canary")
	}
	if strings.TrimSpace(request.Sample.Content) == "" {
		return errors.New("sample.content is required")
	}
	if len(request.Sample.Content) > 256 {
		return errors.New("sample.content must not exceed 256 bytes")
	}
	if request.TimeoutSeconds < 0 || request.TimeoutSeconds > 60 {
		return errors.New("timeoutSeconds must be between 1 and 60 when specified")
	}
	return nil
}

func ValidateInteraction(request CreateInteractionRequest) error {
	if request.Type != "observe" {
		return errors.New("POC accepts only interaction type=observe; shell and desktop input are unavailable")
	}
	if strings.TrimSpace(request.Rationale) == "" || len(request.Rationale) > 256 {
		return errors.New("rationale is required and must not exceed 256 bytes")
	}
	if strings.TrimSpace(request.ExpectedObservation) == "" || len(request.ExpectedObservation) > 256 {
		return errors.New("expectedObservation is required and must not exceed 256 bytes")
	}
	return nil
}

func NewInteraction(id string, request CreateInteractionRequest, now time.Time) InteractionAction {
	return InteractionAction{
		ID:                  id,
		Type:                request.Type,
		Rationale:           strings.TrimSpace(request.Rationale),
		ExpectedObservation: strings.TrimSpace(request.ExpectedObservation),
		CreatedAt:           now.UTC().Format(time.RFC3339),
	}
}

func IsTerminal(phase string) bool {
	switch phase {
	case "Succeeded", "Failed", "Cancelled":
		return true
	default:
		return false
	}
}

func ValidateName(name string) error {
	if len(name) == 0 || len(name) > 63 || !dnsLabel.MatchString(name) {
		return fmt.Errorf("invalid analysis name %q", name)
	}
	return nil
}
