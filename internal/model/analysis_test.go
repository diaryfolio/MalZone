package model

import "testing"

func TestValidateCreateRejectsExecutableInputs(t *testing.T) {
	t.Parallel()
	tests := []CreateAnalysisRequest{
		{Sample: SampleSpec{Kind: "url", Content: "https://example.invalid"}},
		{Sample: SampleSpec{Kind: "command", Content: "whoami"}},
		{Sample: SampleSpec{Kind: "canary", Content: ""}},
		{Sample: SampleSpec{Kind: "canary", Content: "ok"}, TimeoutSeconds: 61},
		{Name: "Not-A-DNS-Label", Sample: SampleSpec{Kind: "canary", Content: "ok"}},
	}
	for _, test := range tests {
		if err := ValidateCreate(test); err == nil {
			t.Fatalf("expected request to be rejected: %#v", test)
		}
	}
}

func TestNewAnalysisAppliesSafeDefault(t *testing.T) {
	t.Parallel()
	analysis := NewAnalysis("malzone-system", CreateAnalysisRequest{
		Name:   "analysis-test",
		Sample: SampleSpec{Kind: "canary", Content: "hello"},
	})
	if analysis.Spec.TimeoutSeconds != 3 {
		t.Fatalf("timeout = %d, want 3", analysis.Spec.TimeoutSeconds)
	}
	if analysis.Metadata.Namespace != "malzone-system" {
		t.Fatalf("namespace = %q", analysis.Metadata.Namespace)
	}
}

func TestValidateInteractionIsObserveOnly(t *testing.T) {
	t.Parallel()
	valid := CreateInteractionRequest{Type: "observe", Rationale: "check state", ExpectedObservation: "runner active"}
	if err := ValidateInteraction(valid); err != nil {
		t.Fatal(err)
	}
	for _, actionType := range []string{"shell", "click", "type_text", "launch"} {
		request := valid
		request.Type = actionType
		if err := ValidateInteraction(request); err == nil {
			t.Fatalf("expected %q to be rejected", actionType)
		}
	}
}
