// Package dynamicmeta implements dynamic blocks with meta-arguments for Terraform.
// It provides a parser that can handle dynamic blocks with meta-arguments
// like depends_on, providers, and lifecycle.
package dynamicmeta

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MetaArg represents a meta-argument that can be applied to dynamic blocks.
type MetaArg struct {
	Name  string
	Value interface{}
}

// DynamicBlock represents a dynamic block with optional meta-arguments.
type DynamicBlock struct {
	Type      string                 `json:"type"`
	ForEach   map[string]interface{} `json:"for_each"`
	Content   map[string]interface{} `json:"content"`
	DependsOn []string               `json:"depends_on,omitempty"`
	Providers map[string]string      `json:"providers,omitempty"`
	Lifecycle map[string]interface{} `json:"lifecycle,omitempty"`
}

// ExpandedBlock represents a single expanded instance from a dynamic block.
type ExpandedBlock struct {
	Type      string                 `json:"type"`
	Key       string                 `json:"key"`
	Config    map[string]interface{} `json:"config"`
	DependsOn []string               `json:"depends_on,omitempty"`
	Providers map[string]string      `json:"providers,omitempty"`
	Lifecycle map[string]interface{} `json:"lifecycle,omitempty"`
}

// Parser handles parsing of dynamic blocks with meta-arguments.
type Parser struct{}

// NewParser creates a new Parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses a dynamic block definition and returns a DynamicBlock.
func (p *Parser) Parse(block map[string]interface{}) (*DynamicBlock, error) {
	db := &DynamicBlock{
		ForEach: make(map[string]interface{}),
		Content: make(map[string]interface{}),
	}

	// Parse type
	if t, ok := block["type"].(string); ok {
		db.Type = t
	} else {
		return nil, fmt.Errorf("missing or invalid type")
	}

	// Parse for_each
	if fe, ok := block["for_each"].(map[string]interface{}); ok {
		db.ForEach = fe
	}

	// Parse content
	if c, ok := block["content"].(map[string]interface{}); ok {
		db.Content = c
	}

	// Parse meta-arguments
	if do, ok := block["depends_on"].([]interface{}); ok {
		for _, d := range do {
			if s, ok := d.(string); ok {
				db.DependsOn = append(db.DependsOn, s)
			}
		}
	}

	if prov, ok := block["providers"].(map[string]interface{}); ok {
		db.Providers = make(map[string]string)
		for k, v := range prov {
			if s, ok := v.(string); ok {
				db.Providers[k] = s
			}
		}
	}

	if lc, ok := block["lifecycle"].(map[string]interface{}); ok {
		db.Lifecycle = lc
	}

	return db, nil
}

// Expand expands a dynamic block into individual instances.
func (p *Parser) Expand(db *DynamicBlock) []ExpandedBlock {
	var blocks []ExpandedBlock
	for key, _ := range db.ForEach {
		block := ExpandedBlock{
			Type:      db.Type,
			Key:       key,
			Config:    make(map[string]interface{}),
			DependsOn: db.DependsOn,
			Providers: db.Providers,
			Lifecycle: db.Lifecycle,
		}
		// Copy content
		for k, v := range db.Content {
			block.Config[k] = v
		}
		blocks = append(blocks, block)
	}
	return blocks
}

// Validate checks if a dynamic block definition is valid.
func (p *Parser) Validate(db *DynamicBlock) []string {
	var issues []string
	if db.Type == "" {
		issues = append(issues, "missing type")
	}
	if len(db.ForEach) == 0 {
		issues = append(issues, "empty for_each")
	}
	if len(db.Content) == 0 {
		issues = append(issues, "empty content")
	}
	return issues
}

// MergeMetaArgs merges meta-arguments from multiple sources.
func MergeMetaArgs(base, override *DynamicBlock) *DynamicBlock {
	result := &DynamicBlock{
		Type:    base.Type,
		ForEach: base.ForEach,
		Content: base.Content,
	}

	// Merge depends_on
	result.DependsOn = append(base.DependsOn, override.DependsOn...)

	// Merge providers (override takes precedence)
	result.Providers = make(map[string]string)
	for k, v := range base.Providers {
		result.Providers[k] = v
	}
	for k, v := range override.Providers {
		result.Providers[k] = v
	}

	// Merge lifecycle (override takes precedence)
	result.Lifecycle = make(map[string]interface{})
	for k, v := range base.Lifecycle {
		result.Lifecycle[k] = v
	}
	for k, v := range override.Lifecycle {
		result.Lifecycle[k] = v
	}

	return result
}

// FormatBlock returns a formatted string representation of an expanded block.
func FormatBlock(block ExpandedBlock) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s.%s {\n", block.Type, block.Key))
	for k, v := range block.Config {
		sb.WriteString(fmt.Sprintf("  %s = %v\n", k, v))
	}
	if len(block.DependsOn) > 0 {
		sb.WriteString("  depends_on = [")
		sb.WriteString(strings.Join(block.DependsOn, ", "))
		sb.WriteString("]\n")
	}
	if len(block.Providers) > 0 {
		sb.WriteString("  providers = {\n")
		for k, v := range block.Providers {
			sb.WriteString(fmt.Sprintf("    %s = %s\n", k, v))
		}
		sb.WriteString("  }\n")
	}
	sb.WriteString("}")
	return sb.String()
}

// ToJSON returns the expanded blocks as JSON.
func ToJSON(blocks []ExpandedBlock) string {
	b, _ := json.MarshalIndent(blocks, "", "  ")
	return string(b)
}
