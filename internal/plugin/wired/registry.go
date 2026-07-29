// Package wired registers isolated plugin-owned WIRED behavior.
package wired

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/condition"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	sdkwired "github.com/niflaot/pixels/sdk/wired"
	"go.uber.org/zap"
)

var (
	// ErrInvalidRegistration reports missing plugin WIRED behavior data.
	ErrInvalidRegistration = errors.New("invalid plugin WIRED registration")
	// ErrDuplicateKey reports a plugin WIRED key claimed more than once.
	ErrDuplicateKey = errors.New("duplicate plugin WIRED key")
)

// behavior stores one plugin-owned descriptor and callback.
type behavior struct {
	// descriptor stores compiler-visible behavior metadata.
	descriptor registry.Descriptor
	// scope stores plugin identity and enabled state.
	scope *pluginruntime.Scope
	// effect executes effect behavior when present.
	effect sdkwired.EffectExecutor
	// condition evaluates condition behavior when present.
	condition sdkwired.ConditionEvaluator
}

// Registry stores plugin-owned WIRED behavior independently of native descriptors.
type Registry struct {
	// mutex protects startup registration and runtime lookup.
	mutex sync.RWMutex
	// behaviors stores namespaced registrations.
	behaviors map[string]behavior
	// timeout bounds plugin callbacks.
	timeout time.Duration
	// log records isolated callback failures.
	log *zap.Logger
}

// NewRegistry creates an empty plugin WIRED registry.
func NewRegistry(timeout time.Duration, log *zap.Logger) *Registry {
	if log == nil {
		log = zap.NewNop()
	}
	return &Registry{behaviors: make(map[string]behavior), timeout: timeout, log: log}
}

// register stores one validated namespaced behavior.
func (registered *Registry) register(scope *pluginruntime.Scope, key string, descriptor sdkwired.Descriptor, executor sdkwired.EffectExecutor, evaluator sdkwired.ConditionEvaluator) error {
	key = strings.TrimSpace(key)
	if scope == nil || key == "" || executor == nil && evaluator == nil {
		return ErrInvalidRegistration
	}
	if !strings.HasPrefix(key, "plugin."+scope.Name()+".") || len(key) == len("plugin."+scope.Name()+".") {
		return pluginruntime.ErrWrongNamespace
	}
	if descriptor.Selection > sdkwired.SelectionRequired || descriptor.Actor > sdkwired.ActorBot {
		return ErrInvalidRegistration
	}
	family := registry.FamilyEffect
	if evaluator != nil {
		family = registry.FamilyCondition
	}
	entry := behavior{
		descriptor: registry.Descriptor{
			Key: key, Family: family, ClientCode: descriptor.ClientCode,
			Selection: registry.SelectionPolicy(descriptor.Selection),
			Actor:     registry.ActorPolicy(descriptor.Actor), Editor: descriptor.Editor,
		},
		scope: scope, effect: executor, condition: evaluator,
	}
	registered.mutex.Lock()
	defer registered.mutex.Unlock()
	if _, found := registered.behaviors[key]; found {
		return fmt.Errorf("%w: %s", ErrDuplicateKey, key)
	}
	registered.behaviors[key] = entry
	return nil
}

// Resolve returns compiler metadata for one plugin key.
func (registered *Registry) Resolve(key string) (registry.Descriptor, bool) {
	registered.mutex.RLock()
	entry, found := registered.behaviors[key]
	registered.mutex.RUnlock()
	if !found || !entry.scope.Enabled() {
		return registry.Descriptor{}, false
	}
	return entry.descriptor, true
}

// ExecuteEffect executes a registered plugin effect when the key matches.
func (registered *Registry) ExecuteEffect(ctx context.Context, node *configuration.Node, event trigger.Event) (effect.Result, bool, error) {
	entry, found := registered.find(node, true)
	if !found {
		return effect.Result{}, false, nil
	}
	var outcome sdkwired.EffectResult
	err := pluginruntime.InvokeCallback(ctx, registered.timeout, entry.scope, "wired.effect", registered.log, func(callbackContext context.Context) error {
		var callbackErr error
		outcome, callbackErr = entry.effect(callbackContext, sdkNode(node), sdkEvent(event))
		return callbackErr
	})
	return effectResult(outcome), true, err
}

// EvaluateCondition evaluates a registered plugin condition when the key matches.
func (registered *Registry) EvaluateCondition(node *configuration.Node, input condition.Context) (condition.Result, bool, error) {
	entry, found := registered.find(node, false)
	if !found {
		return condition.Result{}, false, nil
	}
	ctx := input.CallbackContext
	if ctx == nil {
		ctx = context.Background()
	}
	var pass, valid bool
	err := pluginruntime.InvokeCallback(ctx, registered.timeout, entry.scope, "wired.condition", registered.log, func(callbackContext context.Context) error {
		var callbackErr error
		pass, valid, callbackErr = entry.condition(callbackContext, sdkNode(node), sdkEvent(input.Event))
		return callbackErr
	})
	return condition.Result{Pass: pass, Valid: valid}, true, err
}

// find returns one enabled behavior of the requested family.
func (registered *Registry) find(node *configuration.Node, effectFamily bool) (behavior, bool) {
	if node == nil {
		return behavior{}, false
	}
	registered.mutex.RLock()
	entry, found := registered.behaviors[node.Descriptor.Key]
	registered.mutex.RUnlock()
	if !found || !entry.scope.Enabled() || effectFamily && entry.effect == nil || !effectFamily && entry.condition == nil {
		return behavior{}, false
	}
	return entry, true
}

// sdkNode creates an immutable detached SDK node.
func sdkNode(node *configuration.Node) sdkwired.Node {
	targets := make([]int64, len(node.Targets))
	for index, target := range node.Targets {
		targets[index] = target.ItemID
	}
	return sdkwired.Node{
		Key: node.Descriptor.Key, RoomID: node.RoomID, ItemID: node.ItemID,
		Targets: targets, Text: node.Parameters.Text,
		Values: append([]int32(nil), node.Parameters.Values...), Duration: node.Parameters.Duration,
	}
}

// sdkEvent creates an immutable detached SDK event.
func sdkEvent(event trigger.Event) sdkwired.Event {
	return sdkwired.Event{RoomID: event.RoomID, PlayerID: event.PlayerID, SourceItemID: event.SourceItem}
}

// effectResult converts one SDK result to the native execution result.
func effectResult(result sdkwired.EffectResult) effect.Result {
	status := effect.Skipped
	if result.Status == sdkwired.EffectApplied {
		status = effect.Applied
	} else if result.Status == sdkwired.EffectBlocked {
		status = effect.Blocked
	}
	return effect.Result{Status: status, ResetTimers: result.ResetTimers, CallTargets: append([]int64(nil), result.CallTargets...)}
}
