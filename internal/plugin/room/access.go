// Package room implements bounded room capabilities for dynamic plugins.
package room

import (
	"context"
	"errors"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	muteallchanged "github.com/niflaot/pixels/internal/realm/room/control/events/muteallchanged"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstate "github.com/niflaot/pixels/networking/outbound/room/mute/state"
	"github.com/niflaot/pixels/pkg/bus"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
)

var (
	// ErrUnavailable reports a room capability without its realm dependency.
	ErrUnavailable = errors.New("plugin room access unavailable")
	// ErrInactiveRoom reports a live-only operation against an inactive room.
	ErrInactiveRoom = errors.New("plugin room is not active")
)

// Access exposes scoped room behavior.
type Access struct {
	// service owns durable room records and validation.
	service *roomservice.Service
	// live owns active room occupants and state.
	live *roomlive.Registry
	// connections projects live room state to occupants.
	connections *netconn.Registry
	// events publishes committed room state transitions.
	events bus.Publisher
	// scope identifies the calling plugin.
	scope *pluginruntime.Scope
}

// NewAccess creates one plugin-scoped room facade.
func NewAccess(service *roomservice.Service, live *roomlive.Registry, connections *netconn.Registry, events bus.Publisher, scope *pluginruntime.Scope) *Access {
	return &Access{service: service, live: live, connections: connections, events: events, scope: scope}
}

// Update applies one validated partial room mutation.
func (access *Access) Update(roomID int64, params sdkplugin.RoomUpdateParams) (sdkplugin.RoomSnapshot, error) {
	if access.service == nil || access.scope == nil || !access.scope.Enabled() {
		return sdkplugin.RoomSnapshot{}, ErrUnavailable
	}
	current, found, err := access.service.FindByID(context.Background(), roomID)
	if err != nil {
		return sdkplugin.RoomSnapshot{}, err
	}
	if !found {
		return sdkplugin.RoomSnapshot{}, roomservice.ErrRoomNotFound
	}
	updated, err := access.service.Update(context.Background(), roomID, current.Version.Version, fromSDK(params))
	if err != nil {
		return sdkplugin.RoomSnapshot{}, err
	}
	return snapshot(updated), nil
}

// Find returns one immutable durable room snapshot.
func (access *Access) Find(roomID int64) (sdkplugin.RoomSnapshot, bool) {
	if access.service == nil || access.scope == nil || !access.scope.Enabled() {
		return sdkplugin.RoomSnapshot{}, false
	}
	record, found, err := access.service.FindByID(context.Background(), roomID)
	if err != nil || !found {
		return sdkplugin.RoomSnapshot{}, false
	}
	return snapshot(record), true
}

// Occupants lists connected player identifiers inside an active room.
func (access *Access) Occupants(roomID int64) []int64 {
	if access.live == nil || access.scope == nil || !access.scope.Enabled() {
		return nil
	}
	active, found := access.live.Find(roomID)
	if !found {
		return nil
	}
	occupants := active.Occupants()
	playerIDs := make([]int64, 0, len(occupants))
	for _, occupant := range occupants {
		playerIDs = append(playerIDs, occupant.PlayerID)
	}
	return playerIDs
}

// SetMuteAll toggles whole-room muting for an active room.
func (access *Access) SetMuteAll(roomID int64, muted bool) error {
	if access.live == nil || access.scope == nil || !access.scope.Enabled() {
		return ErrUnavailable
	}
	active, found := access.live.Find(roomID)
	if !found {
		return ErrInactiveRoom
	}
	packet, err := outstate.Encode(muted)
	if err != nil {
		return err
	}
	active.SetMuteAll(muted)
	_ = broadcast.RoomPacket(context.Background(), access.connections, active, packet, 0)
	if access.events == nil {
		return nil
	}
	return access.events.Publish(context.Background(), bus.Event{
		Name: muteallchanged.Name,
		Payload: muteallchanged.Payload{
			RoomID: roomID,
			Muted:  muted,
		},
	})
}

// snapshot converts a durable room record into its SDK representation.
func snapshot(room roommodel.Room) sdkplugin.RoomSnapshot {
	return sdkplugin.RoomSnapshot{
		ID: room.ID, Version: room.Version.Version, OwnerPlayerID: room.OwnerPlayerID,
		OwnerName: room.OwnerName, Name: room.Name, Description: room.Description, ModelName: room.ModelName,
		DoorMode: int16(room.DoorMode), MaxUsers: room.MaxUsers, Score: room.Score, CategoryID: cloneInt64(room.CategoryID),
		TradeMode: int16(room.TradeMode), RollerSpeed: room.RollerSpeed, AllowWalkthrough: room.AllowWalkthrough,
		AllowPets: room.AllowPets, AllowPetsEat: room.AllowPetsEat, HideWalls: room.HideWalls, HideWired: room.HideWired,
		WallThickness: room.WallThickness, FloorThickness: room.FloorThickness, StaffPicked: room.StaffPicked,
		PublicRoom: room.PublicRoom,
	}
}

// fromSDK converts plugin room settings into realm-specific values.
func fromSDK(params sdkevent.RoomUpdateParams) roomservice.UpdateParams {
	return roomservice.UpdateParams{
		Name: params.Name, Description: params.Description, CategoryID: params.CategoryID, Tags: params.Tags,
		MaxUsers: params.MaxUsers, DoorMode: typedPointer[roommodel.DoorMode](params.DoorMode), Password: params.Password,
		TradeMode: typedPointer[roommodel.TradeMode](params.TradeMode), RollerSpeed: params.RollerSpeed,
		AllowWalkthrough: params.AllowWalkthrough, AllowPets: params.AllowPets, AllowPetsEat: params.AllowPetsEat,
		HideWalls: params.HideWalls, HideWired: params.HideWired, WallThickness: params.WallThickness,
		FloorThickness: params.FloorThickness, ChatMode: params.ChatMode, ChatWeight: params.ChatWeight,
		ChatSpeed: params.ChatSpeed, ChatDistance: params.ChatDistance, ChatProtection: params.ChatProtection,
		ModerationMute: typedPointer[roommodel.ModerationPolicy](params.ModerationMute),
		ModerationKick: typedPointer[roommodel.ModerationPolicy](params.ModerationKick),
		ModerationBan:  typedPointer[roommodel.ModerationPolicy](params.ModerationBan), StaffPicked: params.StaffPicked,
	}
}

// typedPointer converts an SDK enum pointer into one room realm enum.
func typedPointer[T ~int16](value *int16) *T {
	if value == nil {
		return nil
	}
	typed := T(*value)
	return &typed
}

// cloneInt64 copies one optional identifier.
func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
