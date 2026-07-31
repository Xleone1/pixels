package selection

import (
	"context"
	"slices"
	"strings"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// resolve evaluates a stack and bounds remote-selector recursion.
func (resolver *Resolver) resolve(ctx context.Context, stack *configuration.Stack, event trigger.Event, generation *configuration.Generation, view View, depth int) (Selection, error) {
	if stack == nil || view == nil || len(stack.Selectors) == 0 || depth > 4 {
		return Selection{}, nil
	}
	result := Selection{}
	var furniSeen idSet
	var actorSeen idSet
	for _, node := range stack.Selectors {
		if err := ctx.Err(); err != nil {
			return Selection{}, err
		}
		furni, actors := resolver.resolveNode(ctx, node, event, generation, view, depth)
		furniSelector, actorSelector := selectorFamilies(node.Descriptor.Key)
		result.FurniResolved = result.FurniResolved || furniSelector
		result.ActorsResolved = result.ActorsResolved || actorSelector
		appendFurni(&result, &furniSeen, furni, resolver.limit)
		appendActors(&result, &actorSeen, actors, resolver.limit)
	}
	sortSelection(&result)
	return result, nil
}

// selectorFamilies reports which target collections one selector replaces.
func selectorFamilies(key string) (bool, bool) {
	if key == "wf_slc_remote" {
		return true, true
	}
	if strings.HasPrefix(key, "wf_slc_furni_") {
		return true, false
	}
	if strings.HasPrefix(key, "wf_slc_users_") {
		return false, true
	}
	return false, false
}

// resolveNode resolves one selector descriptor.
func (resolver *Resolver) resolveNode(ctx context.Context, node *configuration.Node, event trigger.Event, generation *configuration.Generation, view View, depth int) ([]record.Target, []int64) {
	values := node.Parameters.Values
	switch node.Descriptor.Key {
	case "wf_slc_users_byaction":
		if event.Kind == trigger.UserPerformsAction &&
			event.Action == valueAt(values, 0) &&
			event.ActorID > 0 {
			return nil, []int64{event.ActorID}
		}
		return nil, view.UsersByAction(valueAt(values, 0))
	case "wf_slc_users_byname":
		return nil, view.UsersByName(node.Parameters.Text)
	case "wf_slc_users_onfurni":
		return nil, view.UsersOnFurniture(node.Targets)
	case "wf_slc_users_group":
		return nil, view.UsersInGroup()
	case "wf_slc_users_handitem":
		return nil, view.UsersByHanditem(valueAt(values, 0))
	case "wf_slc_users_bytype":
		return nil, view.UsersByType(valueAt(values, 0))
	case "wf_slc_users_neighborhood", "wf_slc_users_area":
		x1, y1, x2, y2 := area(values)
		return nil, view.UsersInArea(x1, y1, x2, y2)
	case "wf_slc_users_signal":
		if len(event.ActorIDs) == 0 && event.ActorID != 0 {
			return nil, []int64{event.ActorID}
		}
		return nil, event.ActorIDs
	case "wf_slc_users_team":
		return nil, view.UsersByTeam(valueAt(values, 0))
	case "wf_slc_furni_bytype":
		return view.FurniByType(node.Targets), nil
	case "wf_slc_furni_altitude":
		return view.FurniByAltitude(valueAt(values, 0), valueAt(values, 1)), nil
	case "wf_slc_furni_onfurni":
		return view.FurniOnFurniture(node.Targets), nil
	case "wf_slc_furni_picks":
		return node.Targets, nil
	case "wf_slc_furni_signal":
		if len(event.FurniTargets) == 0 && event.SourceItem > 0 {
			return []record.Target{{
				ItemID: event.SourceItem, SpriteID: event.SourceSprite,
			}}, nil
		}
		return event.FurniTargets, nil
	case "wf_slc_furni_neighborhood", "wf_slc_furni_area":
		x1, y1, x2, y2 := area(values)
		return view.FurniInArea(x1, y1, x2, y2), nil
	case "wf_slc_furni_with_var":
		return view.FurniByVariable(node.Parameters.Text), nil
	case "wf_slc_users_with_var":
		return nil, view.UsersByVariable(node.Parameters.Text)
	case "wf_slc_remote":
		return resolver.resolveRemote(ctx, node, event, generation, view, depth)
	default:
		return nil, nil
	}
}

// resolveRemote evaluates stacks referenced by the selector's selected boxes.
func (resolver *Resolver) resolveRemote(ctx context.Context, node *configuration.Node, event trigger.Event, generation *configuration.Generation, view View, depth int) ([]record.Target, []int64) {
	merged := Selection{}
	var furniSeen idSet
	var actorSeen idSet
	for _, target := range node.Targets {
		remote := generation.Nodes[target.ItemID]
		if remote == nil {
			continue
		}
		resolved, err := resolver.resolve(ctx, generation.Stacks[remote.Point], event, generation, view, depth+1)
		if err != nil {
			return nil, nil
		}
		appendFurni(&merged, &furniSeen, resolved.Furni, resolver.limit)
		appendActors(&merged, &actorSeen, resolved.Actors, resolver.limit)
		merged.FurniResolved = merged.FurniResolved || resolved.FurniResolved
		merged.ActorsResolved = merged.ActorsResolved || resolved.ActorsResolved
	}
	sortSelection(&merged)
	return merged.Furni, merged.Actors
}

// sortSelection establishes deterministic target order after all unions.
func sortSelection(selected *Selection) {
	slices.SortFunc(selected.Furni, func(left record.Target, right record.Target) int {
		return compareID(left.ItemID, right.ItemID)
	})
	slices.Sort(selected.Actors)
}

// compareID compares identifiers without subtraction overflow.
func compareID(left int64, right int64) int {
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}
