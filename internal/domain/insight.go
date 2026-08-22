package domain

type Pipeline string

const (
	PipelineOpenData Pipeline = "opendata"
	PipelineIssues   Pipeline = "issues"
)

const (
	KindOpportunityAlert = "opportunity_alert"
	KindPainPoint        = "pain_point"
)

type Insight struct {
	Fingerprint string             `json:"fingerprint"`
	Pipeline    Pipeline           `json:"pipeline"`
	Source      string             `json:"source"`
	Kind        string             `json:"kind"`
	Theme       string             `json:"theme"`
	Title       string             `json:"title"`
	URL         string             `json:"url"`
	Score       float64            `json:"score"`
	Metrics     map[string]float64 `json:"metrics"`
	Keywords    []string           `json:"keywords"`
	Summary     string             `json:"summary"`
	Body        string             `json:"body,omitempty"` // issue body for LLM
}
