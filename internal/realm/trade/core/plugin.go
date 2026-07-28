package core

import (
	"context"

	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
)

// EventDispatcher intercepts trade lifecycle mutations.
type EventDispatcher interface {
	// DispatchTradeStart reports whether opening a trade was vetoed.
	DispatchTradeStart(context.Context, *traderuntime.Session) bool
	// DispatchTradeConfirm reports whether settling a trade was vetoed.
	DispatchTradeConfirm(context.Context, *traderuntime.Session) bool
	// DispatchTradeCancel reports whether closing a trade was vetoed.
	DispatchTradeCancel(context.Context, int64, *traderuntime.Session, string) bool
}

// SetPluginRuntime installs the optional dynamic-plugin trade interceptor.
func (service *Service) SetPluginRuntime(events EventDispatcher) {
	service.pluginEvents = events
}

// ForceCancel closes one active trade through the plugin action boundary.
func (service *Service) ForceCancel(playerID int64, reason string) error {
	if !service.closeWithReason(context.Background(), playerID, 0, reason) {
		if _, found := service.registry.Find(playerID); found {
			return ErrCancelledByPlugin
		}
		return ErrUnavailable
	}
	return nil
}
