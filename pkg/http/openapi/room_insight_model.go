package openapi

// RoomInsightRequest contains an attributed room observability read.
type RoomInsightRequest struct {
	RoomIDRequest
	// ActorPlayerID identifies the authorized administrative actor.
	ActorPlayerID int64 `query:"actorPlayerId" required:"true" minimum:"1"`
}

// RoomStatsResponse describes durable room content counters.
type RoomStatsResponse struct {
	// RoomID identifies the room.
	RoomID int64 `json:"roomId" required:"true"`
	// Furniture stores every placed furniture item.
	Furniture int `json:"furniture" required:"true"`
	// FloorFurniture stores placed floor furniture.
	FloorFurniture int `json:"floorFurniture" required:"true"`
	// WallFurniture stores placed wall furniture.
	WallFurniture int `json:"wallFurniture" required:"true"`
	// WiredFurniture stores placed WIRED furniture without its configuration.
	WiredFurniture int `json:"wiredFurniture" required:"true"`
	// Bots stores bots assigned to the room.
	Bots int `json:"bots" required:"true"`
	// Pets stores pets assigned to the room.
	Pets int `json:"pets" required:"true"`
	// Branding stores enabled branding configurations.
	Branding int `json:"branding" required:"true"`
	// Occupants stores current live occupancy.
	Occupants int `json:"occupants" required:"true"`
	// MaxUsers stores room capacity.
	MaxUsers int `json:"maxUsers" required:"true"`
	// RuntimeAvailable reports whether the room is loaded.
	RuntimeAvailable bool `json:"runtimeAvailable" required:"true"`
}

// RoomProfileResponse describes bounded live room profiling.
type RoomProfileResponse struct {
	// Available reports whether the room runtime is loaded.
	Available bool `json:"available" required:"true"`
	// LoadedAt stores when the active runtime opened.
	LoadedAt string `json:"loadedAt,omitempty" format:"date-time"`
	// UptimeMilliseconds stores active runtime age.
	UptimeMilliseconds int64 `json:"uptimeMilliseconds,omitempty"`
	// WorldLoaded reports whether layout and fixtures are loaded.
	WorldLoaded bool `json:"worldLoaded,omitempty"`
	// Occupants stores active sessions.
	Occupants int `json:"occupants,omitempty"`
	// PlayerUnits stores player-controlled units.
	PlayerUnits int `json:"playerUnits,omitempty"`
	// BotUnits stores loaded bot units.
	BotUnits int `json:"botUnits,omitempty"`
	// PetUnits stores loaded pet units.
	PetUnits int `json:"petUnits,omitempty"`
	// Furniture stores loaded furniture items.
	Furniture int `json:"furniture,omitempty"`
	// Fixtures stores indexed furniture footprints.
	Fixtures int `json:"fixtures,omitempty"`
	// PendingTasks stores room-owned scheduled callbacks.
	PendingTasks int `json:"pendingTasks,omitempty"`
	// GridWidth stores layout columns.
	GridWidth int `json:"gridWidth,omitempty"`
	// GridHeight stores layout rows.
	GridHeight int `json:"gridHeight,omitempty"`
	// ValidTiles stores traversable layout tiles.
	ValidTiles int `json:"validTiles,omitempty"`
	// TickCount stores completed owner-loop cycles.
	TickCount uint64 `json:"tickCount,omitempty"`
	// TickErrors stores owner-loop publisher errors.
	TickErrors uint64 `json:"tickErrors,omitempty"`
	// LastTickMicroseconds stores the latest cycle duration.
	LastTickMicroseconds int64 `json:"lastTickMicroseconds,omitempty"`
	// AverageTickMicroseconds stores mean cycle duration.
	AverageTickMicroseconds int64 `json:"averageTickMicroseconds,omitempty"`
	// MaxTickMicroseconds stores the slowest observed cycle.
	MaxTickMicroseconds int64 `json:"maxTickMicroseconds,omitempty"`
	// LastTickAt stores the latest cycle boundary.
	LastTickAt string `json:"lastTickAt,omitempty" format:"date-time"`
	// EstimatedStateBytes stores a conservative state estimate.
	EstimatedStateBytes int64 `json:"estimatedStateBytes,omitempty"`
}
