package runtime

import "github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"

// Matches reports whether an active generation has a matching trigger candidate.
func (engine *Engine) Matches(event trigger.Event) bool {
	value, found := engine.rooms.Load(event.RoomID)
	if !found || !engine.config.Enabled {
		return false
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	for _, candidate := range loaded.byKind[event.Kind] {
		if engine.matcher.Match(candidate, event) {
			return true
		}
	}
	return false
}

// Conflicts returns trigger sprites incompatible with one effect's actor requirements.
func (engine *Engine) Conflicts(roomID int64, itemID int64) []int32 {
	value, found := engine.rooms.Load(roomID)
	if !found {
		return nil
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	node := loaded.generation.Nodes[itemID]
	if node == nil {
		return nil
	}
	stack := loaded.generation.Stacks[node.Point]
	result := make([]int32, 0)
	for _, candidate := range stack.Triggers {
		if node.Descriptor.Actor != 0 && candidate.Descriptor.Actor == 0 {
			result = append(result, candidate.SpriteID)
		}
	}
	return result
}
