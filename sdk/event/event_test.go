package event

import (
	"testing"
	"time"

	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

// TestConcreteEventsExposeTypedState verifies cancellable and notification contracts.
func TestConcreteEventsExposeTypedState(t *testing.T) {
	player := sdkplayer.Player{ID: 7, Username: "alice", Online: true}
	chat := NewChatSend(player, 3, "hello")
	if chat.Name() != ChatSendName || chat.Cancelled() || chat.Text != "hello" {
		t.Fatalf("unexpected chat event: %+v", chat)
	}
	chat.SetCancelled(true)
	if !chat.Cancelled() {
		t.Fatal("expected cancellable event state")
	}
	connected := &PlayerConnected{Player: player}
	if connected.Name() != PlayerConnectedName {
		t.Fatalf("unexpected connected name %q", connected.Name())
	}
}

// TestMutableEventsSatisfyTheExplicitCloneContract verifies every pre-commit event is self-cloning.
func TestMutableEventsSatisfyTheExplicitCloneContract(t *testing.T) {
	events := []Mutable{
		NewChatSend(sdkplayer.Player{}, 1, "chat"),
		NewCurrencyGrant(sdkplayer.Player{}, 5, 10, "plugin"),
		&RoomEnterAttempt{},
		&RoomUpdate{},
		&RoomUnitMove{},
		&SanctionApply{},
		&RoomModerationAction{},
		&TradeStart{},
		&TradeConfirm{},
		&TradeCancel{},
		&FurniturePlace{},
		&CatalogPurchase{},
	}
	for _, event := range events {
		cloned := event.Clone()
		if cloned == event {
			t.Fatalf("%s returned its original mutable value", event.Name())
		}
		cancellable := cloned.(Cancellable)
		cancellable.SetCancelled(true)
		cloned.Apply(event)
		if !event.(Cancellable).Cancelled() {
			t.Fatalf("%s did not apply its cancellation state", event.Name())
		}
	}
}

// TestImmutableEventsCloneTheirSnapshots verifies every notification owns its callback copy.
func TestImmutableEventsCloneTheirSnapshots(t *testing.T) {
	events := []interface {
		Event
		CloneEvent() Event
	}{
		&CurrencyChanged{},
		&CommandAttempt{},
		&PlayerConnected{},
		NewPublished("room.created", map[string]any{"RoomID": int64(7)}),
	}
	for _, event := range events {
		if event.Name() == "" {
			t.Fatalf("empty event name for %T", event)
		}
		if cloned := event.CloneEvent(); cloned == event || cloned.Name() != event.Name() {
			t.Fatalf("invalid clone for %T: %#v", event, cloned)
		}
	}
}

// TestRoomUpdateCloneOwnsNestedValues verifies one listener cannot mutate another callback snapshot.
func TestRoomUpdateCloneOwnsNestedValues(t *testing.T) {
	tags := []string{"one"}
	category := int64(7)
	categoryPointer := &category
	name := "room"
	event := &RoomUpdate{Params: RoomUpdateParams{Name: &name, Tags: &tags, CategoryID: &categoryPointer}}
	cloned := event.Clone().(*RoomUpdate)
	(*cloned.Params.Tags)[0] = "changed"
	**cloned.Params.CategoryID = 8
	*cloned.Params.Name = "changed"
	if *event.Params.Name != "room" || (*event.Params.Tags)[0] != "one" || **event.Params.CategoryID != 7 {
		t.Fatalf("original changed: %#v", event.Params)
	}
}

// TestPublishedReturnsDetachedPayloads verifies post-commit notifications stay immutable.
func TestPublishedReturnsDetachedPayloads(t *testing.T) {
	now := time.Now()
	event := NewPublished("room.created", map[string]any{
		"Room": map[string]any{"Tags": []any{"one"}, "At": now},
	})
	fields := event.Fields()
	fields["Room"].(map[string]any)["Tags"].([]any)[0] = "changed"
	value, found := event.Field("Room")
	if !found || value.(map[string]any)["Tags"].([]any)[0] != "one" {
		t.Fatalf("published payload leaked mutation: %#v", value)
	}
}
