package earlyexit

import (
	"testing"
)

func TestConclusionIsValid(t *testing.T) {
	valid := []Conclusion{ConclusionSuccess, ConclusionFailure, ConclusionNeutral, ConclusionCancelled, ConclusionSkipped, ConclusionTimedOut}
	for _, c := range valid {
		if !c.IsValid() {
			t.Errorf("expected %s to be valid", c)
		}
	}
	invalid := []Conclusion{"invalid", "unknown", ""}
	for _, c := range invalid {
		if c.IsValid() {
			t.Errorf("expected %s to be invalid", c)
		}
	}
}

func TestHandleEarlyExit(t *testing.T) {
	handler := NewExitHandler(ConclusionSuccess)
	result := handler.HandleEarlyExit(ConclusionFailure, "test message")
	if result.Conclusion != ConclusionFailure {
		t.Fatalf("expected failure, got %s", result.Conclusion)
	}
	if result.Message != "test message" {
		t.Fatalf("expected 'test message', got %s", result.Message)
	}
	if result.Code != 1 {
		t.Fatalf("expected exit code 1, got %d", result.Code)
	}
}

func TestHandleEarlyExitInvalid(t *testing.T) {
	handler := NewExitHandler(ConclusionSuccess)
	result := handler.HandleEarlyExit("invalid", "test")
	if result.Conclusion != ConclusionSuccess {
		t.Fatalf("expected success (default), got %s", result.Conclusion)
	}
}

func TestHandleFromEnv(t *testing.T) {
	handler := NewExitHandler(ConclusionSuccess)
	result := handler.HandleFromEnv("env test")
	if result.Conclusion != ConclusionSuccess {
		t.Fatalf("expected success (default), got %s", result.Conclusion)
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		args    []string
		concl   Conclusion
		message string
		wantErr bool
	}{
		{[]string{"--conclusion", "success"}, ConclusionSuccess, "", false},
		{[]string{"-c", "failure", "-m", "msg"}, ConclusionFailure, "msg", false},
		{[]string{"--conclusion", "invalid"}, "", "", true},
		{[]string{"--unknown"}, "", "", true},
	}
	for _, tt := range tests {
		concl, msg, err := ParseFlags(tt.args)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseFlags(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
		}
		if !tt.wantErr && concl != tt.concl {
			t.Errorf("ParseFlags(%v) conclusion = %v, want %v", tt.args, concl, tt.concl)
		}
		if !tt.wantErr && msg != tt.message {
			t.Errorf("ParseFlags(%v) message = %v, want %v", tt.args, msg, tt.message)
		}
	}
}

func TestConclusionToCode(t *testing.T) {
	tests := []struct {
		concl Conclusion
		code  int
	}{
		{ConclusionSuccess, 0},
		{ConclusionFailure, 1},
		{ConclusionNeutral, 2},
		{ConclusionCancelled, 3},
		{ConclusionSkipped, 4},
		{ConclusionTimedOut, 5},
	}
	for _, tt := range tests {
		code := conclusionToCode(tt.concl)
		if code != tt.code {
			t.Errorf("conclusionToCode(%s) = %d, want %d", tt.concl, code, tt.code)
		}
	}
}

func TestCodeToConclusion(t *testing.T) {
	tests := []struct {
		code  int
		concl Conclusion
	}{
		{0, ConclusionSuccess},
		{1, ConclusionFailure},
		{2, ConclusionNeutral},
		{3, ConclusionCancelled},
		{4, ConclusionSkipped},
		{5, ConclusionTimedOut},
		{99, ConclusionFailure},
	}
	for _, tt := range tests {
		concl := CodeToConclusion(tt.code)
		if concl != tt.concl {
			t.Errorf("CodeToConclusion(%d) = %s, want %s", tt.code, concl, tt.concl)
		}
	}
}

func TestFormatReport(t *testing.T) {
	info := ExitInfo{
		Conclusion: ConclusionSuccess,
		Message:    "test",
		Code:       0,
	}
	report := FormatReport(info)
	if report == "" {
		t.Fatal("expected non-empty report")
	}
}

func BenchmarkHandleEarlyExit(b *testing.B) {
	handler := NewExitHandler(ConclusionSuccess)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.HandleEarlyExit(ConclusionSuccess, "bench")
	}
}
