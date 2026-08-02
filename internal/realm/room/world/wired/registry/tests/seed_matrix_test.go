package tests

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// TestDevelopmentLabsPlaceAndConfigureManifest verifies all canonical and compatibility behaviors have QA fixtures.
func TestDevelopmentLabsPlaceAndConfigureManifest(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	definitions := seededDefinitions(t)
	placed := make(map[string]struct{})
	configured := make(map[string]struct{})
	for _, relative := range []string{
		"internal/realm/furniture/database/seed/development/0026_rebuild_wired_labs.sql",
		"internal/realm/furniture/database/seed/development/0038_room_games.sql",
		"internal/realm/furniture/database/seed/development/0056_wired_projectile_obstacles.sql",
	} {
		contents, readErr := os.ReadFile(repositoryPath(t, relative))
		if readErr != nil {
			t.Fatal(readErr)
		}
		for itemID, interaction := range placedItems(t, string(contents), definitions) {
			descriptor, found := registered.Resolve(interaction)
			if !found {
				continue
			}
			placed[descriptor.Key] = struct{}{}
			if configuredItems(string(contents))[itemID] {
				configured[descriptor.Key] = struct{}{}
			}
		}
	}
	want := append(registry.CanonicalManifest(), registry.CompatibilityManifest()...)
	for _, descriptor := range want {
		if isManifestExpansion(descriptor.Key) {
			continue
		}
		if _, found := placed[descriptor.Key]; !found {
			t.Errorf("behavior %s has no placed QA fixture", descriptor.Key)
		}
		if requiresSettings(descriptor) {
			if _, found := configured[descriptor.Key]; !found {
				t.Errorf("behavior %s has no configured QA fixture", descriptor.Key)
			}
		}
	}
	if len(placed) != len(want)-62 {
		t.Fatalf("placed classic behavior keys=%d, want %d", len(placed), len(want)-62)
	}
}

// TestDevelopmentWiredExpansionSeed covers all modern definitions, catalog rows, and dynamic QA placements.
func TestDevelopmentWiredExpansionSeed(t *testing.T) {
	definitions := readSeedFiles(t,
		"internal/realm/furniture/database/seed/development/0050_wired_definitions_expansion.sql",
		"internal/realm/furniture/database/seed/development/0053_wired_clicks_actions.sql",
	)
	catalog := readSeedFiles(t,
		"internal/realm/catalog/database/seed/development/0033_wired_catalog.sql",
		"internal/realm/catalog/database/seed/development/0035_wired_clicks_actions.sql",
	)
	labs := readSeedFiles(t,
		"internal/realm/furniture/database/seed/development/0051_wired_dynamic_labs.sql",
		"internal/realm/furniture/database/seed/development/0053_wired_clicks_actions.sql",
	)
	for _, descriptor := range registry.CanonicalManifest() {
		if !isManifestExpansion(descriptor.Key) {
			continue
		}
		if !strings.Contains(string(definitions), "'"+descriptor.Key+"'") {
			t.Errorf("WIRED definition %s is absent", descriptor.Key)
		}
	}
	if !strings.Contains(string(catalog), "metadata->>'source'='polaris-wired'") ||
		!strings.Contains(string(catalog), "overriding system value") ||
		!strings.Contains(string(labs), "item.id between 1010000 and 1010053") ||
		!strings.Contains(string(labs), "wired-integration") ||
		!strings.Contains(string(labs), "room_invisible_click_tile") {
		t.Fatal("WIRED catalog or integration lab contract is incomplete")
	}
}

// readSeedFiles joins related additive seed documents.
func readSeedFiles(t *testing.T, relatives ...string) []byte {
	t.Helper()
	var result []byte
	for _, relative := range relatives {
		contents, err := os.ReadFile(repositoryPath(t, relative))
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, contents...)
	}
	return result
}

// TestDevelopmentWiredCatalogOrganization verifies the one-level catalog hierarchy.
func TestDevelopmentWiredCatalogOrganization(t *testing.T) {
	path := repositoryPath(
		t,
		"internal/realm/catalog/database/seed/development/0034_wired_catalog_organization.sql",
	)
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	required := []string{
		"'wired_classic'",
		"'wired_advanced'",
		"'wired_advanced_components'",
		"definition.metadata->>'source'='polaris-wired'",
		"when definition.name like 'wf_slc_%' then 1007",
		"when definition.name like 'wf_var_%' then 1008",
	}
	for _, fragment := range required {
		if !strings.Contains(string(contents), fragment) {
			t.Errorf("catalog organization is missing %q", fragment)
		}
	}
}

// isManifestExpansion reports whether a descriptor belongs to the pinned Polaris addition.
func isManifestExpansion(key string) bool {
	for _, descriptor := range registry.CanonicalManifest()[76:] {
		if descriptor.Key == key {
			return true
		}
	}
	return false
}

// TestExtraLabSelectedTriggersUseExplicitSelectionMode prevents invalid seeded target selections.
func TestExtraLabSelectedTriggersUseExplicitSelectionMode(t *testing.T) {
	path := repositoryPath(t, "internal/realm/furniture/database/seed/development/0031_wired_lab_extras.sql")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, itemID := range []string{"426000", "426010", "426020", "426030"} {
		expected := "(" + itemID + ",'[]','',1,0)"
		if !strings.Contains(string(contents), expected) {
			t.Errorf("WIRED item %s has no explicit selection mode", itemID)
		}
	}
}

// seededDefinitions maps development furniture definition ids to interactions.
func seededDefinitions(t *testing.T) map[int64]string {
	t.Helper()
	result := make(map[int64]string)
	for _, relative := range []string{
		"internal/realm/furniture/database/seed/development/0021_wired_definitions.sql",
		"internal/realm/furniture/database/seed/development/0023_wired_compatibility.sql",
		"internal/realm/furniture/database/seed/development/0038_room_games.sql",
		"internal/realm/furniture/database/seed/development/0054_wired_projectile.sql",
	} {
		contents, err := os.ReadFile(repositoryPath(t, relative))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(contents), "\n") {
			fields, valid := tupleFields(line)
			if !valid || len(fields) < 14 {
				continue
			}
			id, parseErr := strconv.ParseInt(fields[0], 10, 64)
			if parseErr == nil {
				result[id] = unquote(fields[13])
			}
		}
	}
	return result
}

// placedItems returns QA furniture ids and their canonical interactions.
func placedItems(t *testing.T, contents string, definitions map[int64]string) map[int64]string {
	t.Helper()
	result := make(map[int64]string)
	for _, match := range regexp.MustCompile(`\((\d+),(\d+),1,11[0-5],`).FindAllStringSubmatch(contents, -1) {
		itemID, itemErr := strconv.ParseInt(match[1], 10, 64)
		definitionID, definitionErr := strconv.ParseInt(match[2], 10, 64)
		if itemErr != nil || definitionErr != nil {
			t.Fatalf("invalid placed item tuple %v", match)
		}
		result[itemID] = definitions[definitionID]
	}
	fixture := statementBody(contents, "fixture(id,definition_id", ")\ninsert into furniture_items")
	for _, match := range regexp.MustCompile(`\((\d+),(\d+),`).FindAllStringSubmatch(fixture, -1) {
		itemID, itemErr := strconv.ParseInt(match[1], 10, 64)
		definitionID, definitionErr := strconv.ParseInt(match[2], 10, 64)
		if itemErr != nil || definitionErr != nil {
			t.Fatalf("invalid fixture tuple %v", match)
		}
		result[itemID] = definitions[definitionID]
	}
	namedFixture := statementBody(contents, "fixture(id,definition_name", ")\ninsert into furniture_items")
	for _, match := range regexp.MustCompile(`\((\d+),'([^']+)',`).FindAllStringSubmatch(namedFixture, -1) {
		itemID, itemErr := strconv.ParseInt(match[1], 10, 64)
		if itemErr != nil {
			t.Fatalf("invalid named fixture tuple %v", match)
		}
		result[itemID] = match[2]
	}
	return result
}

// configuredItems returns item ids from one normalized settings statement.
func configuredItems(contents string) map[int64]bool {
	start := "insert into room_wired_settings"
	if strings.Contains(contents, "with desired(item_id,int_params") {
		start = "with desired(item_id,int_params"
	}
	statement := statementBody(contents, start, "on conflict(item_id)")
	expression := regexp.MustCompile(`\((\d+),`)
	result := make(map[int64]bool)
	for _, match := range expression.FindAllStringSubmatch(statement, -1) {
		if itemID, err := strconv.ParseInt(match[1], 10, 64); err == nil {
			result[itemID] = true
		}
	}
	return result
}

// statementBody returns SQL between one insertion prefix and conflict clause.
func statementBody(contents string, start string, end string) string {
	startIndex := strings.Index(contents, start)
	if startIndex < 0 {
		return ""
	}
	contents = contents[startIndex:]
	endIndex := strings.Index(contents, end)
	if endIndex < 0 {
		return contents
	}
	return contents[:endIndex]
}

// requiresSettings reports whether one QA behavior compiles persisted configuration.
func requiresSettings(descriptor registry.Descriptor) bool {
	return descriptor.Family != registry.FamilyHighscore && descriptor.Key != "wf_blob"
}
