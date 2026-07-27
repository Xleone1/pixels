// Package adminaction centralizes attributed protected HTTP mutations.
package adminaction

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
)

var (
	// ErrForbidden reports a missing actor capability.
	ErrForbidden = errors.New("administrative action forbidden")
	// ErrInvalidAudit reports incomplete mutation attribution.
	ErrInvalidAudit = errors.New("actorPlayerId and reason are required")
)

// Service authorizes and atomically audits administrative mutations.
type Service struct {
	// permissions resolves effective actor capabilities.
	permissions permissionservice.Checker
	// audit persists action records and owns the shared transaction.
	audit AuditStore
}

// AuditStore persists attributed operations in one shared transaction.
type AuditStore interface {
	// WithinTransaction executes one shared database transaction.
	WithinTransaction(context.Context, func(context.Context) error) error
	// InsertAudit appends one attributed action.
	InsertAudit(context.Context, int64, string, string, string) error
}

// DetailedAuditStore persists attributed operations with state snapshots.
type DetailedAuditStore interface {
	// InsertAuditDetails appends one attributed action and its state transition.
	InsertAuditDetails(context.Context, int64, string, string, string, any, any) error
}

// Snapshot resolves one audit state inside the mutation transaction.
type Snapshot func(context.Context) (any, error)

// New creates an attributed administrative action service.
func New(
	permissions permissionservice.Checker,
	audit AuditStore,
) *Service {
	return &Service{permissions: permissions, audit: audit}
}

// Authorize verifies one actor capability.
func (service *Service) Authorize(
	ctx context.Context,
	actorPlayerID int64,
	node permission.Node,
) error {
	if service == nil || service.permissions == nil || actorPlayerID <= 0 {
		return ErrInvalidAudit
	}
	allowed, err := service.permissions.HasPermission(ctx, actorPlayerID, node)
	if err != nil {
		return fmt.Errorf("resolve administrative permission: %w", err)
	}
	if !allowed {
		return ErrForbidden
	}
	return nil
}

// Execute authorizes, commits, and audits one mutation.
func (service *Service) Execute(
	ctx context.Context,
	request Request,
	work func(context.Context) error,
) error {
	request.Reason = strings.TrimSpace(request.Reason)
	if request.TargetPlayerID <= 0 || request.Reason == "" || len(request.Reason) > 500 {
		return ErrInvalidAudit
	}
	if err := service.Authorize(ctx, request.ActorPlayerID, request.Node); err != nil {
		return err
	}
	if service.audit == nil {
		return ErrInvalidAudit
	}
	return service.audit.WithinTransaction(ctx, func(txCtx context.Context) error {
		before, err := resolveSnapshot(txCtx, request.Before)
		if err != nil {
			return fmt.Errorf("capture administrative state before mutation: %w", err)
		}
		if err := work(txCtx); err != nil {
			return err
		}
		after, err := resolveSnapshot(txCtx, request.After)
		if err != nil {
			return fmt.Errorf("capture administrative state after mutation: %w", err)
		}
		entity := fmt.Sprintf("player:%d", request.TargetPlayerID)
		if detailed, ok := service.audit.(DetailedAuditStore); ok {
			return detailed.InsertAuditDetails(
				txCtx,
				request.ActorPlayerID,
				request.Action,
				entity,
				request.Reason,
				before,
				after,
			)
		}
		return service.audit.InsertAudit(
			txCtx,
			request.ActorPlayerID,
			request.Action,
			entity,
			request.Reason,
		)
	})
}

// resolveSnapshot returns one normalized audit state.
func resolveSnapshot(ctx context.Context, snapshot Snapshot) (any, error) {
	if snapshot == nil {
		return map[string]any{}, nil
	}
	return snapshot(ctx)
}

// Request describes one permission-gated attributed action.
type Request struct {
	// Action names the durable audit operation.
	Action string
	// ActorPlayerID identifies the authenticated administrative actor.
	ActorPlayerID int64
	// Node identifies the required dotted permission.
	Node permission.Node
	// Reason explains the mutation.
	Reason string
	// TargetPlayerID identifies the affected player.
	TargetPlayerID int64
	// Before resolves state immediately before the mutation.
	Before Snapshot
	// After resolves state immediately after the mutation.
	After Snapshot
}
