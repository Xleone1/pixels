package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadManifestValidatesAndSorts verifies deterministic validated loading.
func TestLoadManifestValidatesAndSorts(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "reviews.json")
	content := `{"audit_id":"audit","reviews":[{"id":2,"name":"bed","status":"lay","confidence":"high","source":"furniture_data","reason":"source"},{"id":1,"name":"chair","status":"sit","confidence":"medium","source":"visual_review","reason":"icon"}]}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := loadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.Reviews[0].ID != 1 || result.Reviews[1].ID != 2 {
		t.Fatalf("reviews are not sorted: %+v", result.Reviews)
	}
}

// TestLoadManifestRejectsInvalidEvidence verifies malformed reviews fail fast.
func TestLoadManifestRejectsInvalidEvidence(t *testing.T) {
	t.Parallel()
	cases := []string{
		`{}`,
		`{"audit_id":"audit","reviews":[]}`,
		`{"audit_id":"audit","reviews":[{"id":0,"name":"chair","status":"sit","confidence":"high","source":"manual","reason":"x"}]}`,
		`{"audit_id":"audit","reviews":[{"id":1,"name":"","status":"sit","confidence":"high","source":"manual","reason":"x"}]}`,
		`{"audit_id":"audit","reviews":[{"id":1,"name":"chair","status":"dance","confidence":"high","source":"manual","reason":"x"}]}`,
		`{"audit_id":"audit","reviews":[{"id":1,"name":"chair","status":"sit","confidence":"guess","source":"manual","reason":"x"}]}`,
		`{"audit_id":"audit","reviews":[{"id":1,"name":"chair","status":"sit","confidence":"high","source":"unknown","reason":"x"}]}`,
		`{"audit_id":"audit","reviews":[{"id":1,"name":"chair","status":"sit","confidence":"high","source":"manual","reason":""}]}`,
		`{"audit_id":"audit","reviews":[{"id":1,"name":"chair","status":"sit","confidence":"high","source":"manual","reason":"x","previous_allow_sit":true}]}`,
		`{"audit_id":"audit","reviews":[{"id":1,"name":"bed","status":"lay","confidence":"high","source":"manual","reason":"x","previous_allow_lay":true}]}`,
		`{"audit_id":"audit","reviews":[{"id":1,"name":"chair","status":"sit","confidence":"high","source":"manual","reason":"x"},{"id":1,"name":"chair","status":"sit","confidence":"high","source":"manual","reason":"x"}]}`,
		`not json`,
	}
	for index, content := range cases {
		path := filepath.Join(t.TempDir(), "invalid.json")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadManifest(path); err == nil {
			t.Fatalf("case %d unexpectedly passed", index)
		}
	}
	if _, err := loadManifest("missing.json"); err == nil || !strings.Contains(err.Error(), "read reviews") {
		t.Fatalf("unexpected missing file error: %v", err)
	}
}
