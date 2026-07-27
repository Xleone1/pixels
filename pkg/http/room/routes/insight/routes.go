// Package insight exposes protected room statistics and profiling.
package insight

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	roomrealm "github.com/niflaot/pixels/internal/realm/room"
	roominsight "github.com/niflaot/pixels/internal/realm/room/record/insight"
	"github.com/niflaot/pixels/pkg/http/adminaction"
)

// Register mounts protected room insight routes.
func Register(app *fiber.App, service *roominsight.Service, actions *adminaction.Service) {
	app.Get("/api/admin/rooms/:id/stats", statsHandler(service, actions))
	app.Get("/api/admin/rooms/:id/profile", profileHandler(service, actions))
}

// statsHandler returns durable counters with live occupancy.
func statsHandler(service *roominsight.Service, actions *adminaction.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if err := authorize(ctx, actions); err != nil {
			return err
		}
		roomID, err := roomID(ctx)
		if err != nil {
			return err
		}
		stats, found, err := service.Stats(ctx.Context(), roomID)
		if err != nil {
			return err
		}
		if !found {
			return fiber.NewError(fiber.StatusNotFound, "room not found")
		}
		return ctx.JSON(stats)
	}
}

// profileHandler returns bounded live runtime observability.
func profileHandler(service *roominsight.Service, actions *adminaction.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if err := authorize(ctx, actions); err != nil {
			return err
		}
		roomID, err := roomID(ctx)
		if err != nil {
			return err
		}
		return ctx.JSON(service.Profile(roomID))
	}
}

// authorize resolves and checks the administrative actor query parameter.
func authorize(ctx *fiber.Ctx, actions *adminaction.Service) error {
	actorID, err := strconv.ParseInt(ctx.Query("actorPlayerId"), 10, 64)
	if err != nil || actorID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid room insight actor")
	}
	if err = actions.Authorize(ctx.Context(), actorID, roomrealm.InsightRead); err != nil {
		return adminaction.HTTPError(err)
	}
	return nil
}

// roomID parses a positive room identifier.
func roomID(ctx *fiber.Ctx) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid room id")
	}
	return value, nil
}
