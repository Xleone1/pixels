package admin

import (
	"os"
	"testing"
)

// TestLoadConfigReadsEffectClearCompatibility verifies the public environment toggle.
func TestLoadConfigReadsEffectClearCompatibility(t *testing.T) {
	t.Setenv("PIXELS_EFFECT_ALLOW_UNPERMITTED_CLEAR", "false")

	config, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.AllowUnpermittedEffectClear {
		t.Fatal("expected unpermitted effect clearing to be disabled")
	}
}

// TestLoadConfigAllowsEffectClearByDefault verifies the safe self-service default.
func TestLoadConfigAllowsEffectClearByDefault(t *testing.T) {
	value, present := os.LookupEnv("PIXELS_EFFECT_ALLOW_UNPERMITTED_CLEAR")
	_ = os.Unsetenv("PIXELS_EFFECT_ALLOW_UNPERMITTED_CLEAR")
	t.Cleanup(func() {
		if present {
			_ = os.Setenv("PIXELS_EFFECT_ALLOW_UNPERMITTED_CLEAR", value)
		}
	})

	config, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !config.AllowUnpermittedEffectClear {
		t.Fatal("expected unpermitted effect clearing to be enabled by default")
	}
}
