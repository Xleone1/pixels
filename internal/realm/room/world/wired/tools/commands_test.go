package tools

import "testing"

// TestNormalizeTab verifies command aliases cannot request an unknown panel tab.
func TestNormalizeTab(t *testing.T) {
	cases := map[string]string{
		"variables":    "variables",
		" inspection ": "inspection",
		"settings":     "settings",
		"unknown":      "monitor",
	}
	for input, expected := range cases {
		if actual := normalizeTab(input); actual != expected {
			t.Fatalf("normalize tab %q = %q, want %q", input, actual, expected)
		}
	}
}
