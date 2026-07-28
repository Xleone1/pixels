package moderation

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	bannedevent "github.com/niflaot/pixels/internal/realm/room/control/events/banned"
	kickedevent "github.com/niflaot/pixels/internal/realm/room/control/events/kicked"
	mutedevent "github.com/niflaot/pixels/internal/realm/room/control/events/muted"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
	"github.com/niflaot/pixels/pkg/bus"
)

// Nodes stores room moderation permission nodes.
type Nodes struct {
	// OwnKick allows locally authorized kicking.
	OwnKick permission.Node
	// OwnMute allows locally authorized muting.
	OwnMute permission.Node
	// OwnBan allows locally authorized banning.
	OwnBan permission.Node
	// AnyKick allows staff kicking in any room.
	AnyKick permission.Node
	// AnyMute allows staff muting in any room.
	AnyMute permission.Node
	// AnyBan allows staff banning in any room.
	AnyBan permission.Node
	// Unkickable protects moderation targets.
	Unkickable permission.Node
}

// EventDispatcher intercepts authorized room moderation actions.
type EventDispatcher interface {
	// DispatchRoomModerationAction reports whether one action was vetoed.
	DispatchRoomModerationAction(context.Context, string, int64, int64, int64) bool
}

// Service coordinates room moderation policy and state.
type Service struct {
	// config stores normalized moderation limits.
	config Config
	// store persists current moderation state.
	store Store
	// rooms resolves durable room policy.
	rooms RoomFinder
	// rights resolves room-scoped capabilities.
	rights RightsChecker
	// permissions resolves global capability nodes.
	permissions permissionservice.Checker
	// events publishes committed moderation changes.
	events bus.Publisher
	// nodes stores moderation capability nodes.
	nodes Nodes
	// now returns current time for deterministic behavior.
	now func() time.Time
	// pluginEvents intercepts moderation before mutation.
	pluginEvents EventDispatcher
}

// SetPluginRuntime installs the optional dynamic-plugin moderation interceptor.
func (service *Service) SetPluginRuntime(events EventDispatcher) {
	service.pluginEvents = events
}

// New creates a room moderation service.
func New(config Config, store Store, rooms RoomFinder, rights RightsChecker, permissions permissionservice.Checker, events bus.Publisher, nodes Nodes) *Service {
	return &Service{config: config.Normalize(), store: store, rooms: rooms, rights: rights, permissions: permissions, events: events, nodes: nodes, now: time.Now}
}

// Kick commits an immediate room kick action.
func (service *Service) Kick(ctx context.Context, roomID int64, actorID int64, targetID int64) error {
	if _, err := service.authorize(ctx, roomID, actorID, targetID, moderationmodel.ActionKick); err != nil {
		return err
	}
	if service.intercept(ctx, moderationmodel.ActionKick, roomID, actorID, targetID) {
		return ErrCancelledByPlugin
	}

	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		return service.publish(txCtx, kickedevent.Name, kickedevent.Payload{RoomID: roomID, TargetPlayerID: targetID, ActorID: actorID})
	})
}

// Mute creates or replaces a room mute.
func (service *Service) Mute(ctx context.Context, roomID int64, actorID int64, targetID int64, minutes int32) error {
	if minutes < service.config.MinMuteMinutes || minutes > service.config.MaxMuteMinutes {
		return ErrInvalidMuteDuration
	}
	if _, err := service.authorize(ctx, roomID, actorID, targetID, moderationmodel.ActionMute); err != nil {
		return err
	}
	if service.intercept(ctx, moderationmodel.ActionMute, roomID, actorID, targetID) {
		return ErrCancelledByPlugin
	}
	duration := time.Duration(minutes) * time.Minute
	expiresAt := service.now().Add(duration)

	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := service.store.Mute(txCtx, roomID, targetID, expiresAt); err != nil {
			return err
		}
		payload := mutedevent.Payload{RoomID: roomID, TargetPlayerID: targetID, ActorID: actorID, DurationSeconds: int64(duration / time.Second), ExpiresAt: expiresAt}

		return service.publish(txCtx, mutedevent.Name, payload)
	})
}

// SystemKick publishes a non-discretionary room-only kick without user authority checks.
func (service *Service) SystemKick(ctx context.Context, roomID int64, targetID int64) error {
	if roomID <= 0 || targetID <= 0 {
		return ErrInvalidIdentity
	}
	if err := service.authorizeSystemTarget(ctx, roomID, targetID); err != nil {
		return err
	}
	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		return service.publish(txCtx, kickedevent.Name, kickedevent.Payload{RoomID: roomID, TargetPlayerID: targetID})
	})
}

// SystemMute applies a bounded room mute without user authority checks.
func (service *Service) SystemMute(ctx context.Context, roomID int64, targetID int64, minutes int32) error {
	if roomID <= 0 || targetID <= 0 {
		return ErrInvalidIdentity
	}
	if minutes < service.config.MinMuteMinutes || minutes > service.config.MaxMuteMinutes {
		return ErrInvalidMuteDuration
	}
	if err := service.authorizeSystemTarget(ctx, roomID, targetID); err != nil {
		return err
	}
	duration := time.Duration(minutes) * time.Minute
	expiresAt := service.now().Add(duration)
	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := service.store.Mute(txCtx, roomID, targetID, expiresAt); err != nil {
			return err
		}
		payload := mutedevent.Payload{RoomID: roomID, TargetPlayerID: targetID, DurationSeconds: int64(duration / time.Second), ExpiresAt: expiresAt}
		return service.publish(txCtx, mutedevent.Name, payload)
	})
}

// Unmute ends an active room mute.
func (service *Service) Unmute(ctx context.Context, roomID int64, actorID int64, targetID int64) error {
	if _, err := service.authorize(ctx, roomID, actorID, targetID, moderationmodel.ActionMute); err != nil {
		return err
	}
	if service.intercept(ctx, moderationmodel.ActionUnmute, roomID, actorID, targetID) {
		return ErrCancelledByPlugin
	}

	return service.end(ctx, roomID, actorID, targetID, moderationmodel.ActionUnmute)
}

// Ban creates or replaces a room ban.
func (service *Service) Ban(ctx context.Context, roomID int64, actorID int64, targetID int64, banDuration moderationmodel.BanDuration) error {
	duration, valid := banDuration.Duration()
	if !valid {
		return ErrInvalidBanDuration
	}
	if _, err := service.authorize(ctx, roomID, actorID, targetID, moderationmodel.ActionBan); err != nil {
		return err
	}
	if service.intercept(ctx, moderationmodel.ActionBan, roomID, actorID, targetID) {
		return ErrCancelledByPlugin
	}
	expiresAt := service.now().Add(duration)

	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := service.store.Ban(txCtx, roomID, targetID, expiresAt); err != nil {
			return err
		}
		payload := bannedevent.Payload{RoomID: roomID, TargetPlayerID: targetID, ActorID: actorID, DurationSeconds: int64(duration / time.Second), ExpiresAt: expiresAt}

		return service.publish(txCtx, bannedevent.Name, payload)
	})
}

// Unban ends an active room ban.
func (service *Service) Unban(ctx context.Context, roomID int64, actorID int64, targetID int64) error {
	if _, err := service.authorize(ctx, roomID, actorID, targetID, moderationmodel.ActionBan); err != nil {
		return err
	}
	if service.intercept(ctx, moderationmodel.ActionUnban, roomID, actorID, targetID) {
		return ErrCancelledByPlugin
	}

	return service.end(ctx, roomID, actorID, targetID, moderationmodel.ActionUnban)
}

// intercept reports whether a plugin vetoes one authorized action.
func (service *Service) intercept(ctx context.Context, action moderationmodel.Action, roomID int64, actorID int64, targetID int64) bool {
	return service.pluginEvents != nil && service.pluginEvents.DispatchRoomModerationAction(ctx, string(action), roomID, actorID, targetID)
}
