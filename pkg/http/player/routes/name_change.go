package routes

import (
	"context"

	"github.com/gofiber/fiber/v2"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerprofile "github.com/niflaot/pixels/internal/realm/player/profile"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roombroadcast "github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	outroomname "github.com/niflaot/pixels/networking/outbound/user/room/name"
)

// allowNameChange enables one self-service username change.
func allowNameChange(dependencies RemainingDependencies) fiber.Handler {
	return nameChangeAuthorization(dependencies, true)
}

// revokeNameChange disables one self-service username change.
func revokeNameChange(dependencies RemainingDependencies) fiber.Handler {
	return nameChangeAuthorization(dependencies, false)
}

// setNameChangeAuthorization replaces one self-service username policy.
func setNameChangeAuthorization(dependencies RemainingDependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request NameChangeAuthorizationRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid name-change authorization request")
		}
		return applyNameChangeAuthorization(ctx, dependencies, request)
	}
}

// nameChangeAuthorization creates one fixed policy handler.
func nameChangeAuthorization(dependencies RemainingDependencies, allowed bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var request NameChangeAuthorizationRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid name-change authorization request")
		}
		request.Allowed = allowed
		return applyNameChangeAuthorization(ctx, dependencies, request)
	}
}

// applyNameChangeAuthorization persists and projects one policy mutation.
func applyNameChangeAuthorization(ctx *fiber.Ctx, dependencies RemainingDependencies, request NameChangeAuthorizationRequest) error {
	id, err := remainingPlayerID(ctx)
	if err != nil {
		return err
	}
	if !attributed(request.ActorPlayerID, request.Reason) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid name-change authorization request")
	}
	if dependencies.Identity == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "name-change identity service unavailable")
	}
	err = dependencies.Identity.SetAuthorization(ctx.Context(), id, request.Allowed, request.ActorPlayerID, request.Reason)
	if err != nil {
		return identityError(err)
	}
	record, found, err := dependencies.Players.FindByID(ctx.Context(), id)
	if err != nil {
		return playerError(err)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "player not found")
	}
	if err = projectIdentity(ctx.Context(), dependencies, record); err != nil {
		return err
	}
	return ctx.JSON(playerResponse(record))
}

// changeName commits and projects one attributed administrative rename.
func changeName(dependencies RemainingDependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := remainingPlayerID(ctx)
		if err != nil {
			return err
		}
		var request AdminNameChangeRequest
		if err = ctx.BodyParser(&request); err != nil || dependencies.Identity == nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid administrative name-change request")
		}
		result, err := dependencies.Identity.RenameAdmin(ctx.Context(), id, request.Username, request.ActorPlayerID, request.Reason)
		if err != nil {
			return identityError(err)
		}
		record, found, err := dependencies.Players.FindByID(ctx.Context(), id)
		if err != nil {
			return playerError(err)
		}
		if !found {
			return fiber.NewError(fiber.StatusNotFound, "player not found")
		}
		if err = projectIdentity(ctx.Context(), dependencies, record); err != nil {
			return err
		}
		if err = projectRoomName(ctx.Context(), dependencies, id, result.NewUsername); err != nil {
			return err
		}
		return ctx.JSON(playerResponse(record))
	}
}

// readNameChanges returns bounded recent identity history.
func readNameChanges(dependencies RemainingDependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, err := remainingPlayerID(ctx)
		if err != nil {
			return err
		}
		limit, err := nameChangeLimit(ctx.Query("limit"))
		if err != nil || dependencies.Identity == nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid name-change history request")
		}
		changes, err := dependencies.Identity.NameChanges(ctx.Context(), id, limit)
		if err != nil {
			return identityError(err)
		}
		response := make([]NameChangeResponse, 0, len(changes))
		for _, change := range changes {
			response = append(response, nameChangeResponse(change))
		}
		return ctx.JSON(response)
	}
}

// projectIdentity replaces and sends one online durable snapshot.
func projectIdentity(ctx context.Context, dependencies RemainingDependencies, record playerservice.Record) error {
	if dependencies.Live == nil {
		return nil
	}
	player, found := dependencies.Live.Find(record.Player.ID)
	if !found {
		return nil
	}
	snapshot := playerlive.SnapshotFromRecord(record)
	if err := player.ReplaceSnapshot(snapshot); err != nil {
		return err
	}
	if dependencies.Connections == nil {
		return nil
	}
	peer := player.Peer()
	connection, found := dependencies.Connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if !found {
		return nil
	}
	packet, err := playerprofile.EncodeInfo(snapshot)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// projectRoomName replaces and broadcasts one active room occupant name.
func projectRoomName(ctx context.Context, dependencies RemainingDependencies, playerID int64, username string) error {
	if dependencies.Rooms == nil {
		return nil
	}
	active, found := dependencies.Rooms.FindByPlayer(playerID)
	if !found || !active.UpdateOccupantName(playerID, username) {
		return nil
	}
	unit, found := active.Unit(playerID)
	if !found {
		return nil
	}
	packet, err := outroomname.Encode(int32(playerID), int32(unit.UnitID), username)
	if err != nil {
		return err
	}
	return roombroadcast.RoomPacket(ctx, dependencies.Connections, active, packet, 0)
}
