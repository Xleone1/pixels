package routes

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outinfo "github.com/niflaot/pixels/networking/outbound/user/info"
)

// routeIdentityStore records administrative identity behavior.
type routeIdentityStore struct {
	// manager stores the mutable player fixture.
	manager *fakeManager
	// changes stores durable audit fixtures.
	changes []playeridentity.NameChange
}

// Rename applies one self-service fixture rename.
func (store *routeIdentityStore) Rename(_ context.Context, playerID int64, username string) (playeridentity.RenameResult, error) {
	return store.rename(playerID, username, playerID, "self-service", "client"), nil
}

// RenameAdmin applies one attributed fixture rename.
func (store *routeIdentityStore) RenameAdmin(_ context.Context, playerID int64, username string, actorPlayerID int64, reason string) (playeridentity.RenameResult, error) {
	return store.rename(playerID, username, actorPlayerID, reason, "api"), nil
}

// NameChanges returns bounded fixture history.
func (store *routeIdentityStore) NameChanges(_ context.Context, _ int64, limit int) ([]playeridentity.NameChange, error) {
	if limit < len(store.changes) {
		return store.changes[:limit], nil
	}
	return store.changes, nil
}

// SetAuthorization replaces the fixture self-service rename policy.
func (store *routeIdentityStore) SetAuthorization(_ context.Context, _ int64, allowed bool, _ int64, _ string) error {
	store.manager.record.Profile.AllowNameChange = allowed
	return nil
}

// rename mutates the fixture and appends one audit entry.
func (store *routeIdentityStore) rename(playerID int64, username string, actorPlayerID int64, reason string, source string) playeridentity.RenameResult {
	oldUsername := store.manager.record.Player.Username
	store.manager.record.Player.Username = username
	store.manager.record.Profile.AllowNameChange = false
	change := playeridentity.NameChange{ID: int64(len(store.changes) + 1), PlayerID: playerID,
		OldUsername: oldUsername, NewUsername: username, ActorPlayerID: actorPlayerID,
		Reason: reason, Source: source, ChangedAt: time.Now().UTC()}
	store.changes = append([]playeridentity.NameChange{change}, store.changes...)
	return playeridentity.RenameResult{OldUsername: oldUsername, NewUsername: username}
}

// TestNameChangeAdministrationManagesPolicyRenameAndHistory verifies the complete HTTP workflow.
func TestNameChangeAdministrationManagesPolicyRenameAndHistory(t *testing.T) {
	app, manager, _ := nameChangeApplication(t, false)
	response := nameChangeRequest(t, app, http.MethodPost, "/api/admin/players/7/name-change/allow", `{"actorPlayerId":7,"reason":"onboarding"}`)
	if response.StatusCode != fiber.StatusOK || !manager.record.Profile.AllowNameChange {
		t.Fatalf("allow status=%d allowed=%t", response.StatusCode, manager.record.Profile.AllowNameChange)
	}
	response = nameChangeRequest(t, app, http.MethodPut, "/api/admin/players/7/name-change/authorization", `{"allowed":false,"actorPlayerId":7,"reason":"revoked"}`)
	if response.StatusCode != fiber.StatusOK || manager.record.Profile.AllowNameChange {
		t.Fatalf("revoke status=%d allowed=%t", response.StatusCode, manager.record.Profile.AllowNameChange)
	}
	response = nameChangeRequest(t, app, http.MethodPost, "/api/admin/players/7/name-change", `{"username":"renamed","actorPlayerId":7,"reason":"support"}`)
	if response.StatusCode != fiber.StatusOK || manager.record.Player.Username != "renamed" {
		t.Fatalf("rename status=%d username=%q", response.StatusCode, manager.record.Player.Username)
	}
	response = nameChangeRequest(t, app, http.MethodGet, "/api/admin/players/7/name-changes?limit=10", "")
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != fiber.StatusOK || !strings.Contains(string(body), `"newUsername":"renamed"`) || !strings.Contains(string(body), `"source":"api"`) {
		t.Fatalf("history status=%d body=%s", response.StatusCode, body)
	}
}

// TestNameChangeAuthorizationProjectsUserInfo verifies online clients update without reconnecting.
func TestNameChangeAuthorizationProjectsUserInfo(t *testing.T) {
	app, manager, packets := nameChangeApplication(t, true)
	response := nameChangeRequest(t, app, http.MethodPost, "/api/admin/players/7/name-change/allow", `{"actorPlayerId":7,"reason":"onboarding"}`)
	if response.StatusCode != fiber.StatusOK || !manager.record.Profile.AllowNameChange {
		t.Fatalf("status=%d allowed=%t", response.StatusCode, manager.record.Profile.AllowNameChange)
	}
	if len(*packets) != 1 || (*packets)[0].Header != outinfo.Header {
		t.Fatalf("packets=%#v", *packets)
	}
}

// TestNameChangeRoutesRejectInvalidInput verifies stable administrative validation.
func TestNameChangeRoutesRejectInvalidInput(t *testing.T) {
	app, _, _ := nameChangeApplication(t, false)
	for _, request := range []*http.Request{
		requestForTest(t, http.MethodPost, "/api/admin/players/7/name-change", `{"username":"bad name","actorPlayerId":7,"reason":"support"}`),
		requestForTest(t, http.MethodPost, "/api/admin/players/7/name-change/allow", `{"actorPlayerId":0,"reason":""}`),
		requestForTest(t, http.MethodGet, "/api/admin/players/7/name-changes?limit=500", ""),
	} {
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("path=%s status=%d", request.URL.String(), response.StatusCode)
		}
	}
}

// nameChangeApplication creates focused administrative name-change routes.
func nameChangeApplication(t *testing.T, online bool) (*fiber.App, *fakeManager, *[]codec.Packet) {
	t.Helper()
	manager := &fakeManager{record: testRecord()}
	store := &routeIdentityStore{manager: manager}
	service := playeridentity.New(store, manager, nil)
	dependencies := RemainingDependencies{Players: manager, Identity: service}
	packets := make([]codec.Packet, 0, 1)
	if online {
		live, connections := nameChangeLiveFixture(t, manager.record, &packets)
		dependencies.Live = live
		dependencies.Connections = connections
	}
	app := fiber.New()
	RegisterRemaining(app, dependencies)
	return app, manager, &packets
}

// nameChangeLiveFixture creates one online packet-capturing player.
func nameChangeLiveFixture(t *testing.T, record playerservice.Record, packets *[]codec.Packet) (*playerlive.Registry, *netconn.Registry) {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "name-change", Kind: "websocket", Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { *packets = append(*packets, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	connections := netconn.NewRegistry()
	if err = connections.Register(session); err != nil {
		t.Fatal(err)
	}
	peer, err := playerlive.NewSessionPeer("name-change", "websocket", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	player, err := playerlive.NewPlayer(playerlive.SnapshotFromRecord(record), peer)
	if err != nil {
		t.Fatal(err)
	}
	live := playerlive.NewRegistry()
	if err = live.Add(player); err != nil {
		t.Fatal(err)
	}
	return live, connections
}

// nameChangeRequest executes one focused HTTP request.
func nameChangeRequest(t *testing.T, app *fiber.App, method string, path string, body string) *http.Response {
	t.Helper()
	response, err := app.Test(requestForTest(t, method, path, body))
	if err != nil {
		t.Fatal(err)
	}
	return response
}
