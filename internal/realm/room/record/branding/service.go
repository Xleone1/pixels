package branding

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	netconn "github.com/niflaot/pixels/networking/connection"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
)

var (
	// ErrInvalid reports invalid branding mutation input.
	ErrInvalid = errors.New("invalid room branding request")
	// ErrNotFound reports a missing room, item, or branding record.
	ErrNotFound = errors.New("room branding target not found")
	// ErrIncompatible reports furniture without explicit branding support.
	ErrIncompatible = errors.New("furniture item is not branding compatible")
	// ErrConflict reports an optimistic version mismatch.
	ErrConflict = errors.New("room branding version conflict")
	// ErrForbidden reports a missing actor permission.
	ErrForbidden = errors.New("room branding action forbidden")
)

// Store persists room branding.
type Store interface {
	// List lists branding configurations for one room.
	List(context.Context, int64) ([]Config, error)
	// ListCompatible lists placed furniture explicitly marked as branding compatible.
	ListCompatible(context.Context, int64) ([]CompatibleItem, error)
	// Upsert creates or updates one configuration atomically with furniture state.
	Upsert(context.Context, Mutation, string) (Projection, error)
	// Disable disables one configuration atomically with furniture state.
	Disable(context.Context, int64, int64, int64, int64, string) (Projection, error)
}

// Service validates, authorizes, persists, and projects room branding.
type Service struct {
	// store persists branding records.
	store Store
	// permissions resolves actor capabilities.
	permissions permissionservice.Checker
	// manageNode identifies the required actor permission.
	manageNode permission.Node
	// runtime resolves active rooms.
	runtime *roomlive.Registry
	// connections sends live furniture updates.
	connections *netconn.Registry
}

// New creates a room branding service.
func New(store Store, permissions permissionservice.Checker, manageNode permission.Node, runtime *roomlive.Registry, connections *netconn.Registry) *Service {
	return &Service{store: store, permissions: permissions, manageNode: manageNode, runtime: runtime, connections: connections}
}

// List lists room branding configurations for one authorized actor.
func (service *Service) List(ctx context.Context, roomID int64, actorPlayerID int64) ([]Config, error) {
	if roomID <= 0 || actorPlayerID <= 0 {
		return nil, ErrInvalid
	}
	if err := service.authorize(ctx, actorPlayerID); err != nil {
		return nil, err
	}
	return service.store.List(ctx, roomID)
}

// ListCompatible lists branding-capable furniture for one authorized actor.
func (service *Service) ListCompatible(ctx context.Context, roomID int64, actorPlayerID int64) ([]CompatibleItem, error) {
	if roomID <= 0 || actorPlayerID <= 0 {
		return nil, ErrInvalid
	}
	if err := service.authorize(ctx, actorPlayerID); err != nil {
		return nil, err
	}
	return service.store.ListCompatible(ctx, roomID)
}

// Upsert validates and commits one branding configuration.
func (service *Service) Upsert(ctx context.Context, mutation Mutation) (Config, error) {
	mutation.ImageURL = strings.TrimSpace(mutation.ImageURL)
	mutation.ClickURL = strings.TrimSpace(mutation.ClickURL)
	mutation.AssetRef = strings.TrimSpace(mutation.AssetRef)
	if err := service.validate(ctx, mutation); err != nil {
		return Config{}, err
	}
	extraData := encodeExtraData(mutation, true)
	projection, err := service.store.Upsert(ctx, mutation, extraData)
	if err != nil {
		return Config{}, err
	}
	service.project(ctx, projection)
	return projection.Config, nil
}

// Disable disables one branding configuration.
func (service *Service) Disable(ctx context.Context, roomID int64, brandingID int64, expectedVersion int64, actorPlayerID int64) (Config, error) {
	if roomID <= 0 || brandingID <= 0 || expectedVersion <= 0 || actorPlayerID <= 0 {
		return Config{}, ErrInvalid
	}
	if err := service.authorize(ctx, actorPlayerID); err != nil {
		return Config{}, err
	}
	extraData := `{"clickUrl":"","imageUrl":"","offsetX":"0","offsetY":"0","offsetZ":"0","state":"0"}`
	projection, err := service.store.Disable(ctx, roomID, brandingID, expectedVersion, actorPlayerID, extraData)
	if err != nil {
		return Config{}, err
	}
	service.project(ctx, projection)
	return projection.Config, nil
}

// validate validates and authorizes a branding mutation.
func (service *Service) validate(ctx context.Context, mutation Mutation) error {
	if mutation.RoomID <= 0 || mutation.FurnitureItemID <= 0 || !mutation.Kind.Valid() || mutation.ActorPlayerID <= 0 || mutation.AssetRef == "" || len(mutation.AssetRef) > 255 || mutation.State < 0 || mutation.State > 255 {
		return ErrInvalid
	}
	if !validOffset(mutation.OffsetX) || !validOffset(mutation.OffsetY) || !validOffset(mutation.OffsetZ) || len(mutation.ImageURL) > 2048 || len(mutation.ClickURL) > 2048 || !validPublicURL(mutation.ImageURL) {
		return ErrInvalid
	}
	if mutation.Kind == KindBackground && strings.TrimSpace(mutation.ClickURL) != "" || mutation.ClickURL != "" && !validPublicURL(mutation.ClickURL) {
		return ErrInvalid
	}
	return service.authorize(ctx, mutation.ActorPlayerID)
}

// authorize verifies actor capability.
func (service *Service) authorize(ctx context.Context, actorPlayerID int64) error {
	if service.permissions == nil {
		return ErrForbidden
	}
	allowed, err := service.permissions.HasPermission(ctx, actorPlayerID, service.manageNode)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrForbidden
	}
	return nil
}

// project refreshes one active room furniture item.
func (service *Service) project(ctx context.Context, projection Projection) {
	active, found := service.runtime.Find(projection.Config.RoomID)
	if !found {
		return
	}
	active.SetFurnitureExtraData(projection.Config.FurnitureItemID, projection.ExtraData)
	record := outupdate.FloorItem{ID: projection.Config.FurnitureItemID, SpriteID: projection.SpriteID, X: projection.X, Y: projection.Y, Rotation: projection.Rotation, Z: roomprojection.FurnitureHeightValue(projection.Z), ExtraData: projection.ExtraData, OwnerID: projection.OwnerPlayerID}
	record.Data = roomprojection.SpecializedObjectData(projection.InteractionType, projection.ExtraData)
	packet, err := outupdate.Encode(record)
	if err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
}

// validOffset reports whether a renderer offset is bounded.
func validOffset(value int) bool { return value >= -4096 && value <= 4096 }

// validPublicURL reports whether a URL is durable HTTP or HTTPS.
func validPublicURL(value string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	return err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "http")
}
