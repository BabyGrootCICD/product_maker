package allowfailure

import (
	"testing"
)

func TestParsePolicy(t *testing.T) {
	tests := []struct {
		input    string
		expected FailurePolicy
		wantErr  bool
	}{
		{"block", PolicyBlock, false},
		{"ignore", PolicyIgnore, false},
		{"warn", PolicyWarn, false},
		{"report", PolicyReport, false},
		{"", PolicyBlock, false},
		{"BLOCK", PolicyBlock, false},
		{"Ignore", PolicyIgnore, false},
		{"invalid", PolicyBlock, true},
	}
	for _, tt := range tests {
		got, err := ParsePolicy(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParsePolicy(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
		}
		if got != tt.expected {
			t.Errorf("ParsePolicy(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestFailureHandlerSuccess(t *testing.T) {
	handler := NewFailureHandler()
	handler.RegisterJob(&JobConfig{
		Name:         "test-job",
		AllowFailure: true,
	})

	result := handler.HandleFailure("test-job", true, "")
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Allowed {
		t.Fatal("should not be allowed on success")
	}
}

func TestFailureHandlerAllowFailure(t *testing.T) {
	handler := NewFailureHandler()
	handler.RegisterJob(&JobConfig{
		Name:         "test-job",
		AllowFailure: true,
	})

	result := handler.HandleFailure("test-job", false, "test error")
	if result.Success {
		t.Fatal("expected failure")
	}
	if !result.Allowed {
		t.Fatal("should be allowed")
	}
	if result.ShouldBlock() {
		t.Fatal("should not block")
	}
}

func TestFailureHandlerBlock(t *testing.T) {
	handler := NewFailureHandler()
	handler.RegisterJob(&JobConfig{
		Name:         "test-job",
		AllowFailure: false,
	})

	result := handler.HandleFailure("test-job", false, "test error")
	if result.Success {
		t.Fatal("expected failure")
	}
	if result.Allowed {
		t.Fatal("should not be allowed")
	}
	if !result.ShouldBlock() {
		t.Fatal("should block")
	}
}

func TestFailureHandlerPolicyIgnore(t *testing.T) {
	handler := NewFailureHandler()
	handler.RegisterJob(&JobConfig{
		Name:          "test-job",
		FailurePolicy: PolicyIgnore,
	})

	result := handler.HandleFailure("test-job", false, "test error")
	if result.ShouldBlock() {
		t.Fatal("should not block with ignore policy")
	}
}

func TestFailureHandlerUnknownJob(t *testing.T) {
	handler := NewFailureHandler()

	result := handler.HandleFailure("unknown-job", false, "test error")
	if result.Allowed {
		t.Fatal("unknown job should not be allowed")
	}
	if !result.ShouldBlock() {
		t.Fatal("unknown job should block")
	}
}

func TestWorkflowEvaluator(t *testing.T) {
	handler := NewFailureHandler()
	handler.RegisterJob(&JobConfig{Name: "job1", AllowFailure: true})
	handler.RegisterJob(&JobConfig{Name: "job2", AllowFailure: false})

	evaluator := NewWorkflowEvaluator(handler)
	results := []JobResult{
		handler.HandleFailure("job1", false, "error"),
		handler.HandleFailure("job2", true, ""),
	}

	blocked, blockedJobs := evaluator.EvaluateWorkflow(results)
	if !blocked {
		t.Fatal("expected workflow to be blocked")
	}
	if len(blockedJobs) != 0 {
		t.Fatalf("expected no blocked jobs, got %v", blockedJobs)
	}
}

func TestFormatReport(t *testing.T) {
	results := []JobResult{
		{JobName: "job1", Success: true, Policy: PolicyBlock},
		{JobName: "job2", Success: false, Allowed: true, Policy: PolicyIgnore},
	}
	report := FormatReport(results)
	if report == "" {
		t.Fatal("expected non-empty report")
	}
}

func TestJobResultSummary(t *testing.T) {
	r := JobResult{JobName: "test", Success: true}
	if r.Summary() == "" {
		t.Fatal("expected non-empty summary")
	}

	r.Success = false
	r.Allowed = true
	r.Policy = PolicyIgnore
	if r.Summary() == "" {
		t.Fatal("expected non-empty summary")
	}
}

func BenchmarkHandleFailure(b *testing.B) {
	handler := NewFailureHandler()
	handler.RegisterJob(&JobConfig{Name: "bench-job", AllowFailure: true})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.HandleFailure("bench-job", false, "error")
	}
}
