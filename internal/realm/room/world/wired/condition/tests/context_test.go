package tests

import (
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/condition"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/variable"
)

// TestContextVariableCondition verifies read-only event variables participate in comparisons.
func TestContextVariableCondition(t *testing.T) {
	node := nodeFor("wf_cnd_var_val_match")
	node.Parameters.Text = "@context.player_id"
	node.Parameters.Values = []int32{int32(variable.ScopeContext), 0, 77}
	result, err := condition.New().Evaluate(node, condition.Context{
		Event: trigger.Event{RoomID: 4, PlayerID: 77}, Now: time.Unix(10, 0),
	}, view{})
	if err != nil || !result.Valid || !result.Pass {
		t.Fatalf("context condition = %+v, %v", result, err)
	}
}
