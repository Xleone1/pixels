package admin

import (
	"context"
	"strings"
	"testing"
	"time"

	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	"github.com/niflaot/pixels/internal/plugin/loader"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	"github.com/niflaot/pixels/networking/codec"
	outnotification "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/build"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
	"go.uber.org/zap"
)

// commandPlayerAccess records tree feedback and permission decisions.
type commandPlayerAccess struct {
	// allowed controls command permission.
	allowed bool
	// message stores the latest feedback.
	message string
}

// commandEffectManager records command-driven effect selections.
type commandEffectManager struct {
	// enabled stores requested player and effect ids.
	enabled [][2]int64
}

// List returns no effects for command tests.
func (*commandEffectManager) List(context.Context, int64) ([]playereffect.Effect, error) {
	return nil, nil
}

// Grant returns an unused empty effect.
func (*commandEffectManager) Grant(context.Context, int64, int32, int32, playereffect.Source) (playereffect.Effect, error) {
	return playereffect.Effect{}, nil
}

// GrantEnabled returns an unused empty effect.
func (*commandEffectManager) GrantEnabled(context.Context, int64, int32, int32, playereffect.Source) (playereffect.Effect, error) {
	return playereffect.Effect{}, nil
}

// Enable records one selected effect.
func (manager *commandEffectManager) Enable(_ context.Context, playerID int64, effectID int32) error {
	manager.enabled = append(manager.enabled, [2]int64{playerID, int64(effectID)})
	return nil
}

// Activate returns an unused empty effect.
func (*commandEffectManager) Activate(context.Context, int64, int32) (playereffect.Effect, error) {
	return playereffect.Effect{}, nil
}

// Revoke accepts an unused effect revocation.
func (*commandEffectManager) Revoke(context.Context, int64, int32) error {
	return nil
}

// assertEffectSound verifies one sound-only confirmation packet.
func assertEffectSound(t *testing.T, packets []codec.Packet) {
	t.Helper()
	if len(packets) != 1 || packets[0].Header != outnotification.Header {
		t.Fatalf("unexpected confirmation packets %#v", packets)
	}
	definition := codec.Definition{
		codec.StringField, codec.Int32Field,
		codec.StringField, codec.StringField,
		codec.StringField, codec.StringField,
		codec.StringField, codec.StringField,
	}
	values, err := codec.DecodePacketExact(packets[0], definition)
	if err != nil || values[1].Int32 != 3 || values[4].String != "display" || values[5].String != "SOUND" ||
		values[6].String != "sound" || values[7].String != effectConfirmationSound {
		t.Fatalf("unexpected confirmation payload %#v err=%v", values, err)
	}
}

// TestRegisteredCommandsExecuteEveryCorePath verifies the complete command tree.
func TestRegisteredCommandsExecuteEveryCorePath(t *testing.T) {
	fixture := newServiceFixture()
	packets := make([]codec.Packet, 0)
	addServicePlayer(t, fixture, 7, "admin", nil, &packets)
	addServicePlayer(t, fixture, 8, "target", nil, &packets)
	access := &commandPlayerAccess{allowed: true}
	tree := plugincommand.NewTree(":", time.Second, nil, zap.NewNop())
	tree.SetPlayers(access)
	if err := RegisterCommands(tree, fixture.service); err != nil {
		t.Fatalf("register commands: %v", err)
	}
	player := sdkplayer.Player{ID: 7, Username: "admin", Online: true}
	for _, command := range []string{":alert target Please behave", ":halert Maintenance soon", ":trace", ":trace"} {
		handled, err := tree.Execute(context.Background(), player, command)
		if err != nil || !handled {
			t.Fatalf("command=%q handled=%v err=%v", command, handled, err)
		}
	}
	if len(packets) != 2 || fixture.trace.activates != 1 || fixture.trace.finalizes != 1 {
		t.Fatalf("packets=%d activates=%d finalizes=%d", len(packets), fixture.trace.activates, fixture.trace.finalizes)
	}
	values, err := codec.DecodePacketExact(packets[0], codec.Definition{codec.Named("message", codec.StringField)})
	if err != nil || len(values) != 1 || values[0].String != "Please behave" {
		t.Fatalf("direct alert payload=%#v err=%v", values, err)
	}
}

// Message records command feedback.
func (access *commandPlayerAccess) Message(_ int64, message string) error {
	access.message = message
	return nil
}

// HasPermission returns the configured decision.
func (access *commandPlayerAccess) HasPermission(int64, string) (bool, error) {
	return access.allowed, nil
}

// TestRegisterCommandsExecutesPermissionGatedAbout verifies core tree integration.
func TestRegisterCommandsExecutesPermissionGatedAbout(t *testing.T) {
	players := &commandPlayerAccess{}
	tree := plugincommand.NewTree(":", time.Second, nil, zap.NewNop())
	tree.SetPlayers(players)
	service := &Service{build: build.NewInfo("pixels", "v0.0.3", "123456789"), plugins: &loader.Loader{}, log: zap.NewNop()}
	if err := RegisterCommands(tree, service); err != nil {
		t.Fatalf("register commands: %v", err)
	}
	player := sdkplayer.Player{ID: 7, Username: "admin", Online: true}
	handled, err := tree.Execute(context.Background(), player, ":about")
	if err != nil || !handled || !strings.Contains(players.message, "permiso") {
		t.Fatalf("denied handled=%v message=%q err=%v", handled, players.message, err)
	}
	players.allowed = true
	handled, err = tree.Execute(context.Background(), player, ":about")
	if err != nil || !handled || !strings.Contains(players.message, "v0.0.3") {
		t.Fatalf("allowed handled=%v message=%q err=%v", handled, players.message, err)
	}
}

// TestEffectCommandAllowsStaffSelection verifies permitted owned-effect selection.
func TestEffectCommandAllowsStaffSelection(t *testing.T) {
	fixture := newServiceFixture()
	effects := &commandEffectManager{}
	fixture.service.effects = effects
	packets := make([]codec.Packet, 0)
	addServicePlayer(t, fixture, 7, "admin", nil, &packets)
	players := &commandPlayerAccess{allowed: true}
	tree := plugincommand.NewTree(":", time.Second, nil, zap.NewNop())
	tree.SetPlayers(players)
	if err := RegisterCommands(tree, fixture.service); err != nil {
		t.Fatal(err)
	}
	handled, err := tree.Execute(context.Background(), sdkplayer.Player{ID: 7, Username: "admin", Online: true}, ":effect 90")
	if err != nil || !handled || len(effects.enabled) != 1 || effects.enabled[0] != [2]int64{7, 90} || players.message != "" {
		t.Fatalf("handled=%v enabled=%v message=%q err=%v", handled, effects.enabled, players.message, err)
	}
	assertEffectSound(t, packets)
}

// TestEffectCommandLimitsUnpermittedPlayersToClear verifies the environment exception.
func TestEffectCommandLimitsUnpermittedPlayersToClear(t *testing.T) {
	fixture := newServiceFixture()
	fixture.service.config.AllowUnpermittedEffectClear = true
	effects := &commandEffectManager{}
	fixture.service.effects = effects
	packets := make([]codec.Packet, 0)
	addServicePlayer(t, fixture, 7, "member", nil, &packets)
	players := &commandPlayerAccess{}
	tree := plugincommand.NewTree(":", time.Second, nil, zap.NewNop())
	tree.SetPlayers(players)
	if err := RegisterCommands(tree, fixture.service); err != nil {
		t.Fatal(err)
	}
	player := sdkplayer.Player{ID: 7, Username: "member", Online: true}
	if handled, err := tree.Execute(context.Background(), player, ":effect 90"); err != nil || !handled || len(effects.enabled) != 0 || !strings.Contains(players.message, "permiso") {
		t.Fatalf("nonzero handled=%v enabled=%v message=%q err=%v", handled, effects.enabled, players.message, err)
	}
	players.message = ""
	if handled, err := tree.Execute(context.Background(), player, ":effect 0"); err != nil || !handled || len(effects.enabled) != 1 || effects.enabled[0] != [2]int64{7, 0} || players.message != "" {
		t.Fatalf("clear handled=%v enabled=%v message=%q err=%v", handled, effects.enabled, players.message, err)
	}
	assertEffectSound(t, packets)
	fixture.service.config.AllowUnpermittedEffectClear = false
	if handled, err := tree.Execute(context.Background(), player, ":effect 0"); err != nil || !handled || len(effects.enabled) != 1 || !strings.Contains(players.message, "permiso") {
		t.Fatalf("disabled clear handled=%v enabled=%v message=%q err=%v", handled, effects.enabled, players.message, err)
	}
}
