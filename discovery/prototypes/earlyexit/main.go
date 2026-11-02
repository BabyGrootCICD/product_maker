// Package earlyexit implements early-exit functionality for GitHub Actions jobs.
// It provides an exit handler that can set custom check conclusions when
// exiting a job early, similar to GitLab CI's when:manual behavior.
package earlyexit

import (
	"fmt"
	"os"
	"strings"
)

// Conclusion represents a GitHub Actions check conclusion.
type Conclusion string

const (
	// ConclusionSuccess indicates successful completion.
	ConclusionSuccess Conclusion = "success"
	// ConclusionFailure indicates failed completion.
	ConclusionFailure Conclusion = "failure"
	// ConclusionNeutral indicates neutral completion.
	ConclusionNeutral Conclusion = "neutral"
	// ConclusionCancelled indicates cancelled completion.
	ConclusionCancelled Conclusion = "cancelled"
	// ConclusionSkipped indicates skipped completion.
	ConclusionSkipped Conclusion = "skipped"
	// ConclusionTimedOut indicates timed out completion.
	ConclusionTimedOut Conclusion = "timed_out"
)

// String returns the string representation of the conclusion.
func (c Conclusion) String() string {
	return string(c)
}

// IsValid checks if the conclusion is valid.
func (c Conclusion) IsValid() bool {
	switch c {
	case ConclusionSuccess, ConclusionFailure, ConclusionNeutral,
		ConclusionCancelled, ConclusionSkipped, ConclusionTimedOut:
		return true
	default:
		return false
	}
}

// ExitInfo contains information about an early exit.
type ExitInfo struct {
	Conclusion Conclusion
	Message    string
	Code       int
}

// ExitHandler processes early exit commands.
type ExitHandler struct {
	defaultConclusion Conclusion
}

// NewExitHandler creates a new ExitHandler.
func NewExitHandler(defaultConclusion Conclusion) *ExitHandler {
	if !defaultConclusion.IsValid() {
		defaultConclusion = ConclusionSuccess
	}
	return &ExitHandler{
		defaultConclusion: defaultConclusion,
	}
}

// HandleEarlyExit processes an early exit with the given conclusion.
func (h *ExitHandler) HandleEarlyExit(conclusion Conclusion, message string) ExitInfo {
	if !conclusion.IsValid() {
		conclusion = h.defaultConclusion
	}

	code := conclusionToCode(conclusion)

	return ExitInfo{
		Conclusion: conclusion,
		Message:    message,
		Code:       code,
	}
}

// HandleFromEnv reads conclusion from GITHUB_CONCLUSION environment variable.
func (h *ExitHandler) HandleFromEnv(message string) ExitInfo {
	conclusion := Conclusion(os.Getenv("GITHUB_CONCLUSION"))
	if !conclusion.IsValid() {
		conclusion = h.defaultConclusion
	}
	return h.HandleEarlyExit(conclusion, message)
}

// ParseFlags parses command-line flags for early-exit command.
func ParseFlags(args []string) (Conclusion, string, error) {
	conclusion := ConclusionSuccess
	message := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--conclusion", "-c":
			if i+1 < len(args) {
				i++
				conclusion = Conclusion(args[i])
				if !conclusion.IsValid() {
					return "", "", fmt.Errorf("invalid conclusion: %s", args[i])
				}
			} else {
				return "", "", fmt.Errorf("--conclusion requires a value")
			}
		case "--message", "-m":
			if i+1 < len(args) {
				i++
				message = args[i]
			} else {
				return "", "", fmt.Errorf("--message requires a value")
			}
		default:
			return "", "", fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	return conclusion, message, nil
}

// conclusionToCode maps a conclusion to an exit code.
func conclusionToCode(c Conclusion) int {
	switch c {
	case ConclusionSuccess:
		return 0
	case ConclusionFailure:
		return 1
	case ConclusionNeutral:
		return 2
	case ConclusionCancelled:
		return 3
	case ConclusionSkipped:
		return 4
	case ConclusionTimedOut:
		return 5
	default:
		return 0
	}
}

// CodeToConclusion maps an exit code to a conclusion.
func CodeToConclusion(code int) Conclusion {
	switch code {
	case 0:
		return ConclusionSuccess
	case 1:
		return ConclusionFailure
	case 2:
		return ConclusionNeutral
	case 3:
		return ConclusionCancelled
	case 4:
		return ConclusionSkipped
	case 5:
		return ConclusionTimedOut
	default:
		return ConclusionFailure
	}
}

// FormatReport generates a formatted report of early exit information.
func FormatReport(info ExitInfo) string {
	var sb strings.Builder
	sb.WriteString("## Early Exit Report\n\n")
	sb.WriteString(fmt.Sprintf("- **Conclusion**: %s\n", info.Conclusion))
	sb.WriteString(fmt.Sprintf("- **Message**: %s\n", info.Message))
	sb.WriteString(fmt.Sprintf("- **Exit Code**: %d\n", info.Code))
	return sb.String()
}
