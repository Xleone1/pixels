package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// manifest contains one auditable furniture posture review.
type manifest struct {
	AuditID string   `json:"audit_id"`
	Reviews []review `json:"reviews"`
}

// review contains one furniture posture decision and its provenance.
type review struct {
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	Confidence       string `json:"confidence"`
	Source           string `json:"source"`
	Reason           string `json:"reason"`
	PreviousAllowSit bool   `json:"previous_allow_sit"`
	PreviousAllowLay bool   `json:"previous_allow_lay"`
}

// loadManifest reads and validates a posture review manifest.
func loadManifest(path string) (manifest, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return manifest{}, fmt.Errorf("read reviews: %w", err)
	}
	result := manifest{}
	if err = json.Unmarshal(content, &result); err != nil {
		return manifest{}, fmt.Errorf("decode reviews: %w", err)
	}
	if err = validateManifest(result); err != nil {
		return manifest{}, err
	}
	sort.Slice(result.Reviews, func(left, right int) bool {
		return result.Reviews[left].ID < result.Reviews[right].ID
	})
	return result, nil
}

// validateManifest checks identifiers, decisions, evidence, and duplicates.
func validateManifest(value manifest) error {
	if strings.TrimSpace(value.AuditID) == "" {
		return fmt.Errorf("audit_id is required")
	}
	if len(value.Reviews) == 0 {
		return fmt.Errorf("at least one review is required")
	}
	seen := make(map[int64]struct{}, len(value.Reviews))
	for _, item := range value.Reviews {
		if err := validateReview(item); err != nil {
			return fmt.Errorf("review %d: %w", item.ID, err)
		}
		if _, exists := seen[item.ID]; exists {
			return fmt.Errorf("duplicate furniture id %d", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
	return nil
}

// validateReview checks one furniture posture decision.
func validateReview(item review) error {
	if item.ID <= 0 {
		return fmt.Errorf("id must be positive")
	}
	if strings.TrimSpace(item.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if !oneOf(item.Status, "sit", "lay", "none") {
		return fmt.Errorf("unsupported status %q", item.Status)
	}
	if !oneOf(item.Confidence, "high", "medium", "low") {
		return fmt.Errorf("unsupported confidence %q", item.Confidence)
	}
	if !oneOf(item.Source, "furniture_data", "visual_review", "manual") {
		return fmt.Errorf("unsupported source %q", item.Source)
	}
	if strings.TrimSpace(item.Reason) == "" {
		return fmt.Errorf("reason is required")
	}
	if item.Status == "sit" && item.PreviousAllowSit {
		return fmt.Errorf("sit decision was already enabled")
	}
	if item.Status == "lay" && item.PreviousAllowLay {
		return fmt.Errorf("lay decision was already enabled")
	}
	return nil
}

// oneOf reports whether value equals one allowed string.
func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
