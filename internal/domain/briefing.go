package domain

type Overlap struct {
	OpenData Insight `json:"open_data"`
	Issue    Insight `json:"issue"`
	Reason   string  `json:"reason"`
}

type Briefing struct {
	RankedOpenData []Insight `json:"ranked_open_data"`
	RankedIssues   []Insight `json:"ranked_issues"`
	Overlaps       []Overlap `json:"overlaps"`
	Skips          []string  `json:"skips"`
}
