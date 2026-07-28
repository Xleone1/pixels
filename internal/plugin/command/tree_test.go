package command

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	sdkcommand "github.com/niflaot/pixels/sdk/command"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.minekube.com/brigodier"
	"go.uber.org/zap"
)

// fakePlayerAccess supplies command permission and feedback behavior.
type fakePlayerAccess struct {
	// mutex protects feedback for asynchronous callback tests.
	mutex sync.Mutex
	// allowed controls permission resolution.
	allowed bool
	// messages stores replies in delivery order.
	messages []string
}

// attemptRecord stores one observed command submission.
type attemptRecord struct {
	// input stores command text without the prefix.
	input string
	// root stores the first command token.
	root string
}

// attemptRecorder records command-prefixed chat before command resolution.
type attemptRecorder struct {
	// records stores observed attempts in order.
	records []attemptRecord
}

// DispatchCommandAttempt records one command submission.
func (recorder *attemptRecorder) DispatchCommandAttempt(_ context.Context, _ sdkplugin.Player, input string, root string) {
	recorder.records = append(recorder.records, attemptRecord{input: input, root: root})
}

// Message records one player reply.
func (access *fakePlayerAccess) Message(_ int64, message string) error {
	access.mutex.Lock()
	defer access.mutex.Unlock()
	access.messages = append(access.messages, message)
	return nil
}

// HasPermission returns the configured permission decision.
func (access *fakePlayerAccess) HasPermission(_ int64, _ string) (bool, error) {
	return access.allowed, nil
}

// lastMessage returns the latest recorded reply.
func (access *fakePlayerAccess) lastMessage() string {
	access.mutex.Lock()
	defer access.mutex.Unlock()
	if len(access.messages) == 0 {
		return ""
	}
	return access.messages[len(access.messages)-1]
}

// testPlayer creates one stable plugin-facing sender.
func testPlayer() sdkplayer.Player {
	return sdkplayer.Player{ID: 7, Username: "demo", RoomID: 3, Online: true}
}

// TestTreeRejectsCollision verifies exclusive root ownership.
func TestTreeRejectsCollision(t *testing.T) {
	tree := NewTree(":", time.Second, nil, zap.NewNop())
	if err := tree.register(pluginruntime.NewScope("one"), brigodier.Literal("hello")); err != nil {
		t.Fatalf("register first: %v", err)
	}
	if err := tree.register(pluginruntime.NewScope("two"), brigodier.Literal("hello")); !errors.Is(err, ErrCommandExists) {
		t.Fatalf("expected collision, got %v", err)
	}
}

// TestTreeConsumesPermissionGatedCommands verifies permissions and feedback.
func TestTreeConsumesPermissionGatedCommands(t *testing.T) {
	players := &fakePlayerAccess{}
	tree := NewTree(":", time.Second, nil, zap.NewNop())
	tree.SetPlayers(players)
	executed := false
	root := brigodier.Literal("hello").Requires(sdkcommand.RequiresPermission("plugin.test.hello")).Executes(brigodier.CommandFunc(func(*brigodier.CommandContext) error {
		executed = true
		return nil
	}))
	if err := tree.register(pluginruntime.NewScope("test"), root); err != nil {
		t.Fatalf("register command: %v", err)
	}
	handled, err := tree.Execute(context.Background(), testPlayer(), ":hello")
	if err != nil || !handled || executed || players.lastMessage() == "" {
		t.Fatalf("expected consumed denial, handled=%v executed=%v message=%q err=%v", handled, executed, players.lastMessage(), err)
	}
	players.allowed = true
	handled, err = tree.Execute(context.Background(), testPlayer(), ":hello")
	if err != nil || !handled || !executed {
		t.Fatalf("expected execution, handled=%v executed=%v err=%v", handled, executed, err)
	}
}

// TestTreeHonorsConfiguredPrefix verifies the environment-controlled cutover.
func TestTreeHonorsConfiguredPrefix(t *testing.T) {
	tree := NewTree("!", time.Second, nil, zap.NewNop())
	tree.SetPlayers(&fakePlayerAccess{allowed: true})
	_ = tree.register(pluginruntime.NewScope("test"), brigodier.Literal("hello").Executes(brigodier.CommandFunc(func(*brigodier.CommandContext) error { return nil })))
	handled, err := tree.Execute(context.Background(), testPlayer(), ":hello")
	if err != nil || handled {
		t.Fatalf("expected old prefix to remain chat, handled=%v err=%v", handled, err)
	}
	handled, err = tree.Execute(context.Background(), testPlayer(), "!hello")
	if err != nil || !handled {
		t.Fatalf("expected configured prefix consumption, handled=%v err=%v", handled, err)
	}
}

// TestTreeDetectsEveryPrefixedCommandAttempt verifies detection precedes parsing and ownership.
func TestTreeDetectsEveryPrefixedCommandAttempt(t *testing.T) {
	tree := NewTree(":", time.Second, nil, zap.NewNop())
	tree.SetPlayers(&fakePlayerAccess{allowed: true})
	recorder := &attemptRecorder{}
	tree.SetAttemptDispatcher(recorder)

	for _, input := range []string{"hello", ":", ":missing argument", ":hello value"} {
		_, _ = tree.Execute(context.Background(), testPlayer(), input)
	}

	expected := []attemptRecord{
		{input: "", root: ""},
		{input: "missing argument", root: "missing"},
		{input: "hello value", root: "hello"},
	}
	if len(recorder.records) != len(expected) {
		t.Fatalf("unexpected attempts %#v", recorder.records)
	}
	for index := range expected {
		if recorder.records[index] != expected[index] {
			t.Fatalf("attempt %d = %#v, expected %#v", index, recorder.records[index], expected[index])
		}
	}
}

// TestTreeDisablesPanickingPlugin verifies command panic isolation.
func TestTreeDisablesPanickingPlugin(t *testing.T) {
	tree := NewTree(":", time.Second, nil, zap.NewNop())
	players := &fakePlayerAccess{allowed: true}
	tree.SetPlayers(players)
	scope := pluginruntime.NewScope("broken")
	_ = tree.register(scope, brigodier.Literal("panic").Executes(brigodier.CommandFunc(func(*brigodier.CommandContext) error { panic("boom") })))
	handled, err := tree.Execute(context.Background(), testPlayer(), ":panic")
	if err != nil || !handled || scope.Enabled() || players.lastMessage() == "" {
		t.Fatalf("expected isolated panic feedback, handled=%v enabled=%v err=%v", handled, scope.Enabled(), err)
	}
}

// TestTreeReportsEmptyUnknownAndDisabledCommands verifies non-executable input feedback.
func TestTreeReportsEmptyUnknownAndDisabledCommands(t *testing.T) {
	players := &fakePlayerAccess{allowed: true}
	tree := NewTree(":", time.Second, nil, zap.NewNop())
	tree.SetPlayers(players)
	for _, input := range []string{":", ":missing"} {
		handled, err := tree.Execute(context.Background(), testPlayer(), input)
		if err != nil || !handled || players.lastMessage() == "" {
			t.Fatalf("expected feedback for %q, handled=%v err=%v", input, handled, err)
		}
	}
	scope := pluginruntime.NewScope("disabled")
	_ = tree.register(scope, brigodier.Literal("off").Executes(brigodier.CommandFunc(func(*brigodier.CommandContext) error { return nil })))
	scope.Disable()
	handled, err := tree.Execute(context.Background(), testPlayer(), ":off")
	if err != nil || !handled {
		t.Fatalf("expected disabled command feedback, handled=%v err=%v", handled, err)
	}
}

// TestTreeRejectsInvalidRegistrations verifies malformed roots never enter Brigadier.
func TestTreeRejectsInvalidRegistrations(t *testing.T) {
	tree := NewTree(":", time.Second, nil, zap.NewNop())
	if err := tree.register(nil, brigodier.Literal("hello")); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("expected missing scope rejection, got %v", err)
	}
	if err := tree.register(pluginruntime.NewScope("test"), nil); !errors.Is(err, ErrInvalidCommand) {
		t.Fatalf("expected missing command rejection, got %v", err)
	}
}

// TestTreeReportsMalformedArguments verifies Brigadier parse failures are consumed.
func TestTreeReportsMalformedArguments(t *testing.T) {
	players := &fakePlayerAccess{allowed: true}
	tree := NewTree(":", time.Second, nil, zap.NewNop())
	tree.SetPlayers(players)
	root := brigodier.Literal("give").Then(brigodier.Argument("count", brigodier.Int).Executes(brigodier.CommandFunc(func(*brigodier.CommandContext) error { return nil })))
	_ = tree.register(pluginruntime.NewScope("test"), root)
	handled, err := tree.Execute(context.Background(), testPlayer(), ":give nope")
	if err != nil || !handled || players.lastMessage() == "" {
		t.Fatalf("expected malformed feedback, handled=%v message=%q err=%v", handled, players.lastMessage(), err)
	}
}
