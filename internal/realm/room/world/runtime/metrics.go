package runtime

// Metrics contains allocation-free world cardinalities used by room profiling.
type Metrics struct {
	// GridWidth stores layout columns.
	GridWidth int
	// GridHeight stores layout rows.
	GridHeight int
	// ValidTiles stores traversable layout tiles.
	ValidTiles int
	// Furniture stores loaded furniture items.
	Furniture int
	// Fixtures stores indexed furniture footprints.
	Fixtures int
	// Units stores loaded room units.
	Units int
}

// Metrics returns stable world cardinalities.
func (world *World) Metrics() Metrics {
	return Metrics{
		GridWidth:  int(world.grid.Width()),
		GridHeight: int(world.grid.Height()),
		ValidTiles: world.grid.ValidCount(),
		Furniture:  len(world.furniture),
		Fixtures:   len(world.furnitureTiles),
		Units:      len(world.units),
	}
}
