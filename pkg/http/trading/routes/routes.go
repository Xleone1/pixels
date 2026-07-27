// Package routes exposes protected Marketplace and direct-trade administration.
package routes

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	tradeadmin "github.com/niflaot/pixels/internal/realm/trade/admin"
	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	"github.com/niflaot/pixels/pkg/http/adminaction"
	"go.uber.org/fx"
)

// AuditResponse contains one completed direct trade.
type AuditResponse struct {
	// ID identifies the audit row.
	ID int64 `json:"id"`
	// RoomID identifies the settlement room.
	RoomID int64 `json:"roomId"`
	// FirstPlayerID identifies the first participant.
	FirstPlayerID int64 `json:"firstPlayerId"`
	// SecondPlayerID identifies the second participant.
	SecondPlayerID int64 `json:"secondPlayerId"`
	// FirstIP stores the first participant audit address.
	FirstIP string `json:"firstIp,omitempty"`
	// SecondIP stores the second participant audit address.
	SecondIP string `json:"secondIp,omitempty"`
	// FirstItemIDs stores the first offer.
	FirstItemIDs []int64 `json:"firstItemIds"`
	// SecondItemIDs stores the second offer.
	SecondItemIDs []int64 `json:"secondItemIds"`
	// FirstRedeemableCredits stores value delivered to the second participant.
	FirstRedeemableCredits int64 `json:"firstRedeemableCredits"`
	// SecondRedeemableCredits stores value delivered to the first participant.
	SecondRedeemableCredits int64 `json:"secondRedeemableCredits"`
	// CreatedAt stores settlement time.
	CreatedAt time.Time `json:"createdAt"`
}

// Dependencies contains trading administration behavior.
type Dependencies struct {
	// In marks dependencies for Fx injection.
	fx.In
	// Marketplace manages protected listing lifecycle operations.
	Marketplace *marketcore.Service
	// Trade manages locks and audit reads.
	Trade *tradeadmin.Service
	// Sanctions owns the superseding global trade-lock records.
	Sanctions *sanctioncore.Service
	// AdminActions authorizes and audits trading administration.
	AdminActions *adminaction.Service `optional:"true"`
}

// Register mounts protected trading administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	if dependencies.Trade != nil {
		app.Get("/api/admin/trade/players/:playerId/log", dependencies.logs)
		app.Post("/api/admin/trade/players/:playerId/lock", dependencies.lock)
		app.Delete("/api/admin/trade/players/:playerId/lock", dependencies.unlock)
	}
	if dependencies.Marketplace != nil {
		app.Post("/api/admin/marketplace/listings/:id/force-close", dependencies.forceClose)
	}
}

// playerID parses a positive path identifier.
func playerID(ctx *fiber.Ctx, name string) (int64, error) {
	value, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || value <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid path identifier")
	}
	return value, nil
}

// logs returns recent completed trades involving a player.
func (dependencies Dependencies) logs(ctx *fiber.Ctx) error {
	id, err := playerID(ctx, "playerId")
	if err != nil {
		return err
	}
	if dependencies.AdminActions != nil {
		actorID, parseErr := strconv.ParseInt(ctx.Get("X-Actor-Player-ID"), 10, 64)
		if parseErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "X-Actor-Player-ID is required")
		}
		if authorizeErr := dependencies.AdminActions.Authorize(ctx.Context(), actorID, tradecore.ModerationLock); authorizeErr != nil {
			return adminaction.HTTPError(authorizeErr)
		}
	}
	values, err := dependencies.Trade.Logs(ctx.Context(), id)
	if err != nil {
		return err
	}
	items := make([]AuditResponse, len(values))
	for index, value := range values {
		items[index] = AuditResponse{ID: value.ID, RoomID: value.RoomID, FirstPlayerID: value.FirstPlayerID, SecondPlayerID: value.SecondPlayerID, FirstIP: value.FirstIP, SecondIP: value.SecondIP, FirstItemIDs: value.FirstItemIDs, SecondItemIDs: value.SecondItemIDs, FirstRedeemableCredits: value.FirstRedeemableCredits, SecondRedeemableCredits: value.SecondRedeemableCredits, CreatedAt: value.CreatedAt}
	}
	return ctx.JSON(fiber.Map{"items": items, "count": len(items)})
}

// lock disables direct trading for a player.
func (dependencies Dependencies) lock(ctx *fiber.Ctx) error { return dependencies.setLock(ctx, true) }

// unlock enables direct trading for a player.
func (dependencies Dependencies) unlock(ctx *fiber.Ctx) error {
	return dependencies.setLock(ctx, false)
}

// setLock applies one durable trade lock.
func (dependencies Dependencies) setLock(ctx *fiber.Ctx, locked bool) error {
	id, err := playerID(ctx, "playerId")
	if err != nil {
		return err
	}
	if dependencies.Sanctions == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "global sanction service unavailable")
	}
	var request struct {
		ActorPlayerID int64  `json:"actorPlayerId"`
		Reason        string `json:"reason"`
	}
	if dependencies.AdminActions != nil {
		if err = ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid trade lock request")
		}
		err = dependencies.AdminActions.Execute(ctx.Context(), adminaction.Request{
			Action: "trade.lock", ActorPlayerID: request.ActorPlayerID,
			Node: tradecore.ModerationLock, Reason: request.Reason,
			TargetPlayerID: id,
			Before:         dependencies.tradeLockSnapshot(id),
			After:          dependencies.tradeLockSnapshot(id),
		}, func(actionCtx context.Context) error {
			return dependencies.applyLock(actionCtx, id, locked, request)
		})
		if err != nil {
			return adminaction.HTTPError(err)
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
	if err = dependencies.applyLock(ctx.Context(), id, locked, request); err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// applyLock applies or revokes the global direct-trade sanction.
func (dependencies Dependencies) applyLock(ctx context.Context, id int64, locked bool, request struct {
	ActorPlayerID int64  `json:"actorPlayerId"`
	Reason        string `json:"reason"`
}) error {
	if locked {
		issuerID := request.ActorPlayerID
		issuerKind := "system"
		var issuer *int64
		if issuerID > 0 {
			issuerKind = "player"
			issuer = &issuerID
		}
		_, err := dependencies.Sanctions.Apply(ctx, sanctionrecord.ApplyParams{ReceiverPlayerID: id, IssuerPlayerID: issuer, IssuerKind: issuerKind, Kind: sanctionrecord.KindTradeLock, Reason: request.Reason, Source: "admin_http"})
		return err
	}
	history, err := dependencies.Sanctions.History(ctx, id, 500)
	if err != nil {
		return err
	}
	for _, punishment := range history {
		if punishment.Kind == sanctionrecord.KindTradeLock && punishment.ActiveAt(time.Now()) && (punishment.Source == "admin_http" || strings.HasPrefix(punishment.Reason, "Migrated legacy")) {
			if _, err = dependencies.Sanctions.RevokeSystem(ctx, punishment.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// forceClose returns one open Marketplace item to its seller.
func (dependencies Dependencies) forceClose(ctx *fiber.Ctx) error {
	id, err := playerID(ctx, "id")
	if err != nil {
		return err
	}
	err = dependencies.Marketplace.Close(ctx.Context(), id, 0, true)
	if errors.Is(err, marketcore.ErrListingUnavailable) {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	if err != nil {
		return err
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}
