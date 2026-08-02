package tests

import (
	"errors"
	"testing"

	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// TestCompileClickSourcesPreserveEveryEditorMode verifies source-specific target contracts.
func TestCompileClickSourcesPreserveEveryEditorMode(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	tests := []struct {
		name    string
		values  []int32
		mode    int32
		targets []record.Target
		valid   bool
	}{
		{name: "own box empty", valid: true},
		{name: "own box explicit", values: []int32{0}, valid: true},
		{name: "selected", values: []int32{100}, mode: 1, targets: []record.Target{{ItemID: 2}}, valid: true},
		{name: "selector", values: []int32{200}, valid: true},
		{name: "selected missing target", values: []int32{100}},
		{name: "selector with target", values: []int32{200}, mode: 1, targets: []record.Target{{ItemID: 2}}},
		{name: "unknown", values: []int32{300}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, compileErr := compiler.CompileNode(record.Config{
				ItemID: 1, RoomID: 1, Interaction: "wf_trg_user_clicks_tile",
				IntParams: test.values, SelectionMode: test.mode, Targets: test.targets,
			})
			if test.valid && compileErr != nil {
				t.Fatalf("CompileNode() error=%v", compileErr)
			}
			if !test.valid && !errors.Is(compileErr, configuration.ErrInvalid) {
				t.Fatalf("CompileNode() error=%v", compileErr)
			}
		})
	}
}
