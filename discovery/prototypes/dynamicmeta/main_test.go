package dynamicmeta

import (
	"testing"
)

func TestParseDynamicBlock(t *testing.T) {
	parser := NewParser()
	block := map[string]interface{}{
		"type": "ingress",
		"for_each": map[string]interface{}{
			"port80":  80,
			"port443": 443,
		},
		"content": map[string]interface{}{
			"protocol": "tcp",
		},
		"depends_on": []interface{}{"module.network"},
	}
	db, err := parser.Parse(block)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.Type != "ingress" {
		t.Fatalf("expected type 'ingress', got %s", db.Type)
	}
	if len(db.ForEach) != 2 {
		t.Fatalf("expected 2 for_each entries, got %d", len(db.ForEach))
	}
	if len(db.DependsOn) != 1 {
		t.Fatalf("expected 1 depends_on, got %d", len(db.DependsOn))
	}
}

func TestExpand(t *testing.T) {
	parser := NewParser()
	db := &DynamicBlock{
		Type: "ingress",
		ForEach: map[string]interface{}{
			"port80":  80,
			"port443": 443,
		},
		Content: map[string]interface{}{
			"protocol": "tcp",
		},
	}
	blocks := parser.Expand(db)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 expanded blocks, got %d", len(blocks))
	}
}

func TestValidate(t *testing.T) {
	parser := NewParser()
	valid := &DynamicBlock{
		Type:    "ingress",
		ForEach: map[string]interface{}{"a": 1},
		Content: map[string]interface{}{"b": 2},
	}
	issues := parser.Validate(valid)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %v", issues)
	}

	invalid := &DynamicBlock{}
	issues = parser.Validate(invalid)
	if len(issues) != 3 {
		t.Fatalf("expected 3 issues, got %d: %v", len(issues), issues)
	}
}

func TestMergeMetaArgs(t *testing.T) {
	base := &DynamicBlock{
		Type:      "ingress",
		DependsOn: []string{"a"},
		Providers: map[string]string{"aws": "aws.us"},
	}
	override := &DynamicBlock{
		DependsOn: []string{"b"},
		Providers: map[string]string{"aws": "aws.eu"},
	}
	result := MergeMetaArgs(base, override)
	if len(result.DependsOn) != 2 {
		t.Fatalf("expected 2 depends_on, got %d", len(result.DependsOn))
	}
	if result.Providers["aws"] != "aws.eu" {
		t.Fatalf("expected 'aws.eu', got %s", result.Providers["aws"])
	}
}

func TestFormatBlock(t *testing.T) {
	block := ExpandedBlock{
		Type: "ingress",
		Key:  "port80",
		Config: map[string]interface{}{
			"port": 80,
		},
	}
	s := FormatBlock(block)
	if s == "" {
		t.Fatal("expected non-empty string")
	}
}

func TestToJSON(t *testing.T) {
	blocks := []ExpandedBlock{
		{Type: "ingress", Key: "a", Config: map[string]interface{}{}},
	}
	json := ToJSON(blocks)
	if json == "" {
		t.Fatal("expected non-empty JSON")
	}
}

func BenchmarkExpand(b *testing.B) {
	parser := NewParser()
	db := &DynamicBlock{
		Type: "ingress",
		ForEach: map[string]interface{}{
			"port80":  80,
			"port443": 443,
			"port8080": 8080,
		},
		Content: map[string]interface{}{
			"protocol": "tcp",
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser.Expand(db)
	}
}
