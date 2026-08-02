package variable

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

var (
	// ErrInvalid reports a malformed variable assignment.
	ErrInvalid = errors.New("invalid WIRED variable")
	// ErrLimit reports a room variable budget exhaustion.
	ErrLimit = errors.New("WIRED variable room limit reached")
	// ErrReadOnly reports a mutation targeting an immutable system variable.
	ErrReadOnly = errors.New("WIRED system variable is read only")
)

// Service coordinates durable writes with allocation-free warmed lookups.
type Service struct {
	// store persists assignments.
	store Store
	// maximum bounds assignments per room.
	maximum int
	// system resolves immutable live room variables.
	system SystemProvider
	// writes serialize mutations per room without blocking unrelated rooms.
	writes [64]sync.Mutex
	// mutex protects loaded room maps.
	mutex sync.RWMutex
	// rooms stores warmed assignments by room.
	rooms map[int64]map[key]Value
}

// SetSystemProvider installs the immutable live variable provider during startup.
func (service *Service) SetSystemProvider(provider SystemProvider) { service.system = provider }

// New creates a bounded variable service.
func New(store Store, maximum int) *Service {
	if maximum <= 0 {
		maximum = 10000
	}
	return &Service{store: store, maximum: maximum, rooms: make(map[int64]map[key]Value)}
}

// LoadRoom replaces one room's immutable-key variable cache.
func (service *Service) LoadRoom(ctx context.Context, roomID int64) error {
	write := service.writeLock(roomID)
	write.Lock()
	defer write.Unlock()
	values, err := service.store.LoadRoom(ctx, roomID)
	if err != nil {
		return err
	}
	loaded := make(map[key]Value, len(values))
	for _, value := range values {
		value.Name = normalize(value.Name)
		loaded[key{scope: value.Scope, scopeID: value.ScopeID, name: value.Name}] = value
	}
	service.mutex.Lock()
	service.rooms[roomID] = loaded
	service.mutex.Unlock()
	return nil
}

// Close releases one room's warmed variable state.
func (service *Service) Close(roomID int64) {
	write := service.writeLock(roomID)
	write.Lock()
	defer write.Unlock()
	service.mutex.Lock()
	delete(service.rooms, roomID)
	service.mutex.Unlock()
}

// Exists reports whether one warmed assignment exists.
func (service *Service) Exists(roomID int64, scope Scope, scopeID int64, name string) bool {
	if strings.HasPrefix(name, "@") && service.system != nil {
		_, found := service.system.ResolveSystem(roomID, scope, scopeID, name)
		return found
	}
	service.mutex.RLock()
	values := service.rooms[roomID]
	_, found := values[key{scope: scope, scopeID: scopeID, name: name}]
	service.mutex.RUnlock()
	return found
}

// Get returns one warmed assignment.
func (service *Service) Get(roomID int64, scope Scope, scopeID int64, name string) (Value, bool) {
	if strings.HasPrefix(name, "@") && service.system != nil {
		return service.system.ResolveSystem(roomID, scope, scopeID, name)
	}
	service.mutex.RLock()
	value, found := service.rooms[roomID][key{scope: scope, scopeID: scopeID, name: name}]
	service.mutex.RUnlock()
	return value, found
}

// List returns stable durable and system variables for Creator Tools inspection.
func (service *Service) List(roomID int64, scope Scope, scopeID int64) []Value {
	service.mutex.RLock()
	values := service.rooms[roomID]
	result := make([]Value, 0, len(values))
	for currentKey, value := range values {
		if (scope == 0 || currentKey.scope == scope) && (scopeID == 0 || currentKey.scopeID == scopeID) {
			result = append(result, value)
		}
	}
	service.mutex.RUnlock()
	if service.system != nil && scope > 0 && scopeID > 0 {
		result = append(result, service.system.ListSystem(roomID, scope, scopeID)...)
	}
	sort.Slice(result, func(left int, right int) bool {
		if result[left].Scope != result[right].Scope {
			return result[left].Scope < result[right].Scope
		}
		if result[left].ScopeID != result[right].ScopeID {
			return result[left].ScopeID < result[right].ScopeID
		}
		return result[left].Name < result[right].Name
	})
	return result
}

// GetDurable reads warmed state first and persistence only for cold references.
func (service *Service) GetDurable(
	ctx context.Context,
	roomID int64,
	scope Scope,
	scopeID int64,
	name string,
) (Value, bool, error) {
	name = normalize(name)
	if value, found := service.Get(roomID, scope, scopeID, name); found {
		return value, true, nil
	}
	finder, supported := service.store.(Finder)
	if !supported {
		return Value{}, false, nil
	}
	return finder.Find(ctx, roomID, scope, scopeID, name)
}

// Set persists and publishes one assignment.
func (service *Service) Set(ctx context.Context, value Value) (Value, error) {
	if strings.HasPrefix(normalize(value.Name), "@") {
		return Value{}, ErrReadOnly
	}
	write := service.writeLock(value.RoomID)
	write.Lock()
	defer write.Unlock()
	return service.set(ctx, value)
}

// set persists one assignment while its room write lock is held.
func (service *Service) set(ctx context.Context, value Value) (Value, error) {
	value.Name = normalize(value.Name)
	if !valid(value) {
		return Value{}, ErrInvalid
	}
	service.mutex.RLock()
	values := service.rooms[value.RoomID]
	_, exists := values[key{scope: value.Scope, scopeID: value.ScopeID, name: value.Name}]
	full := !exists && len(values) >= service.maximum
	service.mutex.RUnlock()
	if full {
		return Value{}, ErrLimit
	}
	saved, err := service.store.Set(ctx, value)
	if err != nil {
		return Value{}, err
	}
	service.mutex.Lock()
	values = service.rooms[value.RoomID]
	if values != nil {
		values[key{scope: saved.Scope, scopeID: saved.ScopeID, name: saved.Name}] = saved
	}
	service.mutex.Unlock()
	return saved, nil
}

// Change adds a delta to one assignment, creating it at zero when absent.
func (service *Service) Change(ctx context.Context, value Value, delta int64) (Value, int64, error) {
	if strings.HasPrefix(normalize(value.Name), "@") {
		return Value{}, 0, ErrReadOnly
	}
	write := service.writeLock(value.RoomID)
	write.Lock()
	defer write.Unlock()
	current, found, err := service.GetDurable(
		ctx,
		value.RoomID,
		value.Scope,
		value.ScopeID,
		normalize(value.Name),
	)
	if err != nil {
		return Value{}, 0, err
	}
	previous := int64(0)
	if found {
		previous = current.IntValue
		value.CreatedAt = current.CreatedAt
	}
	value.IntValue = previous + delta
	saved, err := service.set(ctx, value)
	return saved, previous, err
}

// Delete removes one assignment from persistence and cache.
func (service *Service) Delete(ctx context.Context, roomID int64, scope Scope, scopeID int64, name string) (Value, bool, error) {
	if strings.HasPrefix(normalize(name), "@") {
		return Value{}, false, ErrReadOnly
	}
	write := service.writeLock(roomID)
	write.Lock()
	defer write.Unlock()
	name = normalize(name)
	previous, found, err := service.GetDurable(
		ctx, roomID, scope, scopeID, name,
	)
	if err != nil {
		return Value{}, false, err
	}
	if !found {
		return Value{}, false, nil
	}
	deleted, err := service.store.Delete(ctx, roomID, scope, scopeID, name)
	if err != nil || !deleted {
		return Value{}, deleted, err
	}
	service.mutex.Lock()
	delete(service.rooms[roomID], key{scope: scope, scopeID: scopeID, name: name})
	service.mutex.Unlock()
	return previous, true, nil
}
