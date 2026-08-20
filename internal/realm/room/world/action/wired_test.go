package action

import (
	"context"
	"testing"
	"time"

	outdance "github.com/niflaot/pixels/networking/outbound/room/entities/dance"
)

// TestValidWiredActionKeepsSemanticBounds verifies the shared client contract.
func TestValidWiredActionKeepsSemanticBounds(t *testing.T) {
	for value := int32(1); value <= 10; value++ {
		if !ValidWiredAction(value) {
			t.Fatalf("ValidWiredAction(%d) = false", value)
		}
	}
	for _, value := range []int32{-1, 0, 11, 12} {
		if ValidWiredAction(value) {
			t.Fatalf("ValidWiredAction(%d) = true", value)
		}
	}
}

// TestServiceResumeForMovementPreservesDance verifies movement clears idle without cancelling active dance.
func TestServiceResumeForMovementPreservesDance(t *testing.T) {
	active := actionRoom(t)
	connections, packets := actionConnection(t)
	service := New(Config{TransitionDelay: time.Nanosecond}, connections, nil)
	if err := service.Dance(context.Background(), active, 7, 3); err != nil {
		t.Fatal(err)
	}
	unit, _ := active.Unit(7)
	*packets = (*packets)[:0]
	if err := service.ResumeForMovement(context.Background(), active, unit); err != nil {
		t.Fatal(err)
	}
	if containsActionValue(t, *packets, outdance.Header, outdance.Definition, 0) {
		t.Fatalf("expected dance not to be cancelled on movement %#v", *packets)
	}
	active.SetUnitIdle(7, true)
	unit, _ = active.Unit(7)
	if err := service.ResumeForMovement(context.Background(), active, unit); err != nil {
		t.Fatal(err)
	}
}

// TestServiceActionsResumeIdleAvatar verifies dance, expression, sign, and posture wake an idle unit.
func TestServiceActionsResumeIdleAvatar(t *testing.T) {
	active := actionRoom(t)
	connections, _ := actionConnection(t)
	service := New(Config{TransitionDelay: time.Nanosecond}, connections, nil)
	active.SetUnitIdle(7, true)
	if err := service.Dance(context.Background(), active, 7, 2); err != nil {
		t.Fatal(err)
	}
	unit, _ := active.Unit(7)
	if unit.Idle {
		t.Fatal("expected dance to wake idle unit")
	}
	active.SetUnitIdle(7, true)
	if err := service.Express(context.Background(), active, 7, 1); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if unit.Idle {
		t.Fatal("expected express to wake idle unit")
	}
	active.SetUnitIdle(7, true)
	if err := service.Sign(context.Background(), active, 7, 3); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if unit.Idle {
		t.Fatal("expected sign to wake idle unit")
	}
	active.SetUnitIdle(7, true)
	if err := service.Posture(context.Background(), active, 7, true); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if unit.Idle {
		t.Fatal("expected posture to wake idle unit")
	}
}
