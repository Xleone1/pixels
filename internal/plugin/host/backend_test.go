package host

import (
	"testing"
	"time"

	plugincommand "github.com/niflaot/pixels/internal/plugin/command"
	pluginevent "github.com/niflaot/pixels/internal/plugin/event"
	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	pluginwired "github.com/niflaot/pixels/internal/plugin/wired"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
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
	"go.uber.org/zap"
)

// TestBackendBuildsEveryScopedCapability verifies host composition boundaries.
func TestBackendBuildsEveryScopedCapability(t *testing.T) {
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	connections := netconn.NewRegistry()
	events := pluginevent.NewHub(time.Second, zap.NewNop())
	commands := plugincommand.NewTree(":", time.Second, nil, zap.NewNop())
	backend := NewBackend(
		players, bindings, connections, netconn.NewHandlerRegistry(), nil,
		pluginroutes.New(), pluginevent.NewHub(time.Second, zap.NewNop()), plugincommand.NewTree(":", time.Second, nil, zap.NewNop()), pluginwired.NewRegistry(time.Second, nil), time.Second, zap.NewNop(),
	)
	currencies := &currencyservice.Service{}
	rooms := roomservice.New(nil, nil)
	liveRooms := roomlive.NewRegistry(nil)
	trades := &tradecore.Service{}
	if backend.WithEconomy(currencies) != backend || backend.WithRooms(rooms, liveRooms) != backend || backend.WithTrades(trades) != backend {
		t.Fatal("capability installers did not preserve the backend")
	}
	host := backend.HostFor(pluginruntime.NewScope("demo"))
	if host.Players() == nil || host.Routes() == nil || host.Events() == nil || host.Commands() == nil || host.Permissions() == nil ||
		host.Economy() == nil || host.Rooms() == nil || host.Trades() == nil || host.Wired() == nil {
		t.Fatal("expected every scoped capability")
	}
	backend.events = events
	backend.commands = commands
	local := bus.New()
	chat := chatsend.New(chatconfig.Config{}, players, bindings, roomlive.NewRegistry(nil), connections, nil, nil, nil, nil, local, nil, chatsend.Nodes{})
	if err := backend.Connect(local, chat, currencies); err != nil {
		t.Fatalf("connect plugin bridges: %v", err)
	}
}

// TestBackendConnectsEveryRealmSeam verifies complete runtime composition.
func TestBackendConnectsEveryRealmSeam(t *testing.T) {
	players := playerlive.NewRegistry()
	bindings := binding.NewRegistry()
	connections := netconn.NewRegistry()
	local := bus.New()
	events := pluginevent.NewHub(time.Second, nil)
	commands := plugincommand.NewTree(":", time.Second, nil, nil)
	backend := NewBackend(
		players, bindings, connections, netconn.NewHandlerRegistry(), nil,
		pluginroutes.New(), events, commands, pluginwired.NewRegistry(time.Second, nil), time.Second, nil,
	)
	chat := chatsend.New(chatconfig.Config{}, players, bindings, roomlive.NewRegistry(nil), connections, nil, nil, nil, nil, local, nil, chatsend.Nodes{})
	if err := backend.ConnectRealms(
		local,
		chat,
		&currencyservice.Service{},
		roomservice.New(nil, nil),
		&roommoderation.Service{},
		&sanctioncore.Service{},
		&tradecore.Service{},
		&furnitureservice.Service{},
		&catalogservice.Service{},
	); err != nil {
		t.Fatalf("connect realms: %v", err)
	}
	if backend.realmEvents != local {
		t.Fatal("publisher was not retained for scoped room actions")
	}
}
