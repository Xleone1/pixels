package tools

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outopen "github.com/niflaot/pixels/networking/outbound/furniture/wired/tools/open"
)

var (
	// wiredConfigureAny identifies global room WIRED configuration authority.
	wiredConfigureAny = permission.Node("room.wired.configure.any")
	// wiredInspect identifies read-only Creator Tools authority.
	wiredInspect = permission.Node("room.wired.inspect")
)

// New creates a WIRED Creator Tools service.
func New(config roomwired.Config, players *playerlive.Registry, bindings *binding.Registry, connections *netconn.Registry, rooms *roomlive.Registry, records roomservice.ConfigManager, runtime *wiredruntime.Engine, variables *variable.Service, permissions permissionservice.Checker) *Service {
	return &Service{
		config: config.Normalize(), players: players, bindings: bindings,
		connections: connections, rooms: rooms, records: records, runtime: runtime,
		variables: variables, permissions: permissions, preferences: make(map[preferenceKey]Preferences),
	}
}

// Actor resolves and authorizes one connected room occupant.
func (service *Service) Actor(ctx context.Context, connection netconn.Context, roomID int64) (int64, PermissionsDocument, error) {
	current, found := service.bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	if !found {
		return 0, PermissionsDocument{}, ErrForbidden
	}
	return service.authorize(ctx, current.PlayerID, roomID)
}

// Open sends the Creator Tools panel to an authorized connected player.
func (service *Service) Open(ctx context.Context, username string, initialTab string) error {
	player, found := service.findPlayer(username)
	if !found {
		return ErrForbidden
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return ErrForbidden
	}
	if _, _, err := service.authorize(ctx, player.ID(), roomID); err != nil {
		return err
	}
	current, found := service.bindings.FindByPlayer(player.ID())
	if !found {
		return ErrForbidden
	}
	connection, found := service.connections.Get(current.ConnectionKind, current.ConnectionID)
	if !found {
		return ErrForbidden
	}
	packet, err := outopen.Encode(roomID, normalizeTab(initialTab))
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// normalizeTab resolves one supported Creator Tools entry tab.
func normalizeTab(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "variables":
		return "variables"
	case "inspection":
		return "inspection"
	case "settings":
		return "settings"
	default:
		return "monitor"
	}
}

// Snapshot returns one authorized Creator Tools snapshot document.
func (service *Service) Snapshot(ctx context.Context, playerID int64, roomID int64, permissions PermissionsDocument) ([]byte, error) {
	if !permissions.CanInspect {
		return nil, ErrForbidden
	}
	return service.snapshot(ctx, playerID, roomID, permissions)
}

// Inspect returns one authorized entity inspection document.
func (service *Service) Inspect(roomID int64, entityType string, entityID int64, permissions PermissionsDocument) ([]byte, error) {
	if !permissions.CanInspect {
		return nil, ErrForbidden
	}
	return service.inspection(roomID, entityType, entityID, permissions)
}

// Mutate applies one bounded variable mutation and returns a fresh snapshot.
func (service *Service) Mutate(ctx context.Context, playerID int64, roomID int64, operation string, scopeValue int32, scopeID int64, name string, intValue string, stringValue string, permissions PermissionsDocument) ([]byte, string, error) {
	if !permissions.CanMutate {
		return nil, "forbidden", ErrForbidden
	}
	scope, scopeID, name, err := service.validateMutation(ctx, roomID, scopeValue, scopeID, name)
	if err != nil {
		return nil, "invalid_request", err
	}
	value := variable.Value{RoomID: roomID, Scope: scope, ScopeID: scopeID, Name: name, StringValue: stringValue, UpdatedByPlayerID: playerID, UpdatedAt: time.Now()}
	switch strings.ToLower(strings.TrimSpace(operation)) {
	case "set":
		value.IntValue, err = strconv.ParseInt(strings.TrimSpace(intValue), 10, 64)
		if err == nil {
			_, err = service.variables.Set(ctx, value)
		}
	case "change":
		var delta int64
		delta, err = strconv.ParseInt(strings.TrimSpace(intValue), 10, 64)
		if err == nil {
			_, _, err = service.variables.Change(ctx, value, delta)
		}
	case "delete":
		var deleted bool
		_, deleted, err = service.variables.Delete(ctx, roomID, scope, scopeID, name)
		if err == nil && !deleted {
			return nil, "not_found", nil
		}
	default:
		return nil, "invalid_request", ErrInvalidRequest
	}
	if err != nil {
		return nil, mutationErrorCode(err), err
	}
	document, err := service.snapshot(ctx, playerID, roomID, permissions)
	return document, "saved", err
}

// Settings persists room visibility and panel refresh preferences.
func (service *Service) Settings(ctx context.Context, playerID int64, roomID int64, liveUpdates bool, hideBoxes bool, permissions PermissionsDocument) ([]byte, error) {
	if !permissions.CanInspect {
		return nil, ErrForbidden
	}
	current, found, err := service.records.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrInvalidRequest
	}
	if current.HideWired != hideBoxes {
		if !permissions.CanConfigure {
			return nil, ErrForbidden
		}
		if _, err = service.records.Update(ctx, roomID, current.Version.Version, roomservice.UpdateParams{HideWired: &hideBoxes}); err != nil {
			return nil, err
		}
	}
	service.setPreferences(playerID, roomID, Preferences{LiveUpdates: liveUpdates, HideBoxes: hideBoxes})
	return service.snapshot(ctx, playerID, roomID, permissions)
}

// authorize resolves room-local and global Creator Tools capabilities.
func (service *Service) authorize(ctx context.Context, playerID int64, roomID int64) (int64, PermissionsDocument, error) {
	player, found := service.players.Find(playerID)
	if !found {
		return 0, PermissionsDocument{}, ErrForbidden
	}
	currentRoom, found := player.CurrentRoom()
	if !found || currentRoom != roomID {
		return 0, PermissionsDocument{}, ErrForbidden
	}
	active, found := service.rooms.Find(roomID)
	if !found {
		return 0, PermissionsDocument{}, ErrForbidden
	}
	canConfigure := active.CanManageFurniture(playerID) || service.allowed(ctx, playerID, wiredConfigureAny)
	canMutate := canConfigure || service.allowed(ctx, playerID, VariablesManage)
	canInspect := canMutate || service.allowed(ctx, playerID, wiredInspect)
	if !canInspect {
		return 0, PermissionsDocument{}, ErrForbidden
	}
	return playerID, PermissionsDocument{CanInspect: canInspect, CanMutate: canMutate, CanConfigure: canConfigure}, nil
}

// allowed resolves one permission and treats resolver errors as denial.
func (service *Service) allowed(ctx context.Context, playerID int64, node permission.Node) bool {
	if service.permissions == nil {
		return false
	}
	allowed, err := service.permissions.HasPermission(ctx, playerID, node)
	return err == nil && allowed
}

// findPlayer resolves one exact connected username.
func (service *Service) findPlayer(username string) (*playerlive.Player, bool) {
	username = strings.TrimSpace(username)
	for _, player := range service.players.Snapshot() {
		if strings.EqualFold(player.Username(), username) {
			return player, true
		}
	}
	return nil, false
}

// mutationErrorCode maps domain errors to stable result codes.
func mutationErrorCode(err error) string {
	var numberError *strconv.NumError
	if errors.As(err, &numberError) {
		return "invalid_request"
	}
	if errors.Is(err, variable.ErrInvalid) || errors.Is(err, variable.ErrReadOnly) {
		return "invalid_request"
	}
	if errors.Is(err, variable.ErrLimit) {
		return "limit_reached"
	}
	return "internal_error"
}
