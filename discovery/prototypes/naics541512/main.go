// Package naics541512 implements search for SAM.gov NAICS 541512 opportunities.
// It provides a search query builder and result parser for finding programming
// services contracts under NAICS code 541512 ("Computer Programming Services").
package naics541512

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// Opportunity represents a contract opportunity from SAM.gov.
type Opportunity struct {
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	Agency             string    `json:"agency"`
	SubAgency          string    `json:"sub_agency,omitempty"`
	SetAside           string    `json:"set_aside,omitempty"`
	NAICSCode          string    `json:"naics_code"`
	PSCCode            string    `json:"psc_code,omitempty"`
	AwardAmount        string    `json:"award_amount,omitempty"`
	ContractType       string    `json:"contract_type,omitempty"`
	PostedDate         time.Time `json:"posted_date"`
	ResponseDueDate    time.Time `json:"response_due_date,omitempty"`
	Link               string    `json:"link"`
}

// SearchQuery represents a search query for SAM.gov.
type SearchQuery struct {
	NAICSCode  string `json:"naics_code"`
	Keyword    string `json:"keyword,omitempty"`
	Agency     string `json:"agency,omitempty"`
	SetAside   string `json:"set_aside,omitempty"`
	StartDate  string `json:"start_date,omitempty"`
	EndDate    string `json:"end_date,omitempty"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

// SearchResult represents search results from SAM.gov.
type SearchResult struct {
	Opportunities []Opportunity `json:"opportunities"`
	TotalCount    int           `json:"total_count"`
	Page          int           `json:"page"`
	PageSize      int           `json:"page_size"`
}

// Searcher searches for SAM.gov opportunities.
type Searcher struct{}

// NewSearcher creates a new Searcher.
func NewSearcher() *Searcher {
	return &Searcher{}
}

// BuildQuery builds a search query for NAICS 541512.
func (s *Searcher) BuildQuery(keyword string, filters map[string]string) SearchQuery {
	query := SearchQuery{
		NAICSCode: "541512",
		Keyword:   keyword,
		Page:      1,
		PageSize:  25,
	}
	if agency, ok := filters["agency"]; ok {
		query.Agency = agency
	}
	if setAside, ok := filters["set_aside"]; ok {
		query.SetAside = setAside
	}
	if startDate, ok := filters["start_date"]; ok {
		query.StartDate = startDate
	}
	if endDate, ok := filters["end_date"]; ok {
		query.EndDate = endDate
	}
	return query
}

// FormatQuery formats a search query as a string.
func (s *Searcher) FormatQuery(query SearchQuery) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("NAICS: %s", query.NAICSCode))
	if query.Keyword != "" {
		parts = append(parts, fmt.Sprintf("Keyword: %s", query.Keyword))
	}
	if query.Agency != "" {
		parts = append(parts, fmt.Sprintf("Agency: %s", query.Agency))
	}
	if query.SetAside != "" {
		parts = append(parts, fmt.Sprintf("Set-Aside: %s", query.SetAside))
	}
	return strings.Join(parts, " | ")
}

// ParseOpportunity parses an opportunity from a map.
func (s *Searcher) ParseOpportunity(data map[string]interface{}) Opportunity {
	opp := Opportunity{}
	if title, ok := data["title"].(string); ok {
		opp.Title = title
	}
	if desc, ok := data["description"].(string); ok {
		opp.Description = desc
	}
	if agency, ok := data["agency"].(string); ok {
		opp.Agency = agency
	}
	if subAgency, ok := data["sub_agency"].(string); ok {
		opp.SubAgency = subAgency
	}
	if setAside, ok := data["set_aside"].(string); ok {
		opp.SetAside = setAside
	}
	if naics, ok := data["naics_code"].(string); ok {
		opp.NAICSCode = naics
	}
	if psc, ok := data["psc_code"].(string); ok {
		opp.PSCCode = psc
	}
	if amount, ok := data["award_amount"].(string); ok {
		opp.AwardAmount = amount
	}
	if contractType, ok := data["contract_type"].(string); ok {
		opp.ContractType = contractType
	}
	if link, ok := data["link"].(string); ok {
		opp.Link = link
	}
	return opp
}

// FormatOpportunity formats an opportunity as a string.
func (s *Searcher) FormatOpportunity(opp Opportunity) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Title: %s\n", opp.Title))
	sb.WriteString(fmt.Sprintf("Agency: %s\n", opp.Agency))
	if opp.SetAside != "" {
		sb.WriteString(fmt.Sprintf("Set-Aside: %s\n", opp.SetAside))
	}
	sb.WriteString(fmt.Sprintf("NAICS: %s\n", opp.NAICSCode))
	if opp.AwardAmount != "" {
		sb.WriteString(fmt.Sprintf("Amount: %s\n", opp.AwardAmount))
	}
	if opp.Link != "" {
		sb.WriteString(fmt.Sprintf("Link: %s\n", opp.Link))
	}
	return sb.String()
}

// SortByDate sorts opportunities by posted date.
func SortByDate(opps []Opportunity) []Opportunity {
	sorted := make([]Opportunity, len(opps))
	copy(sorted, opps)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PostedDate.After(sorted[j].PostedDate)
	})
	return sorted
}

// FilterByAgency filters opportunities by agency.
func FilterByAgency(opps []Opportunity, agency string) []Opportunity {
	var filtered []Opportunity
	for _, opp := range opps {
		if strings.EqualFold(opp.Agency, agency) {
			filtered = append(filtered, opp)
		}
	}
	return filtered
}

// FormatResults formats search results as JSON.
func FormatResults(result SearchResult) string {
	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b)
}
