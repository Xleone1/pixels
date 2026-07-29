// Package wired defines controlled plugin-owned WIRED behavior contracts.
package wired

import (
	"context"
	"time"
)

// SelectionPolicy describes whether a plugin behavior accepts furniture targets.
type SelectionPolicy uint8

const (
	// SelectionNone rejects selected furniture.
	SelectionNone SelectionPolicy = iota
	// SelectionOptional accepts zero or more selected furniture items.
	SelectionOptional
	// SelectionRequired requires selected furniture.
	SelectionRequired
)

// ActorPolicy describes which event actors a plugin behavior accepts.
type ActorPolicy uint8

const (
	// ActorOptional accepts system and room-unit events.
	ActorOptional ActorPolicy = iota
	// ActorPlayer requires a player actor.
	ActorPlayer
	// ActorUnit requires any room-unit actor.
	ActorUnit
	// ActorBot requires a bot actor.
	ActorBot
)

// Descriptor declares reusable stock-Nitro editor and execution policies.
type Descriptor struct {
	// ClientCode stores Nitro's existing editor layout discriminator.
	ClientCode int32
	// Selection stores the target-selection policy.
	Selection SelectionPolicy
	// Actor stores the event-actor policy.
	Actor ActorPolicy
	// Editor reports whether stock Nitro exposes a compatible editor.
	Editor bool
}

// Option changes one plugin behavior descriptor.
type Option func(*Descriptor)

// WithDescriptor replaces the default plugin behavior descriptor.
func WithDescriptor(descriptor Descriptor) Option {
	return func(target *Descriptor) { *target = descriptor }
}

// ResolveDescriptor applies registration options to safe defaults.
func ResolveDescriptor(options []Option) Descriptor {
	descriptor := Descriptor{Selection: SelectionOptional, Actor: ActorOptional, Editor: true}
	for _, option := range options {
		if option != nil {
			option(&descriptor)
		}
	}
	return descriptor
}

// Node is an immutable view of one configured WIRED furniture item.
type Node struct {
	// Key stores the namespaced interaction key.
	Key string
	// RoomID identifies the containing room.
	RoomID int64
	// ItemID identifies the configured furniture item.
	ItemID int64
	// Targets stores selected furniture item identifiers.
	Targets []int64
	// Text stores the configured free-text parameter.
	Text string
	// Values stores configured numeric parameters.
	Values []int32
	// Duration stores the compiled duration parameter.
	Duration time.Duration
}

// Event is an immutable room-event snapshot supplied to plugin WIRED logic.
type Event struct {
	// RoomID identifies the room.
	RoomID int64
	// PlayerID identifies a player actor or zero.
	PlayerID int64
	// SourceItemID identifies source furniture or zero.
	SourceItemID int64
}

// EffectStatus classifies a plugin effect outcome.
type EffectStatus int

const (
	// EffectApplied reports successful behavior.
	EffectApplied EffectStatus = iota
	// EffectBlocked reports a safety or authorization rejection.
	EffectBlocked
	// EffectSkipped reports an absent target or no-op.
	EffectSkipped
)

// EffectResult stores one plugin effect outcome.
type EffectResult struct {
	// Status classifies the outcome.
	Status EffectStatus
	// ResetTimers requests a room timer reset.
	ResetTimers bool
	// CallTargets stores WIRED stacks to enqueue.
	CallTargets []int64
}

// EffectExecutor executes one plugin-owned WIRED effect.
type EffectExecutor func(context.Context, Node, Event) (EffectResult, error)

// ConditionEvaluator evaluates one plugin-owned WIRED condition.
type ConditionEvaluator func(context.Context, Node, Event) (bool, bool, error)
