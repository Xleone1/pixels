// Package branding exposes protected room branding administration.
package branding

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	roombranding "github.com/niflaot/pixels/internal/realm/room/record/branding"
)

// MutationRequest contains one attributed branding mutation.
type MutationRequest struct {
	// Kind identifies background or billboard behavior.
	Kind roombranding.Kind `json:"kind"`
	// AssetRef stores an opaque CMS-owned asset reference.
	AssetRef string `json:"assetRef"`
	// ImageURL stores the durable public image URL.
	ImageURL string `json:"imageUrl"`
	// ClickURL stores the optional billboard target URL.
	ClickURL string `json:"clickUrl"`
	// State stores the visual state.
	State int16 `json:"state"`
	// OffsetX stores the horizontal renderer offset.
	OffsetX int `json:"offsetX"`
	// OffsetY stores the vertical renderer offset.
	OffsetY int `json:"offsetY"`
	// OffsetZ stores the depth renderer offset.
	OffsetZ int `json:"offsetZ"`
	// ExpectedVersion stores zero for creation or the current version for update.
	ExpectedVersion int64 `json:"expectedVersion"`
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
}

// DisableRequest contains one attributed branding disable operation.
type DisableRequest struct {
	// ExpectedVersion stores the current branding version.
	ExpectedVersion int64 `json:"expectedVersion"`
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
}

// ListResponse contains room branding configurations.
type ListResponse struct {
	// Total stores returned configurations.
	Total int `json:"total"`
	// Items stores room branding configurations.
	Items []roombranding.Config `json:"items"`
}

// CompatibleResponse contains compatible furniture items.
type CompatibleResponse struct {
	// Total stores returned furniture items.
	Total int `json:"total"`
	// Items stores compatible furniture items.
	Items []roombranding.CompatibleItem `json:"items"`
}

// Register mounts protected room branding routes.
func Register(app *fiber.App, service *roombranding.Service) {
	app.Get("/api/admin/rooms/:id/branding", listHandler(service))
	app.Get("/api/admin/rooms/:id/branding-compatible", compatibleHandler(service))
	app.Put("/api/admin/rooms/:id/branding/:itemId", upsertHandler(service))
	app.Delete("/api/admin/rooms/:id/branding/:brandingId", disableHandler(service))
}

// listHandler lists room branding.
func listHandler(service *roombranding.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := identifier(ctx, "id")
		if err != nil {
			return err
		}
		actorID, err := actorIdentifier(ctx)
		if err != nil {
			return err
		}
		items, err := service.List(ctx.Context(), roomID, actorID)
		if err != nil {
			return mapError(err)
		}
		return ctx.JSON(ListResponse{Total: len(items), Items: items})
	}
}

// compatibleHandler lists branding-capable furniture.
func compatibleHandler(service *roombranding.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := identifier(ctx, "id")
		if err != nil {
			return err
		}
		actorID, err := actorIdentifier(ctx)
		if err != nil {
			return err
		}
		items, err := service.ListCompatible(ctx.Context(), roomID, actorID)
		if err != nil {
			return mapError(err)
		}
		return ctx.JSON(CompatibleResponse{Total: len(items), Items: items})
	}
}

// upsertHandler creates or updates room branding.
func upsertHandler(service *roombranding.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := identifier(ctx, "id")
		if err != nil {
			return err
		}
		itemID, err := identifier(ctx, "itemId")
		if err != nil {
			return err
		}
		var request MutationRequest
		if err = ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid room branding request")
		}
		config, err := service.Upsert(ctx.Context(), roombranding.Mutation{
			RoomID: roomID, FurnitureItemID: itemID, Kind: request.Kind,
			AssetRef: request.AssetRef, ImageURL: request.ImageURL, ClickURL: request.ClickURL,
			State: request.State, OffsetX: request.OffsetX, OffsetY: request.OffsetY, OffsetZ: request.OffsetZ,
			ExpectedVersion: request.ExpectedVersion, ActorPlayerID: request.ActorPlayerID,
		})
		if err != nil {
			return mapError(err)
		}
		return ctx.JSON(config)
	}
}

// disableHandler disables room branding.
func disableHandler(service *roombranding.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := identifier(ctx, "id")
		if err != nil {
			return err
		}
		brandingID, err := identifier(ctx, "brandingId")
		if err != nil {
			return err
		}
		var request DisableRequest
		if err = ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid room branding disable request")
		}
		config, err := service.Disable(ctx.Context(), roomID, brandingID, request.ExpectedVersion, request.ActorPlayerID)
		if err != nil {
			return mapError(err)
		}
		return ctx.JSON(config)
	}
}

// actorIdentifier parses the attributed administrative reader.
func actorIdentifier(ctx *fiber.Ctx) (int64, error) {
	value, err := strconv.ParseInt(ctx.Query("actorPlayerId"), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid room branding actor")
	}
	return value, nil
}

// identifier parses a positive route identifier.
func identifier(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid room branding identifier")
	}
	return value, nil
}

// mapError maps branding domain failures to HTTP errors.
func mapError(err error) error {
	switch {
	case errors.Is(err, roombranding.ErrInvalid):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, roombranding.ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, err.Error())
	case errors.Is(err, roombranding.ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	case errors.Is(err, roombranding.ErrIncompatible), errors.Is(err, roombranding.ErrConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	default:
		return err
	}
}
