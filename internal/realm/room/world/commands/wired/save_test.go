package wired

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// TestPrivilegedProgression verifies only the three custom progression effects require superwired.
func TestPrivilegedProgression(t *testing.T) {
	tests := map[string]bool{"wf_act_progress_achievement": true, "wf_act_progress_quest": true, "wf_act_start_quest": true, "wf_act_give_score": false, "wf_act_reset_highscore": false}
	for key, expected := range tests {
		if actual := privilegedProgression(key); actual != expected {
			t.Errorf("key=%s actual=%v expected=%v", key, actual, expected)
		}
	}
}

// TestModernFamiliesUseExistingEditorProtocol verifies protocol compatibility.
func TestModernFamiliesUseExistingEditorProtocol(t *testing.T) {
	effectFamilies := []registry.Family{
		registry.FamilyEffect,
		registry.FamilyExtra,
		registry.FamilySelector,
		registry.FamilyVariable,
	}
	for _, family := range effectFamilies {
		if !editableFamily(family) {
			t.Errorf("family %d is not editable", family)
		}
		if actual := familyOf(family); actual != EffectFamily {
			t.Errorf("family %d mapped to %d", family, actual)
		}
	}
	if actual := familyOf(registry.FamilyTrigger); actual != TriggerFamily {
		t.Errorf("trigger mapped to %d", actual)
	}
	if actual := familyOf(registry.FamilyCondition); actual != ConditionFamily {
		t.Errorf("condition mapped to %d", actual)
	}
}
