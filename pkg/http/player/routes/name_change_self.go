package routes

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
)

// checkOwnNameChange validates and reserves one self-service username.
func checkOwnNameChange(dependencies RemainingDependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, request, err := selfNameChangeRequest(ctx, dependencies)
		if err != nil {
			return err
		}
		result, err := dependencies.Identity.Check(
			ctx.Context(),
			id,
			request.Username,
		)
		if err != nil {
			return err
		}
		return ctx.JSON(nameCheckResponse(result))
	}
}

// confirmOwnNameChange consumes one reservation and projects the rename.
func confirmOwnNameChange(dependencies RemainingDependencies) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id, request, err := selfNameChangeRequest(ctx, dependencies)
		if err != nil {
			return err
		}
		result, err := dependencies.Identity.Rename(
			ctx.Context(),
			id,
			request.Username,
		)
		if err != nil {
			if code, expected := selfNameChangeErrorCode(err); expected {
				return ctx.JSON(NameCheckResponse{
					Code:        code,
					Username:    strings.TrimSpace(request.Username),
					Suggestions: []string{},
				})
			}
			return err
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
		if err = projectRoomName(
			ctx.Context(),
			dependencies,
			id,
			result.NewUsername,
		); err != nil {
			return err
		}
		return ctx.JSON(NameCheckResponse{
			Code:        playeridentity.ResultAvailable,
			Username:    result.NewUsername,
			Suggestions: []string{},
		})
	}
}

// selfNameChangeRequest parses one authenticated private command target.
func selfNameChangeRequest(
	ctx *fiber.Ctx,
	dependencies RemainingDependencies,
) (int64, SelfNameChangeRequest, error) {
	id, err := remainingPlayerID(ctx)
	if err != nil {
		return 0, SelfNameChangeRequest{}, err
	}
	var request SelfNameChangeRequest
	if err = ctx.BodyParser(&request); err != nil ||
		dependencies.Identity == nil ||
		strings.TrimSpace(request.Username) == "" {
		return 0, SelfNameChangeRequest{},
			fiber.NewError(fiber.StatusBadRequest, "invalid self-service name-change request")
	}
	return id, request, nil
}

// selfNameChangeErrorCode maps expected commit failures to protocol decisions.
func selfNameChangeErrorCode(err error) (int32, bool) {
	switch {
	case errors.Is(err, playeridentity.ErrRenameDisabled):
		return playeridentity.ResultDisabled, true
	case errors.Is(err, playeridentity.ErrUsernameTaken),
		errors.Is(err, playeridentity.ErrReservationMissing):
		return playeridentity.ResultTaken, true
	default:
		return playeridentity.ResultInvalid, false
	}
}
