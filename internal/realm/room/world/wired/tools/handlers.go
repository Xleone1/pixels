package tools

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininspect "github.com/niflaot/pixels/networking/inbound/furniture/wired/tools/inspect"
	inmutate "github.com/niflaot/pixels/networking/inbound/furniture/wired/tools/mutate"
	insettings "github.com/niflaot/pixels/networking/inbound/furniture/wired/tools/settings"
	insnapshot "github.com/niflaot/pixels/networking/inbound/furniture/wired/tools/snapshot"
	outinspection "github.com/niflaot/pixels/networking/outbound/furniture/wired/tools/inspection"
	outresult "github.com/niflaot/pixels/networking/outbound/furniture/wired/tools/result"
	outsnapshot "github.com/niflaot/pixels/networking/outbound/furniture/wired/tools/snapshot"
)

// RegisterHandlers installs every Creator Tools packet adapter.
func RegisterHandlers(registry *netconn.HandlerRegistry, service *Service) {
	if registry == nil || service == nil {
		return
	}
	_ = registry.Register(insnapshot.Header, service.handleSnapshot)
	_ = registry.Register(inmutate.Header, service.handleMutate)
	_ = registry.Register(ininspect.Header, service.handleInspect)
	_ = registry.Register(insettings.Header, service.handleSettings)
}

// handleSnapshot returns one bounded room document.
func (service *Service) handleSnapshot(connection netconn.Context, packet codec.Packet) error {
	request, err := insnapshot.Decode(packet)
	if err != nil {
		return err
	}
	ctx := context.Background()
	playerID, permissions, err := service.Actor(ctx, connection, request.RoomID)
	if err != nil {
		return service.sendFailure(ctx, connection, err)
	}
	document, err := service.Snapshot(ctx, playerID, request.RoomID, permissions)
	if err != nil {
		return service.sendFailure(ctx, connection, err)
	}
	response, err := outsnapshot.Encode(SchemaVersion, string(document))
	if err != nil {
		return err
	}
	return connection.Send(ctx, response)
}

// handleMutate applies one variable mutation.
func (service *Service) handleMutate(connection netconn.Context, packet codec.Packet) error {
	request, err := inmutate.Decode(packet)
	if err != nil {
		return err
	}
	ctx := context.Background()
	playerID, permissions, err := service.Actor(ctx, connection, request.RoomID)
	if err != nil {
		return service.sendFailure(ctx, connection, err)
	}
	document, code, err := service.Mutate(ctx, playerID, request.RoomID, request.Operation, request.Scope, request.ScopeID, request.Name, request.IntValue, request.StringValue, permissions)
	if err != nil {
		return service.sendResult(ctx, connection, false, code, nil)
	}
	if code == "not_found" {
		return service.sendResult(ctx, connection, false, code, nil)
	}
	return service.sendResult(ctx, connection, true, code, document)
}

// handleInspect returns one entity document.
func (service *Service) handleInspect(connection netconn.Context, packet codec.Packet) error {
	request, err := ininspect.Decode(packet)
	if err != nil {
		return err
	}
	ctx := context.Background()
	_, permissions, err := service.Actor(ctx, connection, request.RoomID)
	if err != nil {
		return service.sendFailure(ctx, connection, err)
	}
	document, err := service.Inspect(request.RoomID, request.EntityType, request.EntityID, permissions)
	if err != nil {
		return service.sendFailure(ctx, connection, err)
	}
	response, err := outinspection.Encode(SchemaVersion, string(document))
	if err != nil {
		return err
	}
	return connection.Send(ctx, response)
}

// handleSettings persists room and panel settings.
func (service *Service) handleSettings(connection netconn.Context, packet codec.Packet) error {
	request, err := insettings.Decode(packet)
	if err != nil {
		return err
	}
	ctx := context.Background()
	playerID, permissions, err := service.Actor(ctx, connection, request.RoomID)
	if err != nil {
		return service.sendFailure(ctx, connection, err)
	}
	document, err := service.Settings(ctx, playerID, request.RoomID, request.LiveUpdates, request.HideBoxes, permissions)
	if err != nil {
		return service.sendFailure(ctx, connection, err)
	}
	return service.sendResult(ctx, connection, true, "settings_saved", document)
}

// sendFailure sends one stable non-disconnecting failure.
func (service *Service) sendFailure(ctx context.Context, connection netconn.Context, err error) error {
	code := "internal_error"
	if errors.Is(err, ErrForbidden) {
		code = "forbidden"
	} else if errors.Is(err, ErrInvalidRequest) || errors.Is(err, ErrDocumentTooLarge) {
		code = "invalid_request"
	}
	return service.sendResult(ctx, connection, false, code, nil)
}

// sendResult sends one mutation or protocol-status document.
func (service *Service) sendResult(ctx context.Context, connection netconn.Context, success bool, code string, document []byte) error {
	response, err := outresult.Encode(success, code, string(document))
	if err != nil {
		return err
	}
	return connection.Send(ctx, response)
}
