package runtime

import (
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// withVariableReference binds a colocated remote definition to one stack event.
func withVariableReference(
	stack *configuration.Stack,
	event trigger.Event,
) trigger.Event {
	for _, definition := range stack.Variables {
		if definition.Descriptor.Key != "wf_var_reference" ||
			len(definition.Parameters.Values) == 0 ||
			definition.Parameters.Values[0] <= 0 ||
			definition.Parameters.Text == "" {
			continue
		}
		event.ReferenceRoomID = int64(definition.Parameters.Values[0])
		event.ReferenceVariable = definition.Parameters.Text
		return event
	}
	return event
}
