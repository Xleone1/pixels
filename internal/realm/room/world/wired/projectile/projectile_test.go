package projectile

import (
	"context"
	"errors"
	"testing"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/selection"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// fakeMover records final projectile persistence.
type fakeMover struct {
	// placements stores committed final placements.
	placements []furnituremodel.Placement
	// err injects one persistence failure.
	err error
}

// Move records one final movement or returns the injected error.
func (mover *fakeMover) Move(_ context.Context, params furnitureservice.MoveParams) (furnituremodel.Item, error) {
	if mover.err != nil {
		return furnituremodel.Item{}, mover.err
	}
	mover.placements = append(mover.placements, params.Placement)
	return furnituremodel.Item{}, nil
}

// TestProjectileMovesOncePerPulseAndPersistsOnlyFinalPosition verifies bounded room-cycle movement.
func TestProjectileMovesOncePerPulseAndPersistsOnlyFinalPosition(t *testing.T) {
	service, active, mover := projectileRoom(t, nil)
	launchProjectile(t, service, 0, []int32{2, 2, 1})
	if err := service.Cycle(context.Background(), active.ID(), time.Now()); err != nil {
		t.Fatal(err)
	}
	if len(mover.placements) != 0 {
		t.Fatal("intermediate projectile step was persisted")
	}
	if err := service.Cycle(context.Background(), active.ID(), time.Now()); err != nil {
		t.Fatal(err)
	}
	item, found := active.FurnitureItem(10)
	if !found || item.Point != grid.MustPoint(3, 1) {
		t.Fatalf("projectile item=%#v found=%t", item, found)
	}
	if len(mover.placements) != 1 || mover.placements[0].X != 3 || mover.placements[0].Y != 1 {
		t.Fatalf("final placements=%#v", mover.placements)
	}
	assertVariable(t, service, active.ID(), systemTilesTravelled, 2)
	assertVariable(t, service, active.ID(), systemMoving, 0)
}

// TestProjectileSweptCollisionsStopBeforeFurnitureAndUsers verifies both collision families.
func TestProjectileSweptCollisionsStopBeforeFurnitureAndUsers(t *testing.T) {
	t.Run("furniture", func(t *testing.T) {
		blocker := worldfurniture.Item{ID: 20, Point: grid.MustPoint(2, 1), Definition: worldfurniture.Definition{Width: 1, Length: 1}}
		service, active, mover := projectileRoom(t, []worldfurniture.Item{blocker})
		launchProjectile(t, service, 0, []int32{2, 2, 1})
		if err := service.Cycle(context.Background(), active.ID(), time.Now()); err != nil {
			t.Fatal(err)
		}
		if len(mover.placements) != 0 {
			t.Fatal("stationary collision was persisted")
		}
		assertVariable(t, service, active.ID(), systemFurniCollisions, 1)
	})
	t.Run("user", func(t *testing.T) {
		service, active, _ := projectileRoom(t, nil)
		if _, err := service.rooms.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "test", ConnectionKind: "test"}); err != nil {
			t.Fatal(err)
		}
		if _, err := active.TeleportUnit(7, grid.MustPoint(2, 1), worldunit.RotationSouth, false); err != nil {
			t.Fatal(err)
		}
		launchProjectile(t, service, 0, []int32{2, 2, 1})
		if err := service.Cycle(context.Background(), active.ID(), time.Now()); err != nil {
			t.Fatal(err)
		}
		assertVariable(t, service, active.ID(), systemUserCollisions, 1)
	})
}

// TestProjectileActorFollowsWithoutFurnitureSlotState verifies optional visual following.
func TestProjectileActorFollowsWithoutFurnitureSlotState(t *testing.T) {
	service, active, _ := projectileRoom(t, nil)
	if _, err := service.rooms.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "test", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := active.TeleportUnit(7, grid.MustPoint(1, 1), worldunit.RotationEast, false); err != nil {
		t.Fatal(err)
	}
	launchProjectile(t, service, 7, []int32{2, 1, 1})
	if err := service.Cycle(context.Background(), active.ID(), time.Now()); err != nil {
		t.Fatal(err)
	}
	unit, found := active.Unit(7)
	if !found || unit.Position.Point != grid.MustPoint(2, 1) {
		t.Fatalf("following actor=%#v found=%t", unit, found)
	}
	if _, occupied := active.SlotOccupant(10); occupied {
		t.Fatal("projectile actor corrupted furniture slot state")
	}
}

// TestProjectilePersistenceFailureRollsBackRuntimeProjection verifies furniture and rider stay aligned.
func TestProjectilePersistenceFailureRollsBackRuntimeProjection(t *testing.T) {
	service, active, mover := projectileRoom(t, nil)
	if _, err := service.rooms.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "test", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	if _, err := active.TeleportUnit(7, grid.MustPoint(1, 1), worldunit.RotationEast, false); err != nil {
		t.Fatal(err)
	}
	mover.err = errors.New("persist final")
	launchProjectile(t, service, 7, []int32{2, 1, 1})
	if err := service.Cycle(context.Background(), active.ID(), time.Now()); err == nil {
		t.Fatal("expected final persistence failure")
	}
	item, _ := active.FurnitureItem(10)
	if item.Point != grid.MustPoint(1, 1) {
		t.Fatalf("rolled back item=%#v", item)
	}
	rider, found := active.UnitMotion(7)
	if !found || rider.Position.Point != grid.MustPoint(1, 1) {
		t.Fatalf("rolled back rider=%#v found=%t", rider, found)
	}
	assertVariable(t, service, active.ID(), systemPositionX, 1)
}

// TestProjectileParametersRejectOutsideConfiguredBounds verifies runtime never clamps invalid state.
func TestProjectileParametersRejectOutsideConfiguredBounds(t *testing.T) {
	service := &Service{maximumDistance: 3, maximumDuration: 4}
	node := projectileNode([]int32{2, 2, 2})
	direction, distance, pulses, valid := service.parameters(node)
	if !valid || direction != 2 || distance != 2 || pulses != 2 {
		t.Fatalf("parameters direction=%d distance=%d pulses=%d valid=%t", direction, distance, pulses, valid)
	}
	for _, values := range [][]int32{{8, 1, 1}, {0, 0, 1}, {0, 4, 1}, {0, 3, 2}, {0, 1, 21}, {0, 1}} {
		node.Parameters.Values = values
		if _, _, _, accepted := service.parameters(node); accepted {
			t.Fatalf("invalid parameters accepted: %v", values)
		}
	}
}

// TestProjectileIdleCycleAllocatesNothing verifies the inactive room tick hot path.
func TestProjectileIdleCycleAllocatesNothing(t *testing.T) {
	service := &Service{}
	allocations := testing.AllocsPerRun(1000, func() {
		_ = service.Cycle(context.Background(), 77, time.Time{})
	})
	if allocations != 0 {
		t.Fatalf("idle allocations=%f, want 0", allocations)
	}
}

// BenchmarkProjectileIdleCycle measures the inactive room tick hot path.
func BenchmarkProjectileIdleCycle(b *testing.B) {
	service := &Service{}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		_ = service.Cycle(ctx, 77, time.Time{})
	}
}

// projectileRoom creates one active test room and bounded service.
func projectileRoom(t *testing.T, additional []worldfurniture.Item) (*Service, *roomlive.Room, *fakeMover) {
	t.Helper()
	rooms := roomlive.NewRegistry(nil, roomlive.WithTickInterval(time.Hour))
	active, err := rooms.Activate(roomlive.Snapshot{ID: 77, MaxUsers: 10})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00000\r00000\r00000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	item := worldfurniture.Item{ID: 10, OwnerPlayerID: 1, Point: grid.MustPoint(1, 1), Definition: worldfurniture.Definition{SpriteID: 10, Width: 1, Length: 1, StackHeight: 1, AllowStack: true, AllowWalk: true}}
	items := append([]worldfurniture.Item{item}, additional...)
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: items, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	mover := &fakeMover{}
	service := &Service{rooms: rooms, furniture: mover, maximum: 4, maximumDistance: 64, maximumDuration: 1280}
	t.Cleanup(func() { _, _, _ = rooms.Close(context.Background(), active.ID()) })
	return service, active, mover
}

// launchProjectile launches one target through the runtime extra boundary.
func launchProjectile(t *testing.T, service *Service, actorID int64, values []int32) {
	t.Helper()
	node := projectileNode(values)
	err := service.ExecuteExtras(context.Background(), []*configuration.Node{node}, trigger.Event{RoomID: 77, ActorID: actorID}, selection.Selection{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
}

// projectileNode creates one compiled custom extra fixture.
func projectileNode(values []int32) *configuration.Node {
	return &configuration.Node{
		Descriptor: registry.Descriptor{Key: Key}, Parameters: configuration.Parameters{Values: values},
		Targets: []record.Target{{ItemID: 10}},
	}
}

// assertVariable verifies one projectile system variable value.
func assertVariable(t *testing.T, service *Service, roomID int64, name string, expected int64) {
	t.Helper()
	value, found := service.ResolveSystem(roomID, variable.ScopeFurni, 10, name)
	if !found || value.IntValue != expected {
		t.Fatalf("variable %s=%d found=%t, want %d", name, value.IntValue, found, expected)
	}
}
