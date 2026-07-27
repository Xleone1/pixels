package routes

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	currencyrealm "github.com/niflaot/pixels/internal/realm/inventory/currency"
	"github.com/niflaot/pixels/pkg/http/adminaction"
)

// walletHandler returns one persistent player's configured wallet.
func walletHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if err := authorizeRead(ctx, dependencies); err != nil {
			return err
		}
		playerID, err := playerIDQuery(ctx)
		if err != nil {
			return err
		}
		if err := requirePlayer(ctx.Context(), dependencies, playerID); err != nil {
			return err
		}

		balances, err := dependencies.Currencies.Wallet(ctx.Context(), playerID)
		if err != nil {
			return fmt.Errorf("read player %d currency wallet: %w", playerID, err)
		}

		return ctx.JSON(WalletResponse{PlayerID: playerID, Balances: balanceResponses(balances)})
	}
}

// typesHandler returns configured currency definitions.
func typesHandler(dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if err := authorizeRead(ctx, dependencies); err != nil {
			return err
		}
		definitions, err := dependencies.Currencies.Types(ctx.Context())
		if err != nil {
			return fmt.Errorf("read currency types: %w", err)
		}

		return ctx.JSON(TypesResponse{Types: typeResponses(definitions)})
	}
}

// authorizeRead validates the optional production actor boundary.
func authorizeRead(ctx *fiber.Ctx, dependencies Dependencies) error {
	if dependencies.AdminActions == nil {
		return nil
	}
	actorID, err := strconv.ParseInt(ctx.Get("X-Actor-Player-ID"), 10, 64)
	if err != nil || actorID <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "X-Actor-Player-ID is required")
	}
	return adminaction.HTTPError(dependencies.AdminActions.Authorize(
		ctx.Context(),
		actorID,
		currencyrealm.AdminManage,
	))
}

// requirePlayer verifies a persistent player identity.
func requirePlayer(ctx context.Context, dependencies Dependencies, playerID int64) error {
	_, found, err := dependencies.Finder.FindByID(ctx, playerID)
	if err != nil {
		return fmt.Errorf("find currency player %d: %w", playerID, err)
	}
	if !found {
		return fiber.NewError(fiber.StatusNotFound, "currency player not found")
	}

	return nil
}
