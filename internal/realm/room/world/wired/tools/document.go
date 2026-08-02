package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// snapshot builds one bounded room document from immutable snapshots.
func (service *Service) snapshot(ctx context.Context, playerID int64, roomID int64, permissions PermissionsDocument) ([]byte, error) {
	active, found := service.rooms.Find(roomID)
	if !found {
		return nil, ErrInvalidRequest
	}
	record, found, err := service.records.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, ErrInvalidRequest
	}
	usage, _ := service.runtime.Usage(roomID)
	values := service.variables.List(roomID, 0, 0)
	revision := variableRevision(values, record.Version.Version)
	durableCount := len(values)
	for _, value := range service.variables.List(roomID, variable.ScopeRoom, roomID) {
		if strings.HasPrefix(value.Name, "@") {
			values = append(values, value)
		}
	}
	if latest, available := service.runtime.Context(roomID); available {
		values = append(values, variable.ListContext(latest, time.Now())...)
	}
	if len(values) > maximumVariables {
		values = values[:maximumVariables]
	}
	document := SnapshotDocument{
		SchemaVersion: SchemaVersion,
		Room:          RoomDocument{ID: roomID, Name: record.Name},
		Revision:      revision,
		Usage:         usageDocument(usage, durableCount, service.config.MaxFurniPerRoom),
		Events:        eventDocuments(service.runtime.Traces(roomID)),
		Variables:     service.variableDocuments(values),
		Entities:      entityDocuments(active),
		Permissions:   permissions,
		Settings:      service.preferencesFor(playerID, roomID, record.HideWired),
	}
	return marshalDocument(document)
}

// inspection builds one entity document with durable and derived variables.
func (service *Service) inspection(roomID int64, entityType string, entityID int64, permissions PermissionsDocument) ([]byte, error) {
	active, found := service.rooms.Find(roomID)
	if !found {
		return nil, ErrInvalidRequest
	}
	entity, scope, found := inspectableEntity(active, entityType, entityID)
	if !found {
		return nil, ErrInvalidRequest
	}
	values := service.variables.List(roomID, scope, entityID)
	if len(values) > maximumVariables {
		values = values[:maximumVariables]
	}
	document := InspectionDocument{
		SchemaVersion: SchemaVersion, Revision: variableRevision(values, 0), Entity: entity,
		Variables: service.variableDocuments(values), Permissions: permissions,
	}
	return marshalDocument(document)
}

// marshalDocument serializes and enforces the protocol string budget.
func marshalDocument(document any) ([]byte, error) {
	encoded, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}
	if len(encoded) > maximumDocumentBytes {
		return nil, ErrDocumentTooLarge
	}
	return encoded, nil
}

// usageDocument maps internal telemetry into the stable JSON contract.
func usageDocument(usage wiredruntime.Usage, variables int, maxWired int) UsageDocument {
	return UsageDocument{
		WiredFurni: usage.WiredFurni, MaxWiredFurni: maxWired, Effects: usage.Effects,
		LatestTraceEffects: usage.LatestTraceEffects, MaxEffects: usage.MaxEffects,
		Stacks: usage.Stacks, LatestTraceStacks: usage.LatestTraceStacks, MaxStacks: usage.MaxStacks,
		Delayed: usage.Delayed, MaxDelayed: usage.MaxDelayed, Variables: variables,
		MaxVariables: usage.MaxVariables, Signals: usage.Signals, MaxSignals: usage.MaxSignals,
		ExecutionMicros: usage.ExecutionMicros,
		CompileFailures: strconv.FormatUint(usage.CompileFailures, 10),
	}
}

// eventDocuments maps the fixed trace ring to stable rows.
func eventDocuments(traces []wiredruntime.Trace) []EventDocument {
	result := make([]EventDocument, 0, len(traces))
	for _, trace := range traces {
		message := fmt.Sprintf("%d stacks · %d effects", trace.Stacks, trace.Effects)
		kind := trigger.Label(trace.Kind)
		if trace.BudgetExhausted {
			kind = trace.BudgetCode
			if kind == "" {
				kind = "EXECUTION_CAP"
			}
			message += " · execution cap"
		}
		stackID := ""
		if trace.Stacks > 0 {
			stackID = strconv.Itoa(trace.StackPoint.X) + ":" + strconv.Itoa(trace.StackPoint.Y)
		}
		result = append(result, EventDocument{
			ID: strconv.FormatUint(trace.ID, 10), Kind: kind,
			OccurredAt: trace.StartedAt.UTC().Format(time.RFC3339Nano), Message: message,
			StackID: stackID,
		})
	}
	return result
}

// variableDocuments maps exact values without lossy numeric JSON conversion.
func (service *Service) variableDocuments(values []variable.Value) []VariableDocument {
	result := make([]VariableDocument, 0, len(values))
	for _, value := range values {
		system := strings.HasPrefix(value.Name, "@")
		valueType := "integer"
		if value.StringValue != "" {
			valueType = "string"
		}
		updatedAt, revision := "", "0"
		if !value.UpdatedAt.IsZero() {
			updatedAt = value.UpdatedAt.UTC().Format(time.RFC3339Nano)
			revision = strconv.FormatInt(value.UpdatedAt.UnixNano(), 10)
		}
		creator := "WIRED"
		if system {
			creator = "Sistema"
		} else if value.UpdatedByPlayerID > 0 {
			creator = "Jugador #" + strconv.FormatInt(value.UpdatedByPlayerID, 10)
			if player, found := service.players.Find(value.UpdatedByPlayerID); found {
				creator = player.Username()
			}
		}
		result = append(result, VariableDocument{
			Scope: scopeName(value.Scope), ScopeID: strconv.FormatInt(value.ScopeID, 10),
			ScopeName: variableScopeName(value), Name: value.Name, Type: valueType,
			IntValue: strconv.FormatInt(value.IntValue, 10), StringValue: value.StringValue,
			CreatorName: creator, UpdatedAt: updatedAt, Revision: revision,
			ReadOnly: system, System: system,
		})
	}
	return result
}

// scopeName maps durable scopes to Creator Tools labels.
func scopeName(scope variable.Scope) string {
	switch scope {
	case variable.ScopeFurni:
		return "furni"
	case variable.ScopeUser:
		return "user"
	case variable.ScopeRoom:
		return "room"
	case variable.ScopeContext:
		return "context"
	default:
		return "reference"
	}
}

// variableScopeName returns a compact human-readable target label.
func variableScopeName(value variable.Value) string {
	identifier := strconv.FormatInt(value.ScopeID, 10)
	switch value.Scope {
	case variable.ScopeFurni:
		return "Furni #" + identifier
	case variable.ScopeUser:
		return "Usuario #" + identifier
	case variable.ScopeRoom:
		return "Sala #" + identifier
	case variable.ScopeContext:
		return "Contexto del evento"
	default:
		return "Referencia #" + identifier
	}
}

// entityDocuments lists room furniture and players in stable room snapshot order.
func entityDocuments(active *roomlive.Room) []EntityDocument {
	items := active.FurnitureItems()
	presences := active.Presences()
	capacity := len(items) + len(presences)
	if capacity > maximumEntities {
		capacity = maximumEntities
	}
	result := make([]EntityDocument, 0, capacity)
	for _, item := range items {
		if len(result) >= maximumEntities {
			return result
		}
		result = append(result, EntityDocument{
			Type: "furni", ID: strconv.FormatInt(item.ID, 10),
			Name: item.Definition.InteractionType,
		})
	}
	for _, presence := range presences {
		if len(result) >= maximumEntities {
			break
		}
		result = append(result, EntityDocument{
			Type: "user", ID: strconv.FormatInt(presence.Occupant.PlayerID, 10),
			Name: presence.Occupant.Username,
		})
	}
	return result
}

// inspectableEntity resolves one exact active entity and variable scope.
func inspectableEntity(active *roomlive.Room, entityType string, entityID int64) (EntityDocument, variable.Scope, bool) {
	switch strings.ToLower(strings.TrimSpace(entityType)) {
	case "furni":
		item, found := active.FurnitureItem(entityID)
		if !found {
			return EntityDocument{}, 0, false
		}
		return EntityDocument{Type: "furni", ID: strconv.FormatInt(entityID, 10), Name: item.Definition.InteractionType}, variable.ScopeFurni, true
	case "user":
		occupant, found := active.Occupant(entityID)
		if !found {
			return EntityDocument{}, 0, false
		}
		return EntityDocument{Type: "user", ID: strconv.FormatInt(entityID, 10), Name: occupant.Username}, variable.ScopeUser, true
	default:
		return EntityDocument{}, 0, false
	}
}
