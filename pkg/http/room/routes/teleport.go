package routes

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	moderationpolicy "github.com/niflaot/pixels/internal/realm/moderation/policy"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	"github.com/niflaot/pixels/pkg/http/adminaction"
)

// teleportHandler forwards one live player to a target room.
func teleportHandler(rooms roomservice.Manager, players *playerlive.Registry, connections *netconn.Registry, entry *roomentry.Service, actions *adminaction.Service) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, request, err := teleportInput(ctx)
		if err != nil {
			return err
		}
		work := func(actionCtx context.Context) error {
			return teleportPlayer(actionCtx, rooms, players, connections, entry, playerID, request)
		}
		if actions != nil {
			err = actions.Execute(ctx.Context(), adminaction.Request{
				Action: "player.teleport", ActorPlayerID: request.ActorPlayerID,
				Node: moderationpolicy.RoomOverride, Reason: request.Reason,
				TargetPlayerID: playerID,
				Before: func(context.Context) (any, error) {
					return playerRoomState(players, playerID), nil
				},
				After: func(context.Context) (any, error) {
					return map[string]any{"roomId": request.TargetRoomID, "bypass": request.Bypass}, nil
				},
			}, work)
		} else {
			err = work(ctx.Context())
		}
		if err != nil {
			return adminaction.HTTPError(err)
		}

		return ctx.JSON(ActionResponse{Matched: 1, Sent: 1})
	}
}

// playerRoomState returns one live player's current room for auditing.
func playerRoomState(players *playerlive.Registry, playerID int64) map[string]any {
	player, found := players.Find(playerID)
	if !found {
		return map[string]any{"online": false}
	}
	roomID, inRoom := player.CurrentRoom()
	return map[string]any{"online": true, "inRoom": inRoom, "roomId": roomID}
}

// teleportPlayer validates the destination and forwards one live session.
func teleportPlayer(ctx context.Context, rooms roomservice.Manager, players *playerlive.Registry, connections *netconn.Registry, entry *roomentry.Service, playerID int64, request TeleportRequest) error {
	target, found, err := rooms.FindByID(ctx, request.TargetRoomID)
	if err != nil {
		return err
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "target room not found")
	}
	player, found := players.Find(playerID)
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "live player not found")
	}
	peer := player.Peer()
	connection, found := connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "player connection not found")
	}
	if request.Bypass && target.DoorMode != roommodel.DoorModeOpen &&
		(entry == nil || !entry.GrantTrusted(playerID, request.TargetRoomID)) {
		return fiber.NewError(fiber.StatusInternalServerError, "room entry bypass could not be granted")
	}
	packet, err := outforward.Encode(int32(request.TargetRoomID))
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// teleportInput parses one player teleport request.
func teleportInput(ctx *fiber.Ctx) (int64, TeleportRequest, error) {
	playerID, err := strconv.ParseInt(ctx.Params("playerId"), 10, 64)
	if err != nil || playerID <= 0 {
		return 0, TeleportRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid player id")
	}
	var request TeleportRequest
	if err := ctx.BodyParser(&request); err != nil {
		return 0, TeleportRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid room teleport request body")
	}
	if request.TargetRoomID <= 0 || request.TargetRoomID > math.MaxInt32 {
		return 0, TeleportRequest{}, fiber.NewError(fiber.StatusBadRequest, "invalid target room id")
	}
	request.Reason = strings.TrimSpace(request.Reason)

	return playerID, request, nil
}
