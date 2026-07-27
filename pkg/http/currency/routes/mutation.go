package routes

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	currencyrealm "github.com/niflaot/pixels/internal/realm/inventory/currency"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	"github.com/niflaot/pixels/pkg/http/adminaction"
)

// mutationAction names one administrative balance operation.
type mutationAction string

const (
	// grantAction adds a positive amount.
	grantAction mutationAction = "grant"

	// deductAction subtracts a positive amount.
	deductAction mutationAction = "deduct"

	// setAction replaces the absolute balance.
	setAction mutationAction = "set"
)

// mutationHandler commits one administrative currency mutation.
func mutationHandler(action mutationAction, dependencies Dependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		input, err := parseMutationInput(ctx, action)
		if err != nil {
			return err
		}
		if err := requirePlayer(ctx.Context(), dependencies, input.playerID); err != nil {
			return err
		}
		if dependencies.AdminActions == nil && input.request.Reason == "" {
			input.request.Reason = "admin_api_" + string(action)
		}

		var amount int64
		var previousAmount int64
		work := func(txCtx context.Context) error {
			var applyErr error
			amount, applyErr = action.apply(txCtx, dependencies, input)
			return applyErr
		}
		if dependencies.AdminActions != nil {
			if _, parseErr := uuid.Parse(input.operationKey); parseErr != nil {
				return fiber.NewError(fiber.StatusBadRequest, "valid Idempotency-Key is required")
			}
			err = dependencies.AdminActions.Execute(ctx.Context(), adminaction.Request{
				Action: "currency." + string(action), ActorPlayerID: input.request.ActorPlayerID,
				Node: currencyrealm.AdminManage, Reason: input.request.Reason,
				TargetPlayerID: input.playerID,
				Before: func(actionCtx context.Context) (any, error) {
					var balanceErr error
					previousAmount, balanceErr = dependencies.Currencies.Balance(actionCtx, input.playerID, input.currencyType)
					return currencyState(input.currencyType, previousAmount), balanceErr
				},
				After: func(context.Context) (any, error) {
					return currencyState(input.currencyType, amount), nil
				},
			}, work)
		} else {
			err = work(ctx.Context())
		}
		if err != nil {
			return mutationError(input, adminaction.HTTPError(err))
		}
		alertSent := sendMutationAlert(ctx, dependencies, action, input, amount)

		return ctx.JSON(MutationResponse{
			PlayerID: input.playerID, CurrencyType: input.currencyType, Amount: amount,
			AlertRequested: input.request.Alert, AlertSent: alertSent,
		})
	}
}

// currencyState creates one compact balance audit state.
func currencyState(currencyType int32, amount int64) map[string]any {
	return map[string]any{"currencyType": currencyType, "balance": amount}
}

// validate validates action-specific request amounts.
func (action mutationAction) validate(amount int64) error {
	if action == setAction {
		if amount < 0 {
			return fiber.NewError(fiber.StatusBadRequest, "currency set amount must be non-negative")
		}
		return nil
	}
	if amount <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "currency mutation amount must be positive")
	}

	return nil
}

// apply executes one action through currency management behavior.
func (action mutationAction) apply(ctx context.Context, dependencies Dependencies, input mutationInput) (int64, error) {
	reason := input.request.Reason
	actorID := input.request.ActorPlayerID
	var actorIDPointer *int64
	if actorID > 0 {
		actorIDPointer = &actorID
	}
	if action == setAction {
		return dependencies.Currencies.Set(ctx, currencyservice.SetParams{
			OperationKey: input.operationKey,
			PlayerID:     input.playerID, CurrencyType: input.currencyType, Amount: input.request.Amount,
			Reason: reason, ActorKind: currencyservice.ActorAdmin, ActorID: actorIDPointer,
		})
	}

	delta := input.request.Amount
	if action == deductAction {
		delta = -delta
	}

	return dependencies.Currencies.Grant(ctx, currencyservice.GrantParams{
		OperationKey: input.operationKey,
		PlayerID:     input.playerID, CurrencyType: input.currencyType, Amount: delta,
		Reason: reason, ActorKind: currencyservice.ActorAdmin, ActorID: actorIDPointer,
	})
}

// mutationError maps domain errors into meaningful HTTP failures.
func mutationError(input mutationInput, err error) error {
	switch {
	case errors.Is(err, currencyservice.ErrInvalidCurrencyType):
		return fiber.NewError(fiber.StatusBadRequest, "currency type is not configured")
	case errors.Is(err, currencyservice.ErrInvalidAmount), errors.Is(err, currencyservice.ErrInvalidActor):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	case errors.Is(err, currencyservice.ErrInsufficientBalance):
		return fiber.NewError(fiber.StatusConflict, "currency deduction exceeds player balance")
	case errors.Is(err, currencyservice.ErrIdempotencyConflict):
		return fiber.NewError(fiber.StatusConflict, err.Error())
	default:
		return fmt.Errorf("mutate player %d currency %d: %w", input.playerID, input.currencyType, err)
	}
}
