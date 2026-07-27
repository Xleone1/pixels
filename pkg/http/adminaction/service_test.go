package adminaction

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
)

// actionChecker returns one configured permission decision.
type actionChecker struct {
	// allowed stores the permission decision.
	allowed bool
}

// HasPermission returns the configured decision.
func (checker actionChecker) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return checker.allowed, nil
}

// actionAudit records transaction and audit calls.
type actionAudit struct {
	// action stores the recorded operation.
	action string
	// reason stores the recorded explanation.
	reason string
	// before stores the recorded prior state.
	before any
	// after stores the recorded resulting state.
	after any
}

// WithinTransaction executes one test transaction.
func (audit *actionAudit) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// InsertAudit records one attributed action.
func (audit *actionAudit) InsertAudit(_ context.Context, _ int64, action string, _ string, reason string) error {
	audit.action = action
	audit.reason = reason
	return nil
}

// InsertAuditDetails records one attributed state transition.
func (audit *actionAudit) InsertAuditDetails(_ context.Context, _ int64, action string, _ string, reason string, before any, after any) error {
	audit.action = action
	audit.reason = reason
	audit.before = before
	audit.after = after
	return nil
}

// TestExecuteAuthorizesMutationAndAudit verifies the shared security boundary.
func TestExecuteAuthorizesMutationAndAudit(t *testing.T) {
	audit := &actionAudit{}
	service := New(actionChecker{allowed: true}, audit)
	called := false
	err := service.Execute(context.Background(), Request{
		Action: "player.teleport", ActorPlayerID: 7,
		Node:   permission.Node("moderation.room.override"),
		Reason: "support request", TargetPlayerID: 8,
		Before: func(context.Context) (any, error) {
			return map[string]any{"roomId": int64(1)}, nil
		},
		After: func(context.Context) (any, error) {
			return map[string]any{"roomId": int64(2)}, nil
		},
	}, func(context.Context) error {
		called = true
		return nil
	})
	if err != nil || !called || audit.action != "player.teleport" || audit.reason != "support request" {
		t.Fatalf("err=%v called=%t audit=%#v", err, called, audit)
	}
	if audit.before == nil || audit.after == nil {
		t.Fatalf("missing audit transition: %#v", audit)
	}
}

// TestExecuteRejectsForbiddenActorBeforeMutation verifies fail-closed ordering.
func TestExecuteRejectsForbiddenActorBeforeMutation(t *testing.T) {
	service := New(actionChecker{}, &actionAudit{})
	called := false
	err := service.Execute(context.Background(), Request{
		Action: "currency.grant", ActorPlayerID: 7,
		Node:   permission.Node("currency.admin.manage"),
		Reason: "support request", TargetPlayerID: 8,
	}, func(context.Context) error {
		called = true
		return nil
	})
	if !errors.Is(err, ErrForbidden) || called {
		t.Fatalf("err=%v called=%t", err, called)
	}
}
