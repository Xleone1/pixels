// Package trade implements bounded live-trade capabilities for plugins.
package trade

import (
	"errors"
	"strings"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
)

var (
	// ErrUnavailable reports a trade capability without its realm dependency.
	ErrUnavailable = errors.New("plugin trade access unavailable")
	// ErrInvalidReason reports an empty force-cancellation reason.
	ErrInvalidReason = errors.New("plugin trade cancellation reason is required")
)

// Access exposes scoped live-trade behavior.
type Access struct {
	// service owns trade lifecycle behavior.
	service *tradecore.Service
	// scope identifies the calling plugin.
	scope *pluginruntime.Scope
}

// NewAccess creates one plugin-scoped trade facade.
func NewAccess(service *tradecore.Service, scope *pluginruntime.Scope) *Access {
	return &Access{service: service, scope: scope}
}

// Active returns one player's active trade.
func (access *Access) Active(playerID int64) (sdkplugin.TradeSnapshot, bool) {
	if access.service == nil || access.scope == nil || !access.scope.Enabled() {
		return sdkplugin.TradeSnapshot{}, false
	}
	session, found := access.service.Registry().Find(playerID)
	if !found {
		return sdkplugin.TradeSnapshot{}, false
	}
	return snapshot(session), true
}

// ForceCancel closes one active trade with a plugin-scoped reason.
func (access *Access) ForceCancel(playerID int64, reason string) error {
	if access.service == nil || access.scope == nil || !access.scope.Enabled() {
		return ErrUnavailable
	}
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ErrInvalidReason
	}
	return access.service.ForceCancel(playerID, "plugin:"+access.scope.Name()+":"+reason)
}

// snapshot copies one live trade into its SDK representation.
func snapshot(session *traderuntime.Session) sdkplugin.TradeSnapshot {
	first, second := session.Snapshot()
	return sdkplugin.TradeSnapshot{
		RoomID: session.RoomID,
		First:  participant(first),
		Second: participant(second),
	}
}

// participant copies one live trade participant.
func participant(value traderuntime.Participant) sdkplugin.TradeParticipant {
	return sdkplugin.TradeParticipant{
		PlayerID: value.PlayerID, UnitID: value.UnitID, Username: value.Username,
		ItemIDs: append([]int64(nil), value.Items...), Accepted: value.Accepted, Confirmed: value.Confirmed,
	}
}
