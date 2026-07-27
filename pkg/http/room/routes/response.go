// Package routes contains protected room administration routes.
package routes

// ListResponse contains room list results.
type ListResponse struct {
	// Total stores the returned room count.
	Total int `json:"total"`
	// Items stores safe room rows.
	Items []RoomResponse `json:"items"`
}

// RoomResponse contains safe room metadata.
type RoomResponse struct {
	// ID identifies the room.
	ID int64 `json:"id"`
	// Version stores the optimistic-lock version.
	Version int64 `json:"version"`
	// Name stores the room name.
	Name string `json:"name"`
	// OwnerPlayerID identifies the owner.
	OwnerPlayerID int64 `json:"ownerPlayerId"`
	// OwnerName stores the owner snapshot.
	OwnerName string `json:"ownerName"`
	// ModelName stores the room layout model.
	ModelName string `json:"modelName"`
	// MaxUsers stores the room capacity.
	MaxUsers int `json:"maxUsers"`
	// CategoryID stores the optional category id.
	CategoryID *int64 `json:"categoryId,omitempty"`
	// Score stores the navigator score.
	Score int `json:"score"`
	// IsBundleTemplate reports whether the room is a bundle source.
	IsBundleTemplate bool `json:"isBundleTemplate"`
	// RollerSpeed stores room cycles between roller steps, or -1 when disabled.
	RollerSpeed int `json:"rollerSpeed"`
}

// RollerSettingsRequest contains an optimistic roller cadence update.
type RollerSettingsRequest struct {
	// ExpectedVersion stores the current durable room version.
	ExpectedVersion int64 `json:"expectedVersion"`
	// RollerSpeed stores cycles between steps from -1 through 20.
	RollerSpeed int `json:"rollerSpeed"`
}

// SettingsRequest contains one attributed optimistic room configuration update.
type SettingsRequest struct {
	// ExpectedVersion stores the current durable room version.
	ExpectedVersion int64 `json:"expectedVersion"`
	// ActorPlayerID identifies the authorized administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the configuration change.
	Reason string `json:"reason"`
	// Name optionally replaces the visible room name.
	Name *string `json:"name"`
	// Description optionally replaces the visible description.
	Description *string `json:"description"`
	// CategoryID optionally replaces the navigator category.
	CategoryID **int64 `json:"categoryId"`
	// Tags optionally replace the normalized room tags.
	Tags *[]string `json:"tags"`
	// MaxUsers optionally replaces room capacity.
	MaxUsers *int `json:"maxUsers"`
	// DoorMode optionally replaces the room access mode.
	DoorMode *int16 `json:"doorMode"`
	// Password optionally replaces the plaintext door password.
	Password *string `json:"password"`
	// TradeMode optionally replaces the room trading mode.
	TradeMode *int16 `json:"tradeMode"`
	// RollerSpeed optionally replaces the room roller cadence.
	RollerSpeed *int `json:"rollerSpeed"`
	// AllowWalkthrough optionally replaces walkthrough behavior.
	AllowWalkthrough *bool `json:"allowWalkthrough"`
	// AllowPets optionally replaces pet admission behavior.
	AllowPets *bool `json:"allowPets"`
	// AllowPetsEat optionally replaces pet feeding behavior.
	AllowPetsEat *bool `json:"allowPetsEat"`
	// HideWalls optionally replaces wall visibility.
	HideWalls *bool `json:"hideWalls"`
	// WallThickness optionally replaces wall thickness.
	WallThickness *int `json:"wallThickness"`
	// FloorThickness optionally replaces floor thickness.
	FloorThickness *int `json:"floorThickness"`
	// ChatMode optionally replaces room chat mode.
	ChatMode *int16 `json:"chatMode"`
	// ChatWeight optionally replaces room chat bubble weight.
	ChatWeight *int16 `json:"chatWeight"`
	// ChatSpeed optionally replaces room chat speed.
	ChatSpeed *int16 `json:"chatSpeed"`
	// ChatDistance optionally replaces room chat distance.
	ChatDistance *int16 `json:"chatDistance"`
	// ChatProtection optionally replaces flood protection.
	ChatProtection *int16 `json:"chatProtection"`
	// ModerationMute optionally replaces the mute policy.
	ModerationMute *int16 `json:"moderationMute"`
	// ModerationKick optionally replaces the kick policy.
	ModerationKick *int16 `json:"moderationKick"`
	// ModerationBan optionally replaces the ban policy.
	ModerationBan *int16 `json:"moderationBan"`
	// StaffPicked optionally replaces official navigator selection.
	StaffPicked *bool `json:"staffPicked"`
}

// OccupancyResponse contains active room occupancy.
type OccupancyResponse struct {
	// RoomID identifies the room.
	RoomID int64 `json:"roomId"`
	// Count stores active occupant count.
	Count int `json:"count"`
	// MaxUsers stores maximum occupancy.
	MaxUsers int `json:"maxUsers"`
	// PlayerIDs stores active player ids.
	PlayerIDs []int64 `json:"playerIds"`
}

// ForwardRequest contains room forwarding input.
type ForwardRequest struct {
	// TargetRoomID identifies the room clients should enter.
	TargetRoomID int64 `json:"targetRoomId"`
}

// TeleportRequest contains one player forwarding request.
type TeleportRequest struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the movement override.
	Reason string `json:"reason"`
	// TargetRoomID identifies the destination room.
	TargetRoomID int64 `json:"targetRoomId"`
	// Bypass reports whether closed-room gating should be skipped once.
	Bypass bool `json:"bypass"`
}

// ActionResponse contains admin runtime action counts.
type ActionResponse struct {
	// Matched stores matched runtime occupants.
	Matched int `json:"matched"`
	// Sent stores successful packet sends.
	Sent int `json:"sent"`
	// Errors stores failed packet sends.
	Errors int `json:"errors"`
}

// CategoryListResponse contains navigator room categories.
type CategoryListResponse struct {
	// Total stores the returned category count.
	Total int `json:"total"`
	// Items stores category rows.
	Items []CategoryResponse `json:"items"`
}

// CategoryResponse contains safe room category data.
type CategoryResponse struct {
	// ID identifies the category.
	ID int64 `json:"id"`
	// Caption stores the visible caption.
	Caption string `json:"caption"`
	// Visible reports whether the category is visible.
	Visible bool `json:"visible"`
	// Order stores navigator order.
	Order int `json:"order"`
}

// LiftedListResponse contains navigator lifted rooms.
type LiftedListResponse struct {
	// Total stores the returned lifted room count.
	Total int `json:"total"`
	// Items stores lifted room rows.
	Items []LiftedResponse `json:"items"`
}

// LiftedResponse contains safe lifted room data.
type LiftedResponse struct {
	// ID identifies the lifted row.
	ID int64 `json:"id"`
	// RoomID identifies the promoted room.
	RoomID int64 `json:"roomId"`
	// AreaID stores the visual area id.
	AreaID int `json:"areaId"`
	// Image stores the image key.
	Image string `json:"image"`
	// AssetRef stores an opaque CMS-owned asset reference.
	AssetRef string `json:"assetRef"`
	// Caption stores the caption.
	Caption string `json:"caption"`
	// Order stores Navigator display ordering.
	Order int `json:"order"`
	// StartsAt optionally stores the publication boundary.
	StartsAt *string `json:"startsAt,omitempty"`
	// EndsAt optionally stores the expiration boundary.
	EndsAt *string `json:"endsAt,omitempty"`
	// Version stores optimistic mutation order.
	Version int64 `json:"version"`
}
