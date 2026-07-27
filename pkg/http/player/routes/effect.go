package routes

import (
	"context"
	"errors"
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	playerrealm "github.com/niflaot/pixels/internal/realm/player"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	"github.com/niflaot/pixels/pkg/http/adminaction"
)

// EffectRequest contains one administrative effect grant.
type EffectRequest struct {
	// ActorPlayerID identifies the administrative actor.
	ActorPlayerID int64 `json:"actorPlayerId"`
	// Reason explains the effect mutation.
	Reason string `json:"reason"`
	// EffectID identifies the Nitro effect.
	EffectID int32 `json:"effectId"`
	// DurationSeconds stores one charge duration; zero means permanent.
	DurationSeconds int32 `json:"durationSeconds"`
	// Enable selects the granted effect immediately and defaults to true.
	Enable *bool `json:"enable"`
}

// EffectResponse contains the resulting effect stack.
type EffectResponse struct {
	// PlayerID identifies the effect owner.
	PlayerID int64 `json:"playerId"`
	// EffectID identifies the Nitro effect.
	EffectID int32 `json:"effectId"`
	// DurationSeconds stores one charge duration.
	DurationSeconds int32 `json:"durationSeconds"`
	// RemainingCharges stores the stack count.
	RemainingCharges int32 `json:"remainingCharges"`
	// Enabled reports whether the effect was selected immediately.
	Enabled bool `json:"enabled"`
}

// grantEffect grants one effect charge to a player.
func (handler handler) grantEffect(ctx *fiber.Ctx) error {
	playerID, err := positivePathID(ctx, "playerId")
	if err != nil {
		return err
	}
	var request EffectRequest
	if err = ctx.BodyParser(&request); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid player effect request body")
	}
	if handler.effects == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "player effects are unavailable")
	}
	enabled := request.Enable == nil || *request.Enable
	var effect playereffect.Effect
	work := func(actionCtx context.Context) error {
		if enabled {
			effect, err = handler.effects.GrantEnabled(actionCtx, playerID, request.EffectID, request.DurationSeconds, playereffect.SourceAdmin)
		} else {
			effect, err = handler.effects.Grant(actionCtx, playerID, request.EffectID, request.DurationSeconds, playereffect.SourceAdmin)
		}
		return err
	}
	if handler.adminActions != nil {
		err = handler.adminActions.Execute(ctx.Context(), adminaction.Request{
			Action: "player.effect.grant", ActorPlayerID: request.ActorPlayerID,
			Node: playerrealm.AdminEffectGrant, Reason: request.Reason,
			TargetPlayerID: playerID,
			Before: func(actionCtx context.Context) (any, error) {
				return handler.effectState(actionCtx, playerID, request.EffectID)
			},
			After: func(actionCtx context.Context) (any, error) {
				return handler.effectState(actionCtx, playerID, request.EffectID)
			},
		}, work)
	} else {
		err = work(ctx.Context())
	}
	if err != nil {
		return effectError(adminaction.HTTPError(err))
	}
	return ctx.JSON(EffectResponse{PlayerID: effect.PlayerID, EffectID: effect.ID, DurationSeconds: effect.DurationSeconds, RemainingCharges: effect.RemainingCharges, Enabled: enabled})
}

// revokeEffect removes one player's durable effect stack.
func (handler handler) revokeEffect(ctx *fiber.Ctx) error {
	playerID, err := positivePathID(ctx, "playerId")
	if err != nil {
		return err
	}
	effectID, err := positivePathID(ctx, "effectId")
	if err != nil || effectID > math.MaxInt32 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid effect id")
	}
	if handler.effects == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "player effects are unavailable")
	}
	var request EffectRequest
	if handler.adminActions != nil {
		if err = ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid player effect request body")
		}
		request.Reason = strings.TrimSpace(request.Reason)
		err = handler.adminActions.Execute(ctx.Context(), adminaction.Request{
			Action: "player.effect.revoke", ActorPlayerID: request.ActorPlayerID,
			Node: playerrealm.AdminEffectGrant, Reason: request.Reason,
			TargetPlayerID: playerID,
			Before: func(actionCtx context.Context) (any, error) {
				return handler.effectState(actionCtx, playerID, int32(effectID))
			},
			After: func(actionCtx context.Context) (any, error) {
				return handler.effectState(actionCtx, playerID, int32(effectID))
			},
		}, func(actionCtx context.Context) error {
			return handler.effects.Revoke(actionCtx, playerID, int32(effectID))
		})
	} else {
		err = handler.effects.Revoke(ctx.Context(), playerID, int32(effectID))
	}
	if err != nil {
		return effectError(adminaction.HTTPError(err))
	}
	return ctx.SendStatus(fiber.StatusNoContent)
}

// effectState returns one compact durable effect state.
func (handler handler) effectState(ctx context.Context, playerID int64, effectID int32) (any, error) {
	effects, err := handler.effects.List(ctx, playerID)
	if err != nil {
		return nil, err
	}
	for _, effect := range effects {
		if effect.ID == effectID {
			return map[string]any{
				"effectId": effect.ID, "durationSeconds": effect.DurationSeconds,
				"remainingCharges": effect.RemainingCharges,
			}, nil
		}
	}
	return map[string]any{"effectId": effectID, "present": false}, nil
}

// positivePathID parses one positive route identifier.
func positivePathID(ctx *fiber.Ctx, name string) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, fiber.NewError(fiber.StatusBadRequest, "invalid "+name)
	}
	return id, nil
}

// effectError maps domain effect failures into meaningful HTTP errors.
func effectError(err error) error {
	if errors.Is(err, playereffect.ErrInvalidEffect) {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if errors.Is(err, playereffect.ErrEffectNotFound) {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return err
}
