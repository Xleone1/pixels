package tests

import (
	"context"
	"testing"

	botbehavior "github.com/niflaot/pixels/internal/realm/bot/behavior"
	bottalked "github.com/niflaot/pixels/internal/realm/bot/events/talked"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	"github.com/niflaot/pixels/pkg/bus"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// speechDispatcher controls bot-speech plugin outcomes.
type speechDispatcher struct {
	// message stores the replacement text.
	message string
	// cancelled reports whether delivery must stop.
	cancelled bool
}

// DispatchBotSpeech returns the configured plugin outcome.
func (dispatcher *speechDispatcher) DispatchBotSpeech(context.Context, int64, int64, string, string) (string, bool) {
	return dispatcher.message, dispatcher.cancelled
}

// TestBotSpeechMutationAndCancellationReachDelivery verifies the persistence boundary.
func TestBotSpeechMutationAndCancellationReachDelivery(t *testing.T) {
	behaviors := botbehavior.NewRegistry()
	if err := botbehavior.RegisterBuiltins(behaviors); err != nil {
		t.Fatalf("register behaviors: %v", err)
	}
	events := bus.New()
	t.Cleanup(func() { _ = events.Close() })
	delivered := make([]string, 0, 1)
	if _, err := events.Subscribe(bottalked.Name, bus.PriorityNormal, func(_ context.Context, event bus.Event) error {
		payload, ok := event.Payload.(bottalked.Payload)
		if !ok {
			t.Fatalf("unexpected payload %#v", event.Payload)
		}
		delivered = append(delivered, payload.Message)
		return nil
	}); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	service, room := serviceFixtureWithStore(t, behaviors, &roomStore{bots: []botrecord.Bot{placedBot(1, botrecord.BehaviorGeneric, false)}}, events)
	if err := service.EnsureRoom(context.Background(), room); err != nil {
		t.Fatalf("ensure room: %v", err)
	}
	dispatcher := &speechDispatcher{message: "rewritten"}
	service.SetPluginRuntime(dispatcher)
	view := sdkbot.Bot{ID: 1, OwnerPlayerID: 1, RoomID: room.ID(), BehaviorType: botrecord.BehaviorGeneric}
	if err := service.Talk(context.Background(), view, "original", sdkbot.ScopeTalk, 0); err != nil {
		t.Fatalf("talk: %v", err)
	}
	if len(delivered) != 1 || delivered[0] != "rewritten" {
		t.Fatalf("delivered=%v", delivered)
	}
	dispatcher.cancelled = true
	if err := service.Talk(context.Background(), view, "blocked", sdkbot.ScopeTalk, 0); err != nil {
		t.Fatalf("cancelled talk: %v", err)
	}
	if len(delivered) != 1 {
		t.Fatalf("cancelled speech was delivered: %v", delivered)
	}
}
