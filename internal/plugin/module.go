// Package plugin wires native dynamic plugins into controlled host capabilities.
package plugin

import (
	"context"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	pluginconfig "github.com/niflaot/pixels/internal/plugin/config"
	pluginevent "github.com/niflaot/pixels/internal/plugin/event"
	pluginhost "github.com/niflaot/pixels/internal/plugin/host"
	"github.com/niflaot/pixels/internal/plugin/loader"
	pluginwired "github.com/niflaot/pixels/internal/plugin/wired"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	realmconn "github.com/niflaot/pixels/internal/realm/connection"
	craftingrecipe "github.com/niflaot/pixels/internal/realm/crafting/recipe"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	groupmembership "github.com/niflaot/pixels/internal/realm/group/membership"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	marketplacecore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	messengercore "github.com/niflaot/pixels/internal/realm/messenger/core"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playerprofile "github.com/niflaot/pixels/internal/realm/player/profile"
	entercmd "github.com/niflaot/pixels/internal/realm/room/access/commands/enter"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomwalk "github.com/niflaot/pixels/internal/realm/room/world/commands/walk"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/http/pluginroutes"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides dynamic plugin discovery and controlled SDK capabilities.
var Module = fx.Module(
	"plugins",
	fx.Provide(
		newEventHub,
		newCommandTree,
		newWiredRegistry,
		newBackend,
		newRoomMoveDispatcher,
		newRoomEntryDispatcher,
		loader.NewNative,
		loader.New,
	),
	fx.Invoke(connectRuntime),
	fx.Invoke(loadPlugins),
)

// newEventHub creates the shared plugin event dispatcher.
func newEventHub(config pluginconfig.Config, log *zap.Logger) *pluginevent.Hub {
	return pluginevent.NewHub(config.CallbackTimeout, log)
}

// newRoomMoveDispatcher exposes the hub through the room movement seam.
func newRoomMoveDispatcher(events *pluginevent.Hub) roomwalk.EventDispatcher { return events }

// newRoomEntryDispatcher exposes the hub through the room admission seam.
func newRoomEntryDispatcher(events *pluginevent.Hub) entercmd.EventDispatcher { return events }

// newCommandTree creates the shared plugin command dispatcher.
func newCommandTree(config pluginconfig.Config, translations i18n.Translator, log *zap.Logger) *plugincommand.Tree {
	return plugincommand.NewTree(config.CommandPrefix, config.CallbackTimeout, translations, log)
}

// newWiredRegistry creates the separate plugin-owned WIRED registry.
func newWiredRegistry(config pluginconfig.Config, log *zap.Logger) *pluginwired.Registry {
	return pluginwired.NewRegistry(config.CallbackTimeout, log)
}

// newBackend composes internal registries behind the controlled plugin facade.
func newBackend(config pluginconfig.Config, players *playerlive.Registry, bindings *binding.Registry, connections *netconn.Registry, handlers *realmconn.Handlers, permissions permissionservice.Checker, routes *pluginroutes.Registry, events *pluginevent.Hub, commands *plugincommand.Tree, wired *pluginwired.Registry, currencies *currencyservice.Service, rooms *roomservice.Service, liveRooms *roomlive.Registry, trades *tradecore.Service, log *zap.Logger) *pluginhost.Backend {
	return pluginhost.NewBackend(players, bindings, connections, handlers.Inbound, permissions, routes, events, commands, wired, config.CallbackTimeout, log).
		WithEconomy(currencies).
		WithRooms(rooms, liveRooms).
		WithTrades(trades)
}

// connectRuntime installs command, chat-event, and lifecycle bridges before loading plugins.
func connectRuntime(backend *pluginhost.Backend, subscriber bus.Subscriber, chat *chatsend.Service, currencies *currencyservice.Service, rooms *roomservice.Service, moderation *roommoderation.Service, sanctions *sanctioncore.Service, trades *tradecore.Service, furniture *furnitureservice.Service, catalog *catalogservice.Service, marketplace *marketplacecore.Service, profile *playerprofile.Service, bots *botcore.Service, groups *groupmembership.Service, messenger *messengercore.Service, crafting *craftingrecipe.Service) error {
	if err := backend.ConnectRealms(subscriber, chat, currencies, rooms, moderation, sanctions, trades, furniture, catalog); err != nil {
		return err
	}
	events := backend.Events()
	marketplace.SetPluginRuntime(events)
	profile.SetPluginRuntime(events)
	bots.SetPluginRuntime(events)
	groups.SetPluginRuntime(events)
	messenger.SetPluginRuntime(events)
	crafting.SetPluginRuntime(events)
	return nil
}

// loadPlugins completes registration before the HTTP application is assembled.
func loadPlugins(plugins *loader.Loader) error {
	return plugins.Load(context.Background())
}
