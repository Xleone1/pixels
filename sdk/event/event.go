// Package event contains typed events exposed to dynamic plugins.
package event

import (
	"context"

	sdkpriority "github.com/niflaot/pixels/sdk/priority"
)

// Event identifies one plugin-facing event.
type Event interface {
	// Name returns the stable event identifier.
	Name() string
}

// Cancellable describes an event whose default action may be vetoed.
type Cancellable interface {
	Event
	// Cancelled reports whether a listener vetoed the default action.
	Cancelled() bool
	// SetCancelled changes whether the default action is vetoed.
	SetCancelled(bool)
}

// Mutable describes an event whose callback-owned state can be committed safely.
type Mutable interface {
	Event
	// Clone returns an isolated copy for one plugin callback.
	Clone() Mutable
	// Apply copies this callback result onto the original event.
	Apply(Mutable)
}

// Listener reacts to one dispatched plugin-facing event.
type Listener func(context.Context, Event) error

// Published is an immutable post-commit realm notification.
type Published struct {
	// name stores the stable realm event identifier.
	name string
	// fields stores a normalized payload snapshot without internal realm types.
	fields map[string]any
}

// NewPublished creates one immutable post-commit event snapshot.
func NewPublished(name string, fields map[string]any) *Published {
	return &Published{name: name, fields: cloneFields(fields)}
}

// Name returns the stable realm event identifier.
func (event *Published) Name() string { return event.name }

// Fields returns a caller-owned copy of every normalized payload field.
func (event *Published) Fields() map[string]any { return cloneFields(event.fields) }

// Field returns one normalized payload field.
func (event *Published) Field(name string) (any, bool) {
	value, found := event.fields[name]
	return cloneValue(value), found
}

// CloneEvent returns one independent immutable notification.
func (event *Published) CloneEvent() Event {
	return NewPublished(event.name, event.fields)
}

// ListenerOptions configures event listener ordering and cancellation behavior.
type ListenerOptions struct {
	// Priority runs larger values before smaller values.
	Priority sdkpriority.Priority
	// IgnoreCancelled skips this listener after an earlier listener cancels.
	IgnoreCancelled bool
}

// cloneFields copies one normalized field map.
func cloneFields(fields map[string]any) map[string]any {
	cloned := make(map[string]any, len(fields))
	for name, value := range fields {
		cloned[name] = cloneValue(value)
	}
	return cloned
}

// cloneValue copies normalized composite values.
func cloneValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return cloneFields(typed)
	case []any:
		cloned := make([]any, len(typed))
		for index, item := range typed {
			cloned[index] = cloneValue(item)
		}
		return cloned
	default:
		return typed
	}
}
