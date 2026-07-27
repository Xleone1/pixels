package routes

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
	roomrealm "github.com/niflaot/pixels/internal/realm/room"
	"github.com/niflaot/pixels/pkg/http/adminaction"
)

// NavigatorMediaRequest contains one attributed Navigator image replacement.
type NavigatorMediaRequest struct {
	// Image stores a durable public URL.
	Image string `json:"image"`
	// AssetRef stores an opaque CMS-owned asset reference.
	AssetRef string `json:"assetRef"`
	// Caption stores visible Navigator copy.
	Caption string `json:"caption"`
	// AreaID stores the visual area id.
	AreaID int `json:"areaId"`
	// Order stores Navigator display ordering.
	Order int `json:"order"`
	// StartsAt optionally stores the publication boundary.
	StartsAt *time.Time `json:"startsAt"`
	// EndsAt optionally stores the expiration boundary.
	EndsAt *time.Time `json:"endsAt"`
	// ExpectedVersion stores zero for creation or the current version for update.
	ExpectedVersion int64 `json:"expectedVersion"`
	// ActorPlayerID identifies the authorized administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the mutation.
	Reason string `json:"reason"`
}

// NavigatorMediaDisableRequest contains one attributed media disable.
type NavigatorMediaDisableRequest struct {
	// ExpectedVersion stores the current media version.
	ExpectedVersion int64 `json:"expectedVersion"`
	// ActorPlayerID identifies the authorized administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the mutation.
	Reason string `json:"reason"`
}

// navigatorMediaHandler creates or updates room Navigator media.
func navigatorMediaHandler(navigator navservice.Manager, actions *adminaction.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := roomIDParam(ctx)
		if err != nil {
			return err
		}
		var request NavigatorMediaRequest
		if err = ctx.BodyParser(&request); err != nil || strings.TrimSpace(request.Reason) == "" {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Navigator media request")
		}
		if err = actions.Authorize(ctx.Context(), request.ActorPlayerID, roomrealm.NavigatorMediaManage); err != nil {
			return adminaction.HTTPError(err)
		}
		value, err := navigator.UpsertLiftedRoom(ctx.Context(), navmodel.LiftedRoomMutation{
			RoomID: roomID, AreaID: request.AreaID, Image: request.Image,
			AssetRef: request.AssetRef, Caption: request.Caption, Order: request.Order,
			StartsAt: request.StartsAt, EndsAt: request.EndsAt,
			ExpectedVersion: request.ExpectedVersion, ActorPlayerID: request.ActorPlayerID,
		})
		if err != nil {
			return navigatorMediaError(err)
		}
		return ctx.JSON(liftedResponse(value))
	}
}

// disableNavigatorMediaHandler soft deletes room Navigator media.
func disableNavigatorMediaHandler(navigator navservice.Manager, actions *adminaction.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := roomIDParam(ctx)
		if err != nil {
			return err
		}
		var request NavigatorMediaDisableRequest
		if err = ctx.BodyParser(&request); err != nil || strings.TrimSpace(request.Reason) == "" {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Navigator media disable request")
		}
		if err = actions.Authorize(ctx.Context(), request.ActorPlayerID, roomrealm.NavigatorMediaManage); err != nil {
			return adminaction.HTTPError(err)
		}
		value, err := navigator.DisableLiftedRoom(ctx.Context(), roomID, request.ExpectedVersion, request.ActorPlayerID)
		if err != nil {
			return navigatorMediaError(err)
		}
		return ctx.JSON(liftedResponse(value))
	}
}

// navigatorMediaError maps expected Navigator media failures.
func navigatorMediaError(err error) error {
	if errors.Is(err, navservice.ErrInvalidLiftedRoom) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if errors.Is(err, navmodel.ErrLiftedRoomNotFound) {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	if errors.Is(err, navmodel.ErrLiftedRoomConflict) {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}
	return err
}

// liftedResponse maps one Navigator media row.
func liftedResponse(room navmodel.LiftedRoom) LiftedResponse {
	return LiftedResponse{
		ID: room.ID, RoomID: room.RoomID, AreaID: room.AreaID, Image: room.Image,
		AssetRef: room.AssetRef, Caption: room.Caption, Order: room.Order,
		StartsAt: timeText(room.StartsAt), EndsAt: timeText(room.EndsAt),
		Version: room.Version.Version,
	}
}

// timeText formats one optional timestamp for HTTP responses.
func timeText(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.UTC().Format(time.RFC3339)
	return &formatted
}
