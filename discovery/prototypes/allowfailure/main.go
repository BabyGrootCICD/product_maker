// Package allowfailure implements allow-failure functionality for GitHub Actions jobs.
// It provides a failure handler that can allow specific jobs to fail without
// blocking the entire workflow, similar to GitLab CI's allow_failure: true.
package allowfailure

import (
	"fmt"
	"strings"
)

// FailurePolicy defines how a job failure should be handled.
type FailurePolicy int

const (
	// PolicyBlock blocks the workflow on failure (default).
	PolicyBlock FailurePolicy = iota
	// PolicyIgnore ignores the failure and continues.
	PolicyIgnore
	// PolicyWarn logs a warning but continues.
	PolicyWarn
	// PolicyReport reports the failure but doesn't block.
	PolicyReport
)

// String returns the string representation of the policy.
func (p FailurePolicy) String() string {
	switch p {
	case PolicyBlock:
		return "block"
	case PolicyIgnore:
		return "ignore"
	case PolicyWarn:
		return "warn"
	case PolicyReport:
		return "report"
	default:
		return "unknown"
	}
}

// ParsePolicy parses a string into a FailurePolicy.
func ParsePolicy(s string) (FailurePolicy, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "block", "":
		return PolicyBlock, nil
	case "ignore":
		return PolicyIgnore, nil
	case "warn":
		return PolicyWarn, nil
	case "report":
		return PolicyReport, nil
	default:
		return PolicyBlock, fmt.Errorf("unknown failure policy: %s", s)
	}
}

// JobConfig represents a GitHub Actions job configuration.
type JobConfig struct {
	Name          string
	AllowFailure  bool
	FailurePolicy FailurePolicy
	Steps         []StepConfig
}

// StepConfig represents a step configuration within a job.
type StepConfig struct {
	Name            string
	ContinueOnError bool
}

// JobResult represents the result of a job execution.
type JobResult struct {
	JobName    string
	Success    bool
	Message    string
	Allowed    bool
	Policy     FailurePolicy
}

// FailureHandler processes job failures according to configured policies.
type FailureHandler struct {
	jobs map[string]*JobConfig
}

// NewFailureHandler creates a new FailureHandler.
func NewFailureHandler() *FailureHandler {
	return &FailureHandler{
		jobs: make(map[string]*JobConfig),
	}
}

// RegisterJob registers a job configuration.
func (h *FailureHandler) RegisterJob(config *JobConfig) {
	h.jobs[config.Name] = config
}

// HandleFailure processes a job failure and returns the result.
func (h *FailureHandler) HandleFailure(jobName string, success bool, message string) JobResult {
	config, exists := h.jobs[jobName]
	if !exists {
		return JobResult{
			JobName: jobName,
			Success: success,
			Message: message,
			Allowed: false,
			Policy:  PolicyBlock,
		}
	}

	if success {
		return JobResult{
			JobName: jobName,
			Success: true,
			Message: message,
			Allowed: false,
			Policy:  config.FailurePolicy,
		}
	}

	// Job failed - check if failure is allowed
	allowed := config.AllowFailure || config.FailurePolicy != PolicyBlock

	return JobResult{
		JobName: jobName,
		Success: false,
		Message: message,
		Allowed: allowed,
		Policy:  config.FailurePolicy,
	}
}

// ShouldBlock returns true if the workflow should be blocked by this result.
func (r JobResult) ShouldBlock() bool {
	if r.Success {
		return false
	}
	return !r.Allowed
}

// Summary returns a human-readable summary of the result.
func (r JobResult) Summary() string {
	if r.Success {
		return fmt.Sprintf("Job %s succeeded", r.JobName)
	}
	if r.Allowed {
		return fmt.Sprintf("Job %s failed (allowed, policy: %s)", r.JobName, r.Policy)
	}
	return fmt.Sprintf("Job %s failed (blocked)", r.JobName)
}

// WorkflowEvaluator evaluates workflow completion based on job results.
type WorkflowEvaluator struct {
	handler *FailureHandler
}

// NewWorkflowEvaluator creates a new WorkflowEvaluator.
func NewWorkflowEvaluator(handler *FailureHandler) *WorkflowEvaluator {
	return &WorkflowEvaluator{handler: handler}
}

// EvaluateWorkflow checks if all jobs passed or if failures are allowed.
func (e *WorkflowEvaluator) EvaluateWorkflow(results []JobResult) (bool, []string) {
	blocked := false
	var blockedJobs []string

	for _, result := range results {
		if result.ShouldBlock() {
			blocked = true
			blockedJobs = append(blockedJobs, result.JobName)
		}
	}

	return !blocked, blockedJobs
}

// FormatReport generates a formatted report of workflow results.
func FormatReport(results []JobResult) string {
	var sb strings.Builder
	sb.WriteString("## Workflow Results\n\n")
	sb.WriteString("| Job | Status | Policy | Blocked |\n")
	sb.WriteString("|-----|--------|--------|--------|\n")

	for _, r := range results {
		status := "✅ Success"
		if !r.Success {
			status = "❌ Failed"
		}
		blocked := "No"
		if r.ShouldBlock() {
			blocked = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			r.JobName, status, r.Policy, blocked))
	}

	return sb.String()
}
