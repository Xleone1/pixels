// Package host composes capability-specific dynamic-plugin facades.
package host

import (
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	plugineconomy "github.com/niflaot/pixels/internal/plugin/economy"
	pluginevent "github.com/niflaot/pixels/internal/plugin/event"
	plugineventbridge "github.com/niflaot/pixels/internal/plugin/event/bridge"
	pluginpermission "github.com/niflaot/pixels/internal/plugin/permission"
	pluginplayer "github.com/niflaot/pixels/internal/plugin/player"
	pluginroom "github.com/niflaot/pixels/internal/plugin/room"
	pluginroute "github.com/niflaot/pixels/internal/plugin/route"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	plugintrade "github.com/niflaot/pixels/internal/plugin/trade"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/http/pluginroutes"
	sdkplugin "github.com/niflaot/pixels/sdk/plugin"
	"go.uber.org/zap"
)

// Backend owns dependencies used to construct scoped plugin hosts.
type Backend struct {
	// players stores connected player state.
	players *playerlive.Registry
	// bindings resolves authenticated connections.
	bindings *binding.Registry
	// connections sends packets and disconnects sessions.
	connections *netconn.Registry
	// inbound stores the shared inbound packet pipeline.
	inbound *netconn.HandlerRegistry
	// permissions resolves real permission nodes.
	permissions permissionservice.Checker
	// routes stores isolated route declarations.
	routes *pluginroutes.Registry
	// events dispatches plugin-facing typed events.
	events *pluginevent.Hub
	// commands dispatches plugin chat commands.
	commands *plugincommand.Tree
	// economy owns durable plugin currency operations.
	economy *currencyservice.Service
	// rooms owns durable room records.
	rooms *roomservice.Service
	// liveRooms owns active room state.
	liveRooms *roomlive.Registry
	// trades owns live direct-trade behavior.
	trades *tradecore.Service
	// realmEvents publishes plugin-triggered realm facts.
	realmEvents bus.Publisher
	// timeout bounds plugin callbacks.
	timeout time.Duration
	// log records isolated plugin failures.
	log *zap.Logger
}

// WithEconomy installs durable currency capabilities.
func (backend *Backend) WithEconomy(service *currencyservice.Service) *Backend {
	backend.economy = service
	return backend
}

// WithRooms installs durable and active room capabilities.
func (backend *Backend) WithRooms(service *roomservice.Service, live *roomlive.Registry) *Backend {
	backend.rooms = service
	backend.liveRooms = live
	return backend
}

// WithTrades installs live direct-trade capabilities.
func (backend *Backend) WithTrades(service *tradecore.Service) *Backend {
	backend.trades = service
	return backend
}

// NewBackend creates the shared dynamic plugin host backend.
func NewBackend(players *playerlive.Registry, bindings *binding.Registry, connections *netconn.Registry, inbound *netconn.HandlerRegistry, permissions permissionservice.Checker, routes *pluginroutes.Registry, events *pluginevent.Hub, commands *plugincommand.Tree, timeout time.Duration, log *zap.Logger) *Backend {
	if log == nil {
		log = zap.NewNop()
	}
	return &Backend{players: players, bindings: bindings, connections: connections, inbound: inbound, permissions: permissions, routes: routes, events: events, commands: commands, timeout: timeout, log: log}
}

// HostFor creates a namespace-enforcing facade for one plugin scope.
func (backend *Backend) HostFor(scope *pluginruntime.Scope) sdkplugin.Host {
	players := pluginplayer.NewAccess(backend.players, backend.bindings, backend.connections, backend.inbound, backend.permissions, backend.timeout, backend.log, scope)
	return &scopedHost{
		players:     players,
		routes:      pluginroute.NewAccess(backend.routes, scope),
		events:      pluginevent.NewAccess(backend.events, scope),
		commands:    plugincommand.NewAccess(backend.commands, scope),
		permissions: pluginpermission.NewAccess(scope),
		economy:     plugineconomy.NewAccess(backend.economy, scope),
		rooms:       pluginroom.NewAccess(backend.rooms, backend.liveRooms, backend.connections, backend.realmEvents, scope),
		trades:      plugintrade.NewAccess(backend.trades, scope),
	}
}

// Connect installs plugin commands, events, and player lifecycle forwarding.
func (backend *Backend) Connect(subscriber bus.Subscriber, chat *chatsend.Service, currencies ...*currencyservice.Service) error {
	var currencyRealm *currencyservice.Service
	if len(currencies) > 0 {
		currencyRealm = currencies[0]
	}
	return backend.connectCore(subscriber, chat, currencyRealm, nil)
}

// ConnectRealms installs every realm bridge used by the complete plugin runtime.
func (backend *Backend) ConnectRealms(subscriber bus.Subscriber, chat *chatsend.Service, currencies *currencyservice.Service, rooms *roomservice.Service, moderation *roommoderation.Service, sanctions *sanctioncore.Service, trades *tradecore.Service, furniture *furnitureservice.Service, catalog *catalogservice.Service) error {
	if err := backend.connectCore(subscriber, chat, currencies, rooms); err != nil {
		return err
	}
	moderation.SetPluginRuntime(backend.events)
	sanctions.SetPluginRuntime(backend.events)
	trades.SetPluginRuntime(backend.events)
	furniture.SetPluginRuntime(backend.events)
	catalog.SetPluginRuntime(backend.events)
	return nil
}

// connectCore installs shared runtime seams and post-commit event forwarding.
func (backend *Backend) connectCore(subscriber bus.Subscriber, chat *chatsend.Service, currencies *currencyservice.Service, rooms *roomservice.Service) error {
	if publisher, ok := subscriber.(bus.Publisher); ok {
		backend.realmEvents = publisher
	}
	access := pluginplayer.NewAccess(backend.players, backend.bindings, backend.connections, backend.inbound, backend.permissions, backend.timeout, backend.log, pluginruntime.NewScope("pixels"))
	backend.commands.SetPlayers(access)
	backend.commands.SetAttemptDispatcher(backend.events)
	backend.events.SetPlayerFinder(access)
	chat.SetPluginRuntime(backend.commands, backend.events)
	if err := backend.events.RegisterPlayerConnected(subscriber, access); err != nil {
		return err
	}
	if err := plugineventbridge.Register(subscriber, backend.events); err != nil {
		return err
	}
	if currencies != nil {
		currencies.SetPluginRuntime(backend.events)
		if err := backend.events.RegisterCurrencyChanged(subscriber, access); err != nil {
			return err
		}
	}
	if rooms != nil {
		rooms.SetPluginRuntime(backend.events)
	}
	return nil
}

// scopedHost implements the single plugin-facing capability facade.
type scopedHost struct {
	// players exposes scoped player operations.
	players sdkplugin.PlayerAccess
	// routes exposes scoped route registration.
	routes sdkplugin.RouteRegistrar
	// events exposes scoped listener registration.
	events sdkplugin.EventHub
	// commands exposes scoped command registration.
	commands sdkplugin.CommandTree
	// permissions exposes scoped node registration.
	permissions sdkplugin.PermissionRegistrar
	// economy exposes scoped currency behavior.
	economy sdkplugin.EconomyAccess
	// rooms exposes scoped room behavior.
	rooms sdkplugin.RoomAccess
	// trades exposes scoped live-trade behavior.
	trades sdkplugin.TradeAccess
}

// Players returns bounded live-player access.
func (host *scopedHost) Players() sdkplugin.PlayerAccess { return host.players }

// Routes returns isolated HTTP route registration.
func (host *scopedHost) Routes() sdkplugin.RouteRegistrar { return host.routes }

// Events returns plugin-facing event registration.
func (host *scopedHost) Events() sdkplugin.EventHub { return host.events }

// Commands returns the shared Brigadier command tree.
func (host *scopedHost) Commands() sdkplugin.CommandTree { return host.commands }

// Permissions returns namespaced permission registration.
func (host *scopedHost) Permissions() sdkplugin.PermissionRegistrar { return host.permissions }

// Economy returns bounded plugin currency behavior.
func (host *scopedHost) Economy() sdkplugin.EconomyAccess { return host.economy }

// Rooms returns bounded plugin room behavior.
func (host *scopedHost) Rooms() sdkplugin.RoomAccess { return host.rooms }

// Trades returns bounded plugin live-trade behavior.
func (host *scopedHost) Trades() sdkplugin.TradeAccess { return host.trades }
