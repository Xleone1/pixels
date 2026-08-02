package tools

import (
	"context"
	"strconv"
	"strings"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// variableRevision returns an exact revision that changes with durable variable writes.
func variableRevision(values []variable.Value, fallback int64) string {
	revision := fallback
	for _, value := range values {
		if candidate := value.UpdatedAt.UnixNano(); candidate > revision {
			revision = candidate
		}
	}
	return strconv.FormatInt(revision, 10)
}

// validateMutation validates one variable target against authoritative room state.
func (service *Service) validateMutation(ctx context.Context, roomID int64, scopeValue int32, scopeID int64, name string) (variable.Scope, int64, string, error) {
	scope := variable.Scope(scopeValue)
	name = strings.ToLower(strings.TrimSpace(name))
	if scope < variable.ScopeFurni || scope > variable.ScopeReference || name == "" || len(name) > 64 || strings.HasPrefix(name, "@") {
		return 0, 0, "", ErrInvalidRequest
	}
	active, found := service.rooms.Find(roomID)
	if !found {
		return 0, 0, "", ErrInvalidRequest
	}
	switch scope {
	case variable.ScopeFurni:
		if _, found = active.FurnitureItem(scopeID); !found {
			return 0, 0, "", ErrInvalidRequest
		}
	case variable.ScopeUser:
		if _, found = active.Occupant(scopeID); !found {
			return 0, 0, "", ErrInvalidRequest
		}
	case variable.ScopeRoom:
		if scopeID == 0 {
			scopeID = roomID
		}
		if scopeID != roomID {
			return 0, 0, "", ErrInvalidRequest
		}
	case variable.ScopeReference:
		if scopeID <= 0 {
			return 0, 0, "", ErrInvalidRequest
		}
		if _, found, err := service.records.FindByID(ctx, scopeID); err != nil || !found {
			return 0, 0, "", ErrInvalidRequest
		}
	}
	return scope, scopeID, name, nil
}

// preferencesFor returns panel preferences with the durable visibility setting.
func (service *Service) preferencesFor(playerID int64, roomID int64, hideBoxes bool) Preferences {
	key := preferenceKey{playerID: playerID, roomID: roomID}
	service.preferencesMutex.RLock()
	preferences, found := service.preferences[key]
	service.preferencesMutex.RUnlock()
	if !found {
		preferences = Preferences{}
	}
	preferences.HideBoxes = hideBoxes
	return preferences
}

// setPreferences stores process-local panel preferences.
func (service *Service) setPreferences(playerID int64, roomID int64, preferences Preferences) {
	key := preferenceKey{playerID: playerID, roomID: roomID}
	service.preferencesMutex.Lock()
	service.preferences[key] = preferences
	service.preferencesMutex.Unlock()
}
