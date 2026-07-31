package selection

import (
	"context"
	"strings"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// selectorView returns one deterministic target from every live-room query.
type selectorView struct{}

// UsersByAction returns one actor.
func (selectorView) UsersByAction(int32) []int64 { return []int64{10} }

// UsersByName returns one actor.
func (selectorView) UsersByName(string) []int64 { return []int64{10} }

// UsersOnFurniture returns one actor.
func (selectorView) UsersOnFurniture([]record.Target) []int64 { return []int64{10} }

// UsersInGroup returns one actor.
func (selectorView) UsersInGroup() []int64 { return []int64{10} }

// UsersByHanditem returns one actor.
func (selectorView) UsersByHanditem(int32) []int64 { return []int64{10} }

// UsersByType returns one actor.
func (selectorView) UsersByType(int32) []int64 { return []int64{10} }

// UsersInArea returns one actor.
func (selectorView) UsersInArea(int, int, int, int) []int64 { return []int64{10} }

// UsersByTeam returns one actor.
func (selectorView) UsersByTeam(int32) []int64 { return []int64{10} }

// FurniByType returns one furniture target.
func (selectorView) FurniByType([]record.Target) []record.Target {
	return []record.Target{{ItemID: 20}}
}

// FurniByAltitude returns one furniture target.
func (selectorView) FurniByAltitude(int32, int32) []record.Target {
	return []record.Target{{ItemID: 20}}
}

// FurniOnFurniture returns one furniture target.
func (selectorView) FurniOnFurniture([]record.Target) []record.Target {
	return []record.Target{{ItemID: 20}}
}

// FurniInArea returns one furniture target.
func (selectorView) FurniInArea(int, int, int, int) []record.Target {
	return []record.Target{{ItemID: 20}}
}

// FurniByVariable returns one furniture target.
func (selectorView) FurniByVariable(string) []record.Target {
	return []record.Target{{ItemID: 20}}
}

// UsersByVariable returns one actor.
func (selectorView) UsersByVariable(string) []int64 { return []int64{10} }

// TestResolverCoversEverySelector verifies every official selector route.
func TestResolverCoversEverySelector(t *testing.T) {
	keys := []string{
		"wf_slc_furni_area", "wf_slc_furni_neighborhood", "wf_slc_furni_bytype",
		"wf_slc_users_area", "wf_slc_users_neighborhood", "wf_slc_furni_altitude",
		"wf_slc_furni_onfurni", "wf_slc_furni_picks", "wf_slc_furni_signal",
		"wf_slc_users_signal", "wf_slc_users_bytype", "wf_slc_users_team",
		"wf_slc_users_byaction", "wf_slc_users_byname", "wf_slc_users_onfurni",
		"wf_slc_users_group", "wf_slc_users_handitem", "wf_slc_furni_with_var",
		"wf_slc_users_with_var",
	}
	for _, key := range keys {
		t.Run(key, func(t *testing.T) {
			node := selectorNode(1, key)
			stack := &configuration.Stack{Selectors: []*configuration.Node{node}}
			event := trigger.Event{
				Kind: trigger.ReceiveSignal, ActorID: 10, ActorIDs: []int64{10},
				FurniTargets: []record.Target{{ItemID: 20}},
			}
			result, err := New(10).ResolveStack(
				context.Background(), stack, event,
				&configuration.Generation{}, selectorView{},
			)
			if err != nil {
				t.Fatal(err)
			}
			if strings.HasPrefix(key, "wf_slc_users_") {
				if len(result.Actors) != 1 || result.Actors[0] != 10 {
					t.Fatalf("actor selection=%+v", result)
				}
			} else if len(result.Furni) != 1 {
				t.Fatalf("furniture selection=%+v", result)
			}
		})
	}
}

// TestRemoteSelectorResolvesReferencedStack verifies bounded cross-stack selection.
func TestRemoteSelectorResolvesReferencedStack(t *testing.T) {
	remote := selectorNode(1, "wf_slc_remote")
	remote.Targets = []record.Target{{ItemID: 2}}
	target := selectorNode(2, "wf_slc_users_group")
	target.Point = configuration.Point{X: 2}
	generation := &configuration.Generation{
		Nodes: map[int64]*configuration.Node{2: target},
		Stacks: map[configuration.Point]*configuration.Stack{
			target.Point: {Point: target.Point, Selectors: []*configuration.Node{target}},
		},
	}
	result, err := New(10).ResolveStack(
		context.Background(),
		&configuration.Stack{Selectors: []*configuration.Node{remote}},
		trigger.Event{},
		generation,
		selectorView{},
	)
	if err != nil || len(result.Actors) != 1 || result.Actors[0] != 10 {
		t.Fatalf("remote selection=%+v error=%v", result, err)
	}
}

// TestIDSetPromotesWithoutLosingEntries verifies exceptional selector deduplication.
func TestIDSetPromotesWithoutLosingEntries(t *testing.T) {
	var seen idSet
	for identifier := int64(1); identifier <= 12; identifier++ {
		if !seen.add(identifier) {
			t.Fatalf("identifier %d was rejected before insertion", identifier)
		}
	}
	for identifier := int64(1); identifier <= 12; identifier++ {
		if seen.add(identifier) {
			t.Fatalf("identifier %d was duplicated after promotion", identifier)
		}
	}
}

// TestActorSelectionRetainsNonPlayerEntityKeys verifies bot and pet targets.
func TestActorSelectionRetainsNonPlayerEntityKeys(t *testing.T) {
	selected := Selection{}
	var seen idSet
	appendActors(&selected, &seen, []int64{-8, 0, 7}, 10)
	sortSelection(&selected)
	if len(selected.Actors) != 2 ||
		selected.Actors[0] != -8 || selected.Actors[1] != 7 {
		t.Fatalf("actors=%v", selected.Actors)
	}
}

// TestSignalSelectorsFallBackToSourceContext verifies non-selector senders.
func TestSignalSelectorsFallBackToSourceContext(t *testing.T) {
	resolver := New(10)
	event := trigger.Event{
		ActorID: 7, SourceItem: 8, SourceSprite: 9,
	}
	user, err := resolver.ResolveStack(
		context.Background(),
		&configuration.Stack{
			Selectors: []*configuration.Node{
				selectorNode(1, "wf_slc_users_signal"),
			},
		},
		event,
		&configuration.Generation{},
		selectorView{},
	)
	if err != nil || len(user.Actors) != 1 || user.Actors[0] != 7 {
		t.Fatalf("user signal=%+v err=%v", user, err)
	}
	furni, err := resolver.ResolveStack(
		context.Background(),
		&configuration.Stack{
			Selectors: []*configuration.Node{
				selectorNode(2, "wf_slc_furni_signal"),
			},
		},
		event,
		&configuration.Generation{},
		selectorView{},
	)
	if err != nil || len(furni.Furni) != 1 ||
		furni.Furni[0].ItemID != 8 || furni.Furni[0].SpriteID != 9 {
		t.Fatalf("furniture signal=%+v err=%v", furni, err)
	}
}

// selectorNode creates one fully parameterized selector node.
func selectorNode(itemID int64, key string) *configuration.Node {
	return &configuration.Node{
		ItemID:     itemID,
		Descriptor: registry.Descriptor{Key: key, Family: registry.FamilySelector},
		Parameters: configuration.Parameters{
			Values: []int32{1, 1, 4, 4},
			Text:   "score",
		},
		Targets: []record.Target{{ItemID: 9}},
	}
}

// BenchmarkResolveArea measures bounded selector CPU and heap cost.
func BenchmarkResolveArea(b *testing.B) {
	stack := &configuration.Stack{
		Selectors: []*configuration.Node{selectorNode(1, "wf_slc_users_area")},
	}
	resolver := New(1000)
	b.ReportAllocs()
	for range b.N {
		result, err := resolver.ResolveStack(
			context.Background(), stack, trigger.Event{},
			&configuration.Generation{}, selectorView{},
		)
		if err != nil || len(result.Actors) != 1 {
			b.Fatal("selector failed")
		}
	}
}
