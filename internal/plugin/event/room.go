package event

import (
	"context"
	"errors"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

// DispatchRoomUpdate sends one cancellable room settings mutation event.
func (hub *Hub) DispatchRoomUpdate(ctx context.Context, player sdkplayer.Player, roomID int64, params roomservice.UpdateParams) (roomservice.UpdateParams, bool) {
	if !hub.HasListeners(sdkevent.RoomUpdateName) {
		return params, false
	}
	event := &sdkevent.RoomUpdate{Player: player, RoomID: roomID, Params: roomUpdateToSDK(params)}
	err := hub.Dispatch(ctx, event)
	return roomUpdateFromSDK(event.Params, params.AllowReservedTags), errors.Is(err, ErrEventCancelled)
}

// DispatchRoomUnitMove sends one cancellable live movement event.
func (hub *Hub) DispatchRoomUnitMove(ctx context.Context, player sdkplayer.Player, roomID int64, targetX int, targetY int) (int, int, bool) {
	if !hub.HasListeners(sdkevent.RoomUnitMoveName) {
		return targetX, targetY, false
	}
	event := &sdkevent.RoomUnitMove{Player: player, RoomID: roomID, TargetX: targetX, TargetY: targetY}
	err := hub.Dispatch(ctx, event)
	return event.TargetX, event.TargetY, errors.Is(err, ErrEventCancelled)
}

// DispatchRoomEnterAttempt sends one cancellable room admission request.
func (hub *Hub) DispatchRoomEnterAttempt(ctx context.Context, player sdkplayer.Player, roomID int64, trusted bool) bool {
	if !hub.HasListeners(sdkevent.RoomEnterAttemptName) {
		return false
	}
	event := &sdkevent.RoomEnterAttempt{Player: player, RoomID: roomID, Trusted: trusted}
	return errors.Is(hub.Dispatch(ctx, event), ErrEventCancelled)
}

// DispatchRoomCreate sends one cancellable room creation.
func (hub *Hub) DispatchRoomCreate(ctx context.Context, params roomservice.CreateParams) (roomservice.CreateParams, bool) {
	if !hub.HasListeners(sdkevent.RoomCreateName) {
		return params, false
	}
	event := &sdkevent.RoomCreate{
		Player: hub.player(params.OwnerPlayerID), RoomName: params.Name, Description: params.Description,
		ModelName: params.ModelName, MaxUsers: params.MaxUsers, CategoryID: params.CategoryID,
		TradeMode: int16(params.TradeMode), Tags: append([]string(nil), params.Tags...),
	}
	err := hub.Dispatch(ctx, event)
	params.Name, params.Description, params.ModelName, params.MaxUsers = event.RoomName, event.Description, event.ModelName, event.MaxUsers
	params.CategoryID, params.TradeMode, params.Tags = event.CategoryID, roommodel.TradeMode(event.TradeMode), append([]string(nil), event.Tags...)
	return params, errors.Is(err, ErrEventCancelled)
}

// roomUpdateToSDK converts internal settings without exposing realm types.
func roomUpdateToSDK(params roomservice.UpdateParams) sdkevent.RoomUpdateParams {
	return sdkevent.RoomUpdateParams{
		Name: params.Name, Description: params.Description, CategoryID: params.CategoryID,
		Tags: params.Tags, MaxUsers: params.MaxUsers, DoorMode: int16Pointer(params.DoorMode),
		Password: params.Password, TradeMode: int16Pointer(params.TradeMode), RollerSpeed: params.RollerSpeed,
		AllowWalkthrough: params.AllowWalkthrough, AllowPets: params.AllowPets, AllowPetsEat: params.AllowPetsEat,
		HideWalls: params.HideWalls, HideWired: params.HideWired, WallThickness: params.WallThickness,
		FloorThickness: params.FloorThickness, ChatMode: params.ChatMode, ChatWeight: params.ChatWeight,
		ChatSpeed: params.ChatSpeed, ChatDistance: params.ChatDistance, ChatProtection: params.ChatProtection,
		ModerationMute: int16Pointer(params.ModerationMute), ModerationKick: int16Pointer(params.ModerationKick),
		ModerationBan: int16Pointer(params.ModerationBan), StaffPicked: params.StaffPicked,
	}
}

// roomUpdateFromSDK converts plugin settings back to validated realm values.
func roomUpdateFromSDK(params sdkevent.RoomUpdateParams, allowReservedTags bool) roomservice.UpdateParams {
	return roomservice.UpdateParams{
		Name: params.Name, Description: params.Description, CategoryID: params.CategoryID,
		Tags: params.Tags, MaxUsers: params.MaxUsers, DoorMode: typedPointer[roommodel.DoorMode](params.DoorMode),
		Password: params.Password, TradeMode: typedPointer[roommodel.TradeMode](params.TradeMode), RollerSpeed: params.RollerSpeed,
		AllowWalkthrough: params.AllowWalkthrough, AllowPets: params.AllowPets, AllowPetsEat: params.AllowPetsEat,
		HideWalls: params.HideWalls, HideWired: params.HideWired, WallThickness: params.WallThickness,
		FloorThickness: params.FloorThickness, ChatMode: params.ChatMode, ChatWeight: params.ChatWeight,
		ChatSpeed: params.ChatSpeed, ChatDistance: params.ChatDistance, ChatProtection: params.ChatProtection,
		ModerationMute: typedPointer[roommodel.ModerationPolicy](params.ModerationMute),
		ModerationKick: typedPointer[roommodel.ModerationPolicy](params.ModerationKick),
		ModerationBan:  typedPointer[roommodel.ModerationPolicy](params.ModerationBan),
		StaffPicked:    params.StaffPicked, AllowReservedTags: allowReservedTags,
	}
}

// int16Pointer copies one typed int16 pointer into its SDK representation.
func int16Pointer[T ~int16](value *T) *int16 {
	if value == nil {
		return nil
	}
	result := int16(*value)
	return &result
}

// typedPointer copies one SDK int16 pointer into a realm-specific type.
func typedPointer[T ~int16](value *int16) *T {
	if value == nil {
		return nil
	}
	result := T(*value)
	return &result
}
