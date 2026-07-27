package routes

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	roomrealm "github.com/niflaot/pixels/internal/realm/room"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/pkg/http/adminaction"
)

// settingsHandler updates safe durable room configuration fields.
func settingsHandler(rooms roomservice.ConfigManager, runtime *roomlive.Registry, actions *adminaction.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		roomID, err := roomIDParam(ctx)
		if err != nil {
			return err
		}
		var request SettingsRequest
		if err = ctx.BodyParser(&request); err != nil || request.ExpectedVersion <= 0 || request.ActorPlayerID <= 0 || strings.TrimSpace(request.Reason) == "" {
			return fiber.NewError(fiber.StatusBadRequest, "invalid room settings request")
		}
		if err = actions.Authorize(ctx.Context(), request.ActorPlayerID, roomrealm.SettingsAnyManage); err != nil {
			return adminaction.HTTPError(err)
		}
		updated, err := rooms.Update(ctx.Context(), roomID, request.ExpectedVersion, settingsParams(request))
		if err != nil {
			return settingsError(err)
		}
		if active, found := runtime.Find(roomID); found {
			active.UpdateSettings(updated.CategoryID, updated.MaxUsers, updated.ChatDistance, updated.ChatProtection)
			active.UpdatePetSettings(updated.AllowPets, updated.AllowPetsEat)
			active.UpdateCategoryAndTrade(updated.CategoryID, int16(updated.TradeMode))
			active.UpdateRollerSpeed(updated.RollerSpeed)
		}
		return ctx.JSON(roomResponse(updated))
	}
}

// settingsParams maps the explicit HTTP shape to room domain options.
func settingsParams(request SettingsRequest) roomservice.UpdateParams {
	params := roomservice.UpdateParams{
		Name: request.Name, Description: request.Description, CategoryID: request.CategoryID,
		Tags: request.Tags, MaxUsers: request.MaxUsers, Password: request.Password,
		RollerSpeed: request.RollerSpeed, AllowWalkthrough: request.AllowWalkthrough,
		AllowPets: request.AllowPets, AllowPetsEat: request.AllowPetsEat,
		HideWalls: request.HideWalls, WallThickness: request.WallThickness,
		FloorThickness: request.FloorThickness, ChatMode: request.ChatMode,
		ChatWeight: request.ChatWeight, ChatSpeed: request.ChatSpeed,
		ChatDistance: request.ChatDistance, ChatProtection: request.ChatProtection,
		StaffPicked: request.StaffPicked, AllowReservedTags: true,
	}
	if request.DoorMode != nil {
		value := roommodel.DoorMode(*request.DoorMode)
		params.DoorMode = &value
	}
	if request.TradeMode != nil {
		value := roommodel.TradeMode(*request.TradeMode)
		params.TradeMode = &value
	}
	if request.ModerationMute != nil {
		value := roommodel.ModerationPolicy(*request.ModerationMute)
		params.ModerationMute = &value
	}
	if request.ModerationKick != nil {
		value := roommodel.ModerationPolicy(*request.ModerationKick)
		params.ModerationKick = &value
	}
	if request.ModerationBan != nil {
		value := roommodel.ModerationPolicy(*request.ModerationBan)
		params.ModerationBan = &value
	}
	return params
}

// settingsError maps expected optimistic room update failures.
func settingsError(err error) error {
	if errors.Is(err, roomservice.ErrRoomNotFound) {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	if errors.Is(err, roomservice.ErrVersionConflict) {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}
	return fiber.NewError(fiber.StatusBadRequest, err.Error())
}
