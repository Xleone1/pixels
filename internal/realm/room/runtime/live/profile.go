package live

import (
	"time"

	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// Profile contains bounded observability for one active room.
type Profile struct {
	// Available reports whether the room runtime is loaded.
	Available bool `json:"available"`
	// LoadedAt stores when the room runtime opened.
	LoadedAt time.Time `json:"loadedAt"`
	// UptimeMilliseconds stores runtime age.
	UptimeMilliseconds int64 `json:"uptimeMilliseconds"`
	// WorldLoaded reports whether layout and fixtures are available.
	WorldLoaded bool `json:"worldLoaded"`
	// Occupants stores active room sessions.
	Occupants int `json:"occupants"`
	// PlayerUnits stores player-controlled world units.
	PlayerUnits int `json:"playerUnits"`
	// BotUnits stores loaded bot units.
	BotUnits int `json:"botUnits"`
	// PetUnits stores loaded pet units.
	PetUnits int `json:"petUnits"`
	// Furniture stores loaded furniture items.
	Furniture int `json:"furniture"`
	// Fixtures stores indexed furniture footprints.
	Fixtures int `json:"fixtures"`
	// PendingTasks stores scheduled room-owned callbacks.
	PendingTasks int `json:"pendingTasks"`
	// GridWidth stores layout columns.
	GridWidth int `json:"gridWidth"`
	// GridHeight stores layout rows.
	GridHeight int `json:"gridHeight"`
	// ValidTiles stores traversable layout tiles.
	ValidTiles int `json:"validTiles"`
	// TickCount stores completed owner-loop cycles.
	TickCount uint64 `json:"tickCount"`
	// TickErrors stores publisher failures observed by the owner loop.
	TickErrors uint64 `json:"tickErrors"`
	// LastTickMicroseconds stores the latest cycle duration.
	LastTickMicroseconds int64 `json:"lastTickMicroseconds"`
	// AverageTickMicroseconds stores mean cycle duration.
	AverageTickMicroseconds int64 `json:"averageTickMicroseconds"`
	// MaxTickMicroseconds stores the slowest cycle duration.
	MaxTickMicroseconds int64 `json:"maxTickMicroseconds"`
	// LastTickAt stores the latest completed cycle time.
	LastTickAt *time.Time `json:"lastTickAt,omitempty"`
	// EstimatedStateBytes stores a conservative runtime-state estimate.
	EstimatedStateBytes int64 `json:"estimatedStateBytes"`
}

// Profile returns one bounded active-room profiling snapshot.
func (room *Room) Profile() Profile {
	room.mutex.RLock()
	profile := Profile{Available: true, LoadedAt: room.loadedAt, UptimeMilliseconds: time.Since(room.loadedAt).Milliseconds(), Occupants: len(room.occupants), PendingTasks: room.tasks.Len()}
	if room.world != nil {
		metrics := room.world.Metrics()
		profile.WorldLoaded = true
		profile.Furniture = metrics.Furniture
		profile.Fixtures = metrics.Fixtures
		profile.GridWidth = metrics.GridWidth
		profile.GridHeight = metrics.GridHeight
		profile.ValidTiles = metrics.ValidTiles
		for _, unit := range room.world.Units() {
			switch unit.Kind {
			case worldunit.KindPlayer:
				profile.PlayerUnits++
			case worldunit.KindBot:
				profile.BotUnits++
			case worldunit.KindPet:
				profile.PetUnits++
			}
		}
	}
	room.mutex.RUnlock()
	profile.TickCount = room.tickCount.Load()
	profile.TickErrors = room.tickErrors.Load()
	profile.LastTickMicroseconds = room.lastTickNanos.Load() / int64(time.Microsecond)
	profile.MaxTickMicroseconds = room.maxTickNanos.Load() / int64(time.Microsecond)
	if profile.TickCount > 0 {
		profile.AverageTickMicroseconds = room.totalTickNanos.Load() / int64(profile.TickCount) / int64(time.Microsecond)
	}
	if value := room.lastTickUnixNanos.Load(); value > 0 {
		last := time.Unix(0, value).UTC()
		profile.LastTickAt = &last
	}
	profile.EstimatedStateBytes = int64(1024 + profile.Furniture*384 + profile.Fixtures*96 + (profile.PlayerUnits+profile.BotUnits+profile.PetUnits)*512 + profile.PendingTasks*96)
	return profile
}

// recordTick stores one completed owner-loop observation.
func (room *Room) recordTick(started time.Time, failed bool) {
	elapsed := time.Since(started).Nanoseconds()
	room.tickCount.Add(1)
	room.totalTickNanos.Add(elapsed)
	room.lastTickNanos.Store(elapsed)
	room.lastTickUnixNanos.Store(time.Now().UnixNano())
	if failed {
		room.tickErrors.Add(1)
	}
	for current := room.maxTickNanos.Load(); elapsed > current && !room.maxTickNanos.CompareAndSwap(current, elapsed); current = room.maxTickNanos.Load() {
	}
}
