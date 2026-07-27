package bypass

import (
	"context"
	"strings"
	"testing"
	"time"

	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	catalogpricing "github.com/niflaot/pixels/internal/realm/catalog/pricing"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
	"go.uber.org/zap"
)

// playerAccess records permission decisions and command feedback.
type playerAccess struct {
	// allowed controls the permission result.
	allowed bool
	// message stores the latest command reply.
	message string
}

// Message records one player-facing command reply.
func (access *playerAccess) Message(_ int64, message string) error {
	access.message = message

	return nil
}

// HasPermission returns the configured permission decision.
func (access *playerAccess) HasPermission(int64, string) (bool, error) {
	return access.allowed, nil
}

// TestCommandRequiresPermissionAndTogglesPricing verifies the complete command path.
func TestCommandRequiresPermissionAndTogglesPricing(t *testing.T) {
	state := catalogpricing.New(false)
	players := &playerAccess{}
	tree := plugincommand.NewTree(":", time.Second, nil, zap.NewNop())
	tree.SetPlayers(players)
	if err := Register(tree, state, nil); err != nil {
		t.Fatalf("register: %v", err)
	}
	player := sdkplayer.Player{ID: 7, Username: "admin", Online: true}
	handled, err := tree.Execute(context.Background(), player, ":catalogbypass")
	if err != nil || !handled || state.FreeItems() || !strings.Contains(players.message, "permiso") {
		t.Fatalf("denied handled=%v free=%v message=%q error=%v", handled, state.FreeItems(), players.message, err)
	}
	players.allowed = true
	handled, err = tree.Execute(context.Background(), player, ":catalogbypass")
	if err != nil || !handled || !state.FreeItems() || !strings.Contains(players.message, "activado") {
		t.Fatalf("enabled handled=%v free=%v message=%q error=%v", handled, state.FreeItems(), players.message, err)
	}
	handled, err = tree.Execute(context.Background(), player, ":catalogbypass")
	if err != nil || !handled || state.FreeItems() || !strings.Contains(players.message, "desactivado") {
		t.Fatalf("disabled handled=%v free=%v message=%q error=%v", handled, state.FreeItems(), players.message, err)
	}
}
