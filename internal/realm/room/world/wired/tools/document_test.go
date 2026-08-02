package tools

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// TestVariableDocumentsPreserveExactValues verifies lossless decimal encoding and system flags.
func TestVariableDocumentsPreserveExactValues(t *testing.T) {
	now := time.Unix(10, 20)
	service := &Service{players: nil}
	documents := service.variableDocuments([]variable.Value{
		{Scope: variable.ScopeRoom, ScopeID: 4, Name: "score", IntValue: 9223372036854775807, UpdatedAt: now},
		{Scope: variable.ScopeContext, ScopeID: 9, Name: "@context.signal", StringValue: "go", UpdatedAt: now},
	})
	if len(documents) != 2 || documents[0].IntValue != "9223372036854775807" {
		t.Fatalf("unexpected variable documents %#v", documents)
	}
	if documents[1].Scope != "context" || !documents[1].ReadOnly || !documents[1].System {
		t.Fatalf("unexpected context document %#v", documents[1])
	}
}

// TestEventDocumentsExposeExecutionCaps verifies the bounded monitor summary.
func TestEventDocumentsExposeExecutionCaps(t *testing.T) {
	documents := eventDocuments([]wiredruntime.Trace{{
		ID: 7, Kind: trigger.ReceiveSignal, Stacks: 2, Effects: 3,
		Signals: 4, StackPoint: configuration.Point{X: 6, Y: 7},
		BudgetExhausted: true, BudgetCode: "SIGNAL_CAP", StartedAt: time.Unix(10, 0),
	}})
	if len(documents) != 1 || documents[0].ID != "7" || documents[0].Kind != "SIGNAL_CAP" || documents[0].StackID != "6:7" || !strings.Contains(documents[0].Message, "execution cap") {
		t.Fatalf("unexpected event documents %#v", documents)
	}
}

// TestUsageDocumentPreservesTelemetry verifies compatible runtime denominators and counters.
func TestUsageDocumentPreservesTelemetry(t *testing.T) {
	document := usageDocument(wiredruntime.Usage{
		Effects: 9, LatestTraceEffects: 3, MaxEffects: 10,
		Stacks: 7, LatestTraceStacks: 2, MaxStacks: 8,
		Signals: 4, MaxSignals: 5, CompileFailures: 9223372036854775808,
	}, 6, 12)
	if document.LatestTraceEffects != 3 || document.LatestTraceStacks != 2 || document.Signals != 4 || document.CompileFailures != "9223372036854775808" {
		t.Fatalf("unexpected usage document %#v", document)
	}
}

// TestMarshalDocumentEnforcesBudget verifies versioned documents reject oversized payloads.
func TestMarshalDocumentEnforcesBudget(t *testing.T) {
	encoded, err := marshalDocument(map[string]string{"value": "ok"})
	if err != nil || !json.Valid(encoded) {
		t.Fatalf("marshal valid document: %v", err)
	}
	if _, err = marshalDocument(map[string]string{"value": strings.Repeat("x", maximumDocumentBytes)}); err != ErrDocumentTooLarge {
		t.Fatalf("oversized document error = %v", err)
	}
}

// TestSnapshotDocumentIncludesRevision verifies the Renderer parser contract.
func TestSnapshotDocumentIncludesRevision(t *testing.T) {
	encoded, err := marshalDocument(SnapshotDocument{
		SchemaVersion: SchemaVersion,
		Room:          RoomDocument{ID: 7, Name: "QA"},
		Revision:      "13",
	})
	if err != nil || !strings.Contains(string(encoded), `"revision":"13"`) {
		t.Fatalf("marshal snapshot document: %s: %v", encoded, err)
	}
}
