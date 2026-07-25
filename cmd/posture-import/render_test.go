package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderSQLIncludesGuardsMetadataAndRollback verifies safe SQL generation.
func TestRenderSQLIncludesGuardsMetadataAndRollback(t *testing.T) {
	t.Parallel()
	value := manifest{
		AuditID: "audit",
		Reviews: []review{
			{ID: 1, Name: "chair", Status: "sit", Confidence: "high", Source: "manual", Reason: "owner's chair"},
			{ID: 2, Name: "bed", Status: "lay", Confidence: "high", Source: "furniture_data", Reason: "source"},
		},
	}
	output := string(renderSQL(value))
	expected := []string{
		"posture_review_definitions_must_match",
		"posture_review_must_not_override_manual_metadata",
		"'strategy',case when review.status='none' then 'none' else 'derived_footprint' end",
		"jsonb_exists(definition.metadata,'slots')",
		"allow_sit=false where id in (1)",
		"allow_lay=false where id in (2)",
		"owner''s chair",
	}
	for _, fragment := range expected {
		if !strings.Contains(output, fragment) {
			t.Fatalf("missing %q in:\n%s", fragment, output)
		}
	}
}

// TestRunWritesGeneratedSQL verifies the command pipeline writes its output.
func TestRunWritesGeneratedSQL(t *testing.T) {
	t.Parallel()
	if err := run(config{}); err == nil {
		t.Fatal("expected missing input validation")
	}
	if err := run(config{reviewsPath: "missing"}); err == nil {
		t.Fatal("expected missing output validation")
	}
	directory := t.TempDir()
	reviewsPath := filepath.Join(directory, "reviews.json")
	outputPath := filepath.Join(directory, "posture.sql")
	content := `{"audit_id":"audit","reviews":[{"id":1,"name":"chair","status":"sit","confidence":"high","source":"manual","reason":"chair"}]}`
	if err := os.WriteFile(reviewsPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(config{reviewsPath: reviewsPath, outputPath: outputPath}); err != nil {
		t.Fatal(err)
	}
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), "furniture-seed-posture-metadata-0046") {
		t.Fatalf("unexpected generated SQL: %s", output)
	}
	if err = run(config{reviewsPath: reviewsPath, outputPath: directory}); err == nil {
		t.Fatal("expected output write failure")
	}
}
