// Package routes contains protected subscription administration routes.
package routes

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	subaccess "github.com/niflaot/pixels/internal/realm/subscription/access"
	subadmin "github.com/niflaot/pixels/internal/realm/subscription/admin"
	"github.com/niflaot/pixels/internal/realm/subscription/core"
	subrecord "github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/pkg/http/adminaction"
	"go.uber.org/fx"
)

const (
	// basePath stores the subscription administration prefix.
	basePath = "/api/admin/subscriptions"
)

// Dependencies contains subscription administration collaborators.
type Dependencies struct {
	fx.In

	// Subscriptions manages administrative subscription behavior.
	Subscriptions *subadmin.Service
	// AdminActions authorizes and audits membership mutations.
	AdminActions *adminaction.Service `optional:"true"`
}

// Register registers protected subscription administration routes.
func Register(app *fiber.App, dependencies Dependencies) {
	if dependencies.Subscriptions == nil {
		return
	}
	app.Get(basePath+"/club-offers", listClubOffers(dependencies))
	app.Post(basePath+"/club-offers", saveClubOffer(dependencies, false))
	app.Patch(basePath+"/club-offers/:id", saveClubOffer(dependencies, true))
	app.Get(basePath+"/targeted-offers", listTargetedOffers(dependencies))
	app.Post(basePath+"/targeted-offers", saveTargetedOffer(dependencies, false))
	app.Patch(basePath+"/targeted-offers/:id", saveTargetedOffer(dependencies, true))
	app.Get(basePath+"/calendar/campaigns", listCampaigns(dependencies))
	app.Post(basePath+"/calendar/campaigns", saveCampaign(dependencies, false))
	app.Patch(basePath+"/calendar/campaigns/:id", saveCampaign(dependencies, true))
	app.Get(basePath+"/:playerId", membership(dependencies))
	app.Post(basePath+"/:playerId/grant", grant(dependencies))
	app.Delete(basePath+"/:playerId", revoke(dependencies))
}

// membership returns membership and payday history.
func membership(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if dependencies.AdminActions != nil {
			actorID, parseErr := strconv.ParseInt(ctx.Get("X-Actor-Player-ID"), 10, 64)
			if parseErr != nil || actorID <= 0 {
				return fiber.NewError(fiber.StatusBadRequest, "X-Actor-Player-ID is required")
			}
			if authorizeErr := dependencies.AdminActions.Authorize(ctx.Context(), actorID, subaccess.MembershipGrant); authorizeErr != nil {
				return adminaction.HTTPError(authorizeErr)
			}
		}
		playerID, err := identifier(ctx, "playerId")
		if err != nil {
			return err
		}
		membership, payday, paydays, found, err := dependencies.Subscriptions.Membership(ctx.Context(), playerID)
		if err != nil {
			return err
		}
		if !found {
			return fiber.NewError(fiber.StatusNotFound, "subscription membership not found")
		}
		return ctx.JSON(fiber.Map{"membership": membership, "paydayProjection": payday,
			"giftsAvailable": core.RemainingClubGifts(membership), "paydays": paydays})
	}
}

// grant grants or extends one membership.
func grant(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := identifier(ctx, "playerId")
		if err != nil {
			return err
		}
		var request GrantRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid membership grant request")
		}
		var membership subrecord.Membership
		work := func(actionCtx context.Context) error {
			var grantErr error
			membership, grantErr = dependencies.Subscriptions.Grant(actionCtx, playerID, request.Level, time.Duration(request.DurationSeconds)*time.Second)
			return grantErr
		}
		if dependencies.AdminActions != nil {
			err = dependencies.AdminActions.Execute(ctx.Context(), adminaction.Request{
				Action:        "subscription.membership.grant",
				ActorPlayerID: request.ActorPlayerID, Node: subaccess.MembershipGrant,
				Reason: request.Reason, TargetPlayerID: playerID,
				Before: dependencies.membershipSnapshot(playerID),
				After:  dependencies.membershipSnapshot(playerID),
			}, work)
		} else {
			err = work(ctx.Context())
		}
		if err != nil {
			return routeError(adminaction.HTTPError(err))
		}
		return ctx.JSON(membership)
	}
}

// revoke revokes one membership.
func revoke(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		playerID, err := identifier(ctx, "playerId")
		if err != nil {
			return err
		}
		var request RevokeRequest
		if dependencies.AdminActions != nil {
			if err = ctx.BodyParser(&request); err != nil {
				return fiber.NewError(fiber.StatusBadRequest, "invalid membership revoke request")
			}
			err = dependencies.AdminActions.Execute(ctx.Context(), adminaction.Request{
				Action:        "subscription.membership.revoke",
				ActorPlayerID: request.ActorPlayerID, Node: subaccess.MembershipGrant,
				Reason: request.Reason, TargetPlayerID: playerID,
				Before: dependencies.membershipSnapshot(playerID),
				After:  dependencies.membershipSnapshot(playerID),
			}, func(actionCtx context.Context) error {
				return dependencies.Subscriptions.Revoke(actionCtx, playerID)
			})
		} else {
			err = dependencies.Subscriptions.Revoke(ctx.Context(), playerID)
		}
		if err != nil {
			return routeError(adminaction.HTTPError(err))
		}
		return ctx.SendStatus(fiber.StatusNoContent)
	}
}

// identifier parses one positive route identifier.
func identifier(ctx *fiber.Ctx, name string) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid subscription record id")
	}
	return id, nil
}

// routeError maps subscription administration failures.
func routeError(err error) error {
	if errors.Is(err, subadmin.ErrInvalidInput) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if errors.Is(err, subadmin.ErrNotFound) {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return err
}
