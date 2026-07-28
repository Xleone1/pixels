// Package economy implements bounded plugin currency capabilities.
package economy

import (
	"context"
	"errors"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
)

var (
	// ErrUnavailable reports missing economy composition in the host.
	ErrUnavailable = errors.New("plugin economy unavailable")
)

// Access exposes scoped currency operations backed by the real realm service.
type Access struct {
	// service owns durable currency behavior.
	service *currencyservice.Service
	// scope identifies the calling plugin for audit reasons.
	scope *pluginruntime.Scope
}

// NewAccess creates one plugin-scoped economy facade.
func NewAccess(service *currencyservice.Service, scope *pluginruntime.Scope) *Access {
	return &Access{service: service, scope: scope}
}

// Grant applies a signed delta using the plugin actor family.
func (access *Access) Grant(playerID int64, currencyType int32, amount int64) (int64, error) {
	if !access.available() {
		return 0, ErrUnavailable
	}
	return access.service.Grant(context.Background(), currencyservice.GrantParams{
		PlayerID: playerID, CurrencyType: currencyType, Amount: amount,
		Reason: access.reason("grant"), ActorKind: currencyservice.ActorPlugin,
	})
}

// Set replaces a balance using the plugin actor family.
func (access *Access) Set(playerID int64, currencyType int32, amount int64) (int64, error) {
	if !access.available() {
		return 0, ErrUnavailable
	}
	return access.service.Set(context.Background(), currencyservice.SetParams{
		PlayerID: playerID, CurrencyType: currencyType, Amount: amount,
		Reason: access.reason("set"), ActorKind: currencyservice.ActorPlugin,
	})
}

// Balance reads one configured player balance.
func (access *Access) Balance(playerID int64, currencyType int32) (int64, error) {
	if !access.available() {
		return 0, ErrUnavailable
	}
	return access.service.Balance(context.Background(), playerID, currencyType)
}

// Types lists configured currencies without exposing internal records.
func (access *Access) Types() ([]sdkplugin.CurrencyDefinition, error) {
	if !access.available() {
		return nil, ErrUnavailable
	}
	definitions, err := access.service.Types(context.Background())
	if err != nil {
		return nil, err
	}
	result := make([]sdkplugin.CurrencyDefinition, 0, len(definitions))
	for _, definition := range definitions {
		result = append(result, sdkplugin.CurrencyDefinition{
			Type: definition.Type, Key: definition.Key, Ledger: definition.Ledger, Color: definition.Color,
		})
	}
	return result, nil
}

// available reports whether the scoped facade may reach its realm.
func (access *Access) available() bool {
	return access.service != nil && access.scope != nil && access.scope.Enabled()
}

// reason creates one durable plugin-owned audit reason.
func (access *Access) reason(action string) string {
	return "plugin:" + access.scope.Name() + ":" + action
}
