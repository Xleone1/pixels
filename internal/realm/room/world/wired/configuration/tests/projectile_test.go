package tests

import (
	"errors"
	"testing"

	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// TestCompileRejectsRoomFurnitureAboveConfiguredLimit verifies the monitor ceiling is authoritative.
func TestCompileRejectsRoomFurnitureAboveConfiguredLimit(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{MaxFurniPerRoom: 1})
	_, err = compiler.Compile(1, 1, []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room"},
		{ItemID: 2, RoomID: 1, Interaction: "wf_trg_enter_room"},
	})
	if !errors.Is(err, configuration.ErrInvalid) {
		t.Fatalf("Compile() error = %v", err)
	}
}

// TestCompileAcceptsBoundedProjectileParameters verifies the custom editor schema.
func TestCompileAcceptsBoundedProjectileParameters(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	node, err := compiler.CompileNode(record.Config{
		ItemID: 1, RoomID: 1, Interaction: "wf_xtra_projectile",
		IntParams: []int32{7, 64, 20}, SelectionMode: 1,
		Targets: []record.Target{{ItemID: 2}},
	})
	if err != nil || len(node.Parameters.Values) != 3 {
		t.Fatalf("projectile node=%#v err=%v", node, err)
	}
}

// TestCompileRejectsProjectileParametersOutsideConfiguredLimits verifies editor saves fail instead of clamping.
func TestCompileRejectsProjectileParametersOutsideConfiguredLimits(t *testing.T) {
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{
		MaxProjectileDistance: 3, MaxProjectileDurationPulses: 4,
	})
	tests := []struct {
		name   string
		values []int32
	}{
		{name: "distance", values: []int32{2, 4, 1}},
		{name: "duration", values: []int32{2, 3, 2}},
		{name: "speed", values: []int32{2, 1, 21}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, compileErr := compiler.CompileNode(record.Config{
				ItemID: 1, RoomID: 1, Interaction: "wf_xtra_projectile",
				IntParams: test.values, SelectionMode: 1,
				Targets: []record.Target{{ItemID: 2}},
			})
			if !errors.Is(compileErr, configuration.ErrInvalid) {
				t.Fatalf("CompileNode() error=%v", compileErr)
			}
		})
	}
}
