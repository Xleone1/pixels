package tests

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"
)

// TestDevelopmentLabRoomNamesFitPersistence verifies every QA room satisfies the durable room-name constraint.
func TestDevelopmentLabRoomNamesFitPersistence(t *testing.T) {
	contents, err := os.ReadFile(repositoryPath(t, "internal/realm/room/database/seed/development/0012_rebuild_wired_labs.sql"))
	if err != nil {
		t.Fatal(err)
	}
	expression := regexp.MustCompile(`when 11\d then '([^']+)'`)
	names := statementBody(string(contents), "set name = case id", "end,\n    description")
	matches := expression.FindAllStringSubmatch(names, -1)
	if len(matches) != 6 {
		t.Fatalf("room fixtures=%d, want 6", len(matches))
	}
	for _, match := range matches {
		length := utf8.RuneCountInString(match[1])
		if length < 3 || length > 25 {
			t.Errorf("room name %q has %d characters, want 3..25", match[1], length)
		}
	}
	if !strings.Contains(string(contents), "staff_picked = false") {
		t.Fatal("WIRED QA rooms must not be staff picked")
	}
}

// TestDedicatedModernLabsAndProjectileObstacles verifies reproducible QA room fixtures.
func TestDedicatedModernLabsAndProjectileObstacles(t *testing.T) {
	rooms := readSeedFiles(t,
		"internal/realm/room/database/seed/development/0023_wired_selector_variable_rooms.sql",
	)
	furniture := readSeedFiles(t,
		"internal/realm/furniture/database/seed/development/0055_wired_selector_variable_labs.sql",
		"internal/realm/furniture/database/seed/development/0056_wired_projectile_obstacles.sql",
	)
	bots := readSeedFiles(t,
		"internal/realm/bot/database/seed/development/0004_wired_projectile_obstacle.sql",
	)
	for _, fragment := range []string{
		"lower(username)='niflaot'", "'WIRED QA Selectors'",
		"'WIRED QA Variables'", "'model_5'",
	} {
		if !strings.Contains(string(rooms), fragment) {
			t.Errorf("modern room seed is missing %q", fragment)
		}
	}
	for _, fragment := range []string{
		"definition.interaction_type like 'wf_slc_%'", "definition.interaction_type like 'wf_var_%'",
		"(1030700,'wf_xtra_projectile'", "(1030706,'table_plasto_4leg'",
		"(1030703,1030705", "'[2,5,1]'",
	} {
		if !strings.Contains(string(furniture), fragment) {
			t.Errorf("modern furniture seed is missing %q", fragment)
		}
	}
	if !strings.Contains(string(bots), "'ProjectileUser'") ||
		!strings.Contains(string(bots), "'F',7,7,0,6,false") {
		t.Fatal("projectile user obstacle is incomplete")
	}
}
