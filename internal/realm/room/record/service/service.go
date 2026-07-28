package service

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/pkg/bus"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
)

// updateActorKey identifies an optional player-authored settings mutation.
type updateActorKey struct{}

// Service validates and coordinates room persistence behavior.
type Service struct {
	// store reads and writes room persistence records.
	store Store

	// layouts validates room layout references.
	layouts layout.Manager

	// profanity validates user-facing room text when configured.
	profanity ProfanityChecker

	// events publishes committed room record transitions.
	events bus.Publisher

	// pluginEvents emits cancellable pre-persistence room mutations.
	pluginEvents EventDispatcher
}

// EventDispatcher emits cancellable room settings mutations.
type EventDispatcher interface {
	// DispatchRoomUpdate returns possibly replaced settings and veto state.
	DispatchRoomUpdate(context.Context, sdkplayer.Player, int64, UpdateParams) (UpdateParams, bool)
}

// New creates a room service.
func New(store Store, layouts layout.Manager) *Service {
	return &Service{store: store, layouts: layouts}
}

// WithEvents configures optional committed room event publication.
func (service *Service) WithEvents(events bus.Publisher) *Service {
	service.events = events
	return service
}

// SetPluginRuntime installs the optional plugin room event seam.
func (service *Service) SetPluginRuntime(events EventDispatcher) { service.pluginEvents = events }

// WithUpdateActor attaches the immutable player responsible for a settings mutation.
func WithUpdateActor(ctx context.Context, actor sdkplayer.Player) context.Context {
	return context.WithValue(ctx, updateActorKey{}, actor)
}

// updateActor returns the player responsible for a settings mutation.
func updateActor(ctx context.Context) sdkplayer.Player {
	actor, _ := ctx.Value(updateActorKey{}).(sdkplayer.Player)
	return actor
}

// WithProfanity configures optional global content validation.
func (service *Service) WithProfanity(checker ProfanityChecker) *Service {
	service.profanity = checker

	return service
}
