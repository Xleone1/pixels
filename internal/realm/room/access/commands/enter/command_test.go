package enter

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	outentryerror "github.com/niflaot/pixels/networking/outbound/room/entryerror"
	outscore "github.com/niflaot/pixels/networking/outbound/room/score"
	outdesktop "github.com/niflaot/pixels/networking/outbound/session/desktop"
	outerror "github.com/niflaot/pixels/networking/outbound/session/error"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// voteReaderForTest returns fixed entry vote state.
type voteReaderForTest struct{}

// enterEventsForTest records and vetoes one room entry.
type enterEventsForTest struct {
	// calls stores dispatch count.
	calls int
	// currentRoomID stores the source room snapshot.
	currentRoomID int64
}

// DispatchRoomEnterAttempt vetoes one authorized room entry.
func (events *enterEventsForTest) DispatchRoomEnterAttempt(_ context.Context, player sdkplayer.Player, roomID int64, _ bool) bool {
	if player.ID == 7 && roomID == 9 {
		events.calls++
		events.currentRoomID = player.RoomID
	}
	return true
}

// State returns a current score and eligible player.
func (voteReaderForTest) State(context.Context, int64, int64) (roomvotes.State, error) {
	return roomvotes.State{Score: 12, CanVote: true}, nil
}

// List returns no durable votes.
func (voteReaderForTest) List(context.Context, roomvotes.Query) ([]roomvotes.Vote, error) {
	return nil, nil
}

// TestCommandName verifies the stable command name.
func TestCommandName(t *testing.T) {
	if (Command{}).CommandName() != Name {
		t.Fatalf("unexpected command name %s", (Command{}).CommandName())
	}
}

// TestCommandLoggingRedactsPassword verifies command diagnostics never expose plaintext.
func TestCommandLoggingRedactsPassword(t *testing.T) {
	var output bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	logger := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(&output), zap.DebugLevel))
	logger.Debug("command", zap.Object("command", Command{RoomID: 9, Password: "private-room-password", Trusted: true}))
	logged := output.String()
	if strings.Contains(logged, "private-room-password") || !strings.Contains(logged, "password_provided") {
		t.Fatalf("unexpected command log %s", logged)
	}
}

// TestSendEntryErrorUsesNitroProtocol verifies password and room rejection packets.
func TestSendEntryErrorUsesNitroProtocol(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	handler := Handler{}
	if err := handler.sendEntryError(context.Background(), connection, roomentry.ErrWrongPassword); err != nil {
		t.Fatalf("send wrong password: %v", err)
	}
	if err := handler.sendEntryError(context.Background(), connection, roomentry.ErrAccessDenied); err != nil {
		t.Fatalf("send access denied: %v", err)
	}
	if err := handler.sendEntryError(context.Background(), connection, roomentry.ErrEntryLocked); err != nil {
		t.Fatalf("send entry locked: %v", err)
	}
	if len(*sent) != 3 || (*sent)[0].Header != outerror.Header || (*sent)[1].Header != outentryerror.Header || (*sent)[2].Header != outdesktop.Header {
		t.Fatalf("unexpected packets %#v", *sent)
	}
}

// TestLoadRoomPropagatesStoreErrors verifies persistence errors.
func TestLoadRoomPropagatesStoreErrors(t *testing.T) {
	storeErr := errors.New("store failed")
	handler := Handler{Rooms: roomManagerForTest{err: storeErr}}
	_, _, err := handler.loadRoom(context.Background(), 9)
	if !errors.Is(err, storeErr) {
		t.Fatalf("expected store error, got %v", err)
	}

	handler = Handler{
		Rooms:   roomManagerForTest{room: roomForTest(), found: true},
		Layouts: layoutManagerForTest{err: storeErr},
	}
	_, _, err = handler.loadRoom(context.Background(), 9)
	if !errors.Is(err, storeErr) {
		t.Fatalf("expected layout error, got %v", err)
	}
}

// TestJoinAllowsMissingEventBus verifies publish is optional.
func TestJoinAllowsMissingEventBus(t *testing.T) {
	handler := Handler{Runtime: roomlive.NewRegistry(nil)}
	_, err := handler.join(context.Background(), playerForTest(t), connectionForTest(), roomForTest(), layoutForTest())
	if err != nil {
		t.Fatalf("join without events: %v", err)
	}
}

// TestCommandEnvelopeValid verifies command envelope naming.
func TestCommandEnvelopeValid(t *testing.T) {
	envelope := command.Envelope[Command]{Command: Command{RoomID: 9}}
	if !envelope.Valid() {
		t.Fatal("expected valid command envelope")
	}
}

// TestPluginRoomEntryVetoesBeforeRuntimeMutation verifies admission interception ordering.
func TestPluginRoomEntryVetoesBeforeRuntimeMutation(t *testing.T) {
	for _, previousRoomID := range []int64{0, 3} {
		t.Run(string(rune('0'+previousRoomID)), func(t *testing.T) {
			player := playerForTest(t)
			connection, sent := sessionConnectionForTest(t)
			events := &enterEventsForTest{}
			runtime := roomlive.NewRegistry(nil)
			if previousRoomID > 0 {
				if _, err := runtime.Activate(roomlive.Snapshot{ID: previousRoomID, MaxUsers: 5}); err != nil {
					t.Fatal(err)
				}
				if _, err := runtime.Join(context.Background(), previousRoomID, occupantForTest(7)); err != nil {
					t.Fatal(err)
				}
				if err := player.EnterRoom(previousRoomID); err != nil {
					t.Fatal(err)
				}
			}
			handler := Handler{
				Players: playerRegistryForTest(t, player), Bindings: bindingRegistryForTest(t, 7),
				Rooms:   roomManagerForTest{room: roomForTest(), found: true},
				Layouts: layoutManagerForTest{roomLayout: layoutForTest(), found: true},
				Runtime: runtime, PluginEvents: events,
			}
			err := handler.Handle(context.Background(), command.Envelope[Command]{Command: Command{Handler: connection, RoomID: 9}})
			currentRoomID, inRoom := player.CurrentRoom()
			if err != nil || events.calls != 1 || events.currentRoomID != previousRoomID ||
				inRoom != (previousRoomID > 0) || currentRoomID != previousRoomID {
				t.Fatalf("current=%d eventCurrent=%d inRoom=%v calls=%d err=%v", currentRoomID, events.currentRoomID, inRoom, events.calls, err)
			}
			if _, found := handler.Runtime.Find(9); found || len(*sent) != 1 {
				t.Fatalf("room activated=%v packets=%#v", found, *sent)
			}
		})
	}
}

// TestSendVoteStateBootstrapsRoomScore verifies entry score projection.
func TestSendVoteStateBootstrapsRoomScore(t *testing.T) {
	connection, sent := sessionConnectionForTest(t)
	if err := (Handler{Votes: voteReaderForTest{}}).sendVoteState(context.Background(), connection, 9, 2); err != nil {
		t.Fatalf("send vote state: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outscore.Header {
		t.Fatalf("packets=%+v", *sent)
	}
	values, err := codec.DecodePacketExact((*sent)[0], outscore.Definition)
	if err != nil || values[0].Int32 != 12 || !values[1].Boolean {
		t.Fatalf("values=%+v err=%v", values, err)
	}
}
