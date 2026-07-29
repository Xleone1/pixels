package wired

import (
	"context"
	"errors"
	"testing"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/condition"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	nativeregistry "github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	sdkwired "github.com/niflaot/pixels/sdk/wired"
)

// TestAccessEnforcesNamespaceAndDuplicates verifies registration ownership.
func TestAccessEnforcesNamespaceAndDuplicates(t *testing.T) {
	registered := NewRegistry(time.Second, nil)
	access := NewAccess(registered, pluginruntime.NewScope("demo"))
	executor := func(context.Context, sdkwired.Node, sdkwired.Event) (sdkwired.EffectResult, error) {
		return sdkwired.EffectResult{Status: sdkwired.EffectApplied}, nil
	}
	if err := access.RegisterEffect("plugin.other.effect", executor); !errors.Is(err, pluginruntime.ErrWrongNamespace) {
		t.Fatalf("expected namespace error, got %v", err)
	}
	if err := access.RegisterEffect("plugin.demo.effect", executor); err != nil {
		t.Fatalf("register effect: %v", err)
	}
	if err := access.RegisterEffect("plugin.demo.effect", executor); !errors.Is(err, ErrDuplicateKey) {
		t.Fatalf("expected duplicate error, got %v", err)
	}
}

// TestRegistryCompilesAndExecutesPluginBehavior verifies every fallback boundary.
func TestRegistryCompilesAndExecutesPluginBehavior(t *testing.T) {
	registered := NewRegistry(time.Second, nil)
	access := NewAccess(registered, pluginruntime.NewScope("demo"))
	if err := access.RegisterEffect("plugin.demo.effect", func(_ context.Context, node sdkwired.Node, event sdkwired.Event) (sdkwired.EffectResult, error) {
		if node.ItemID != 7 || event.PlayerID != 9 {
			t.Fatalf("unexpected callback input node=%#v event=%#v", node, event)
		}
		return sdkwired.EffectResult{Status: sdkwired.EffectApplied, ResetTimers: true}, nil
	}, sdkwired.WithDescriptor(sdkwired.Descriptor{Selection: sdkwired.SelectionNone, Editor: true})); err != nil {
		t.Fatalf("register effect: %v", err)
	}
	native, err := nativeregistry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(native, roomwired.Config{}).WithExtensions(registered)
	descriptor, found := compiler.ResolveDescriptor("plugin.demo.effect")
	if !found || !descriptor.Editor || descriptor.Family != nativeregistry.FamilyEffect {
		t.Fatalf("descriptor=%#v found=%v", descriptor, found)
	}
	node, err := compiler.CompileNode(record.Config{ItemID: 7, RoomID: 3, Interaction: "plugin.demo.effect"})
	if err != nil {
		t.Fatalf("compile plugin node: %v", err)
	}
	result, err := effect.New(effect.Services{}).WithExtensions(registered).Execute(context.Background(), node, trigger.Event{RoomID: 3, PlayerID: 9})
	if err != nil || result.Status != effect.Applied || !result.ResetTimers {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if _, err = compiler.CompileNode(record.Config{ItemID: 8, RoomID: 3, Interaction: "wf_act_reset_timers"}); err != nil {
		t.Fatalf("native descriptor regressed: %v", err)
	}
}

// TestConditionAndPanicIsolation verifies condition results and scope disabling.
func TestConditionAndPanicIsolation(t *testing.T) {
	registered := NewRegistry(time.Second, nil)
	scope := pluginruntime.NewScope("demo")
	access := NewAccess(registered, scope)
	if err := access.RegisterCondition("plugin.demo.condition", func(context.Context, sdkwired.Node, sdkwired.Event) (bool, bool, error) {
		return true, true, nil
	}); err != nil {
		t.Fatal(err)
	}
	node := &configuration.Node{Descriptor: nativeregistry.Descriptor{Key: "plugin.demo.condition"}}
	result, found, err := registered.EvaluateCondition(node, condition.Context{CallbackContext: context.Background()})
	if err != nil || !found || !result.Pass || !result.Valid {
		t.Fatalf("result=%#v found=%v err=%v", result, found, err)
	}
	if err = access.RegisterEffect("plugin.demo.panic", func(context.Context, sdkwired.Node, sdkwired.Event) (sdkwired.EffectResult, error) {
		panic("boom")
	}); err != nil {
		t.Fatal(err)
	}
	panicNode := &configuration.Node{Descriptor: nativeregistry.Descriptor{Key: "plugin.demo.panic"}}
	_, found, err = registered.ExecuteEffect(context.Background(), panicNode, trigger.Event{})
	if !found || !errors.Is(err, pluginruntime.ErrCallbackPanic) || scope.Enabled() {
		t.Fatalf("found=%v enabled=%v err=%v", found, scope.Enabled(), err)
	}
}
