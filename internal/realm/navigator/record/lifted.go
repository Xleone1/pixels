package record

import (
	"time"

	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// LiftedRoom stores a promoted navigator room entry.
type LiftedRoom struct {
	// Base contains shared durable record fields.
	sharedmodel.Base

	// RoomID identifies the promoted room.
	RoomID int64

	// AreaID stores the visual area id.
	AreaID int

	// Image stores the image key or URL.
	Image string

	// AssetRef stores an opaque CMS-owned asset reference.
	AssetRef string

	// Caption stores the promotion caption.
	Caption string

	// Order stores navigator display ordering.
	Order int

	// StartsAt optionally stores promotion start time.
	StartsAt *time.Time

	// EndsAt optionally stores promotion end time.
	EndsAt *time.Time

	// CreatedByPlayerID optionally identifies the creating administrative actor.
	CreatedByPlayerID *int64

	// UpdatedByPlayerID optionally identifies the latest administrative actor.
	UpdatedByPlayerID *int64
}

// LiftedRoomMutation contains one attributed Navigator image mutation.
type LiftedRoomMutation struct {
	// RoomID identifies the promoted room.
	RoomID int64
	// AreaID stores the visual area id.
	AreaID int
	// Image stores a durable public URL.
	Image string
	// AssetRef stores an opaque CMS-owned asset reference.
	AssetRef string
	// Caption stores visible Navigator copy.
	Caption string
	// Order stores display ordering.
	Order int
	// StartsAt optionally stores the publication boundary.
	StartsAt *time.Time
	// EndsAt optionally stores the expiration boundary.
	EndsAt *time.Time
	// ExpectedVersion stores zero for creation or the current version for update.
	ExpectedVersion int64
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64
}
