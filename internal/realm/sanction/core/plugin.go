package core

import (
	"context"
	"time"
)

// EventDispatcher intercepts global sanctions before persistence.
type EventDispatcher interface {
	// DispatchSanctionApply returns mutable sanction fields and cancellation.
	DispatchSanctionApply(context.Context, *int64, int64, string, string, *time.Time) (string, *time.Time, bool)
}

// SetPluginRuntime installs the optional dynamic-plugin sanction interceptor.
func (service *Service) SetPluginRuntime(events EventDispatcher) {
	service.pluginEvents = events
}
