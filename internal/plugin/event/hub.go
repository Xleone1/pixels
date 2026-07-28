// Package event implements typed and cancellable dynamic-plugin events.
package event

import (
	"cmp"
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.uber.org/zap"
)

var (
	// ErrInvalidListener reports an incomplete plugin event registration.
	ErrInvalidListener = errors.New("invalid plugin event listener")
	// ErrEventCancelled reports a cancellable event vetoed by a listener.
	ErrEventCancelled = errors.New("plugin event cancelled")
)

// PlayerFinder resolves immutable connected-player snapshots.
type PlayerFinder interface {
	// Find returns one connected player snapshot.
	Find(int64) (sdkplugin.Player, bool)
}

// listenerRegistration stores one ordered plugin listener.
type listenerRegistration struct {
	// options controls priority and cancellation skipping.
	options sdkevent.ListenerOptions
	// listener executes plugin event behavior.
	listener sdkevent.Listener
	// scope stores plugin identity and health.
	scope *pluginruntime.Scope
	// order preserves registration order across equal priorities.
	order uint64
}

// Hub dispatches typed plugin events without exposing the internal bus.
type Hub struct {
	// mutex protects startup registration and dispatch snapshots.
	mutex sync.RWMutex
	// listeners stores event listeners by stable name.
	listeners map[string][]listenerRegistration
	// order stores the next stable registration sequence.
	order uint64
	// timeout bounds listener callbacks.
	timeout time.Duration
	// log records isolated listener failures.
	log *zap.Logger
	// players resolves immutable player snapshots for realm interceptors.
	players PlayerFinder
}

// SetPlayerFinder installs the shared immutable player resolver.
func (hub *Hub) SetPlayerFinder(players PlayerFinder) {
	hub.mutex.Lock()
	hub.players = players
	hub.mutex.Unlock()
}

// player returns a connected snapshot or an identity-only fallback.
func (hub *Hub) player(playerID int64) sdkplugin.Player {
	hub.mutex.RLock()
	players := hub.players
	hub.mutex.RUnlock()
	if players != nil {
		if player, found := players.Find(playerID); found {
			return player
		}
	}
	return sdkplugin.Player{ID: playerID}
}

// eventCloner creates an isolated callback-owned immutable notification.
type eventCloner interface {
	// CloneEvent returns one independent event value.
	CloneEvent() sdkevent.Event
}

// NewHub creates an empty plugin-facing event dispatcher.
func NewHub(timeout time.Duration, log *zap.Logger) *Hub {
	if log == nil {
		log = zap.NewNop()
	}
	return &Hub{listeners: make(map[string][]listenerRegistration), timeout: timeout, log: log}
}

// Access scopes listener registration to one plugin.
type Access struct {
	// hub stores the shared dispatcher.
	hub *Hub
	// scope stores plugin identity and health.
	scope *pluginruntime.Scope
}

// NewAccess creates one plugin-scoped event registrar.
func NewAccess(hub *Hub, scope *pluginruntime.Scope) *Access {
	return &Access{hub: hub, scope: scope}
}

// Listen registers one listener for the calling plugin.
func (access *Access) Listen(name string, options sdkevent.ListenerOptions, listener sdkevent.Listener) error {
	return access.hub.listen(access.scope, name, options, listener)
}

// listen stores one listener in deterministic priority order.
func (hub *Hub) listen(scope *pluginruntime.Scope, name string, options sdkevent.ListenerOptions, listener sdkevent.Listener) error {
	name = strings.TrimSpace(name)
	if scope == nil || name == "" || listener == nil {
		return ErrInvalidListener
	}
	hub.mutex.Lock()
	defer hub.mutex.Unlock()
	hub.order++
	entries := append(hub.listeners[name], listenerRegistration{options: options, listener: listener, scope: scope, order: hub.order})
	slices.SortStableFunc(entries, func(left listenerRegistration, right listenerRegistration) int {
		if order := cmp.Compare(right.options.Priority, left.options.Priority); order != 0 {
			return order
		}
		return cmp.Compare(left.order, right.order)
	})
	hub.listeners[name] = entries
	return nil
}

// Dispatch invokes every active listener and reports final cancellation.
func (hub *Hub) Dispatch(ctx context.Context, event sdkevent.Event) error {
	if event == nil || strings.TrimSpace(event.Name()) == "" {
		return ErrInvalidListener
	}
	hub.mutex.RLock()
	entries := append([]listenerRegistration(nil), hub.listeners[event.Name()]...)
	hub.mutex.RUnlock()
	for _, entry := range entries {
		if !entry.scope.Enabled() || ignoreCancelled(event, entry.options) {
			continue
		}
		callbackEvent := clone(event)
		err := pluginruntime.InvokeCallback(ctx, hub.timeout, entry.scope, "event "+event.Name(), hub.log, func(callbackContext context.Context) error {
			return entry.listener(callbackContext, callbackEvent)
		})
		if err == nil {
			apply(event, callbackEvent)
		}
		if err != nil && !errors.Is(err, pluginruntime.ErrPluginDisabled) {
			hub.log.Error("plugin event listener failed", zap.String("plugin", entry.scope.Name()), zap.String("event", event.Name()), zap.Error(err))
		}
	}
	if cancellable, ok := event.(sdkevent.Cancellable); ok && cancellable.Cancelled() {
		return ErrEventCancelled
	}
	return nil
}

// HasListeners reports whether at least one enabled listener observes a name.
func (hub *Hub) HasListeners(name string) bool {
	hub.mutex.RLock()
	defer hub.mutex.RUnlock()
	for _, entry := range hub.listeners[strings.TrimSpace(name)] {
		if entry.scope.Enabled() {
			return true
		}
	}
	return false
}

// ignoreCancelled reports whether listener options skip a vetoed event.
func ignoreCancelled(event sdkevent.Event, options sdkevent.ListenerOptions) bool {
	cancellable, ok := event.(sdkevent.Cancellable)
	return ok && options.IgnoreCancelled && cancellable.Cancelled()
}

// clone creates a callback-owned event snapshot safe from late mutations.
func clone(event sdkevent.Event) sdkevent.Event {
	if mutable, ok := event.(sdkevent.Mutable); ok {
		return mutable.Clone()
	}
	if cloner, ok := event.(eventCloner); ok {
		return cloner.CloneEvent()
	}
	return event
}

// apply commits a successful listener's mutable event result.
func apply(target sdkevent.Event, result sdkevent.Event) {
	targetMutable, targetOK := target.(sdkevent.Mutable)
	resultMutable, resultOK := result.(sdkevent.Mutable)
	if !targetOK || !resultOK {
		return
	}
	resultMutable.Apply(targetMutable)
}
