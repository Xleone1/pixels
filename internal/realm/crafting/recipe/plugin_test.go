package recipe

import (
	"context"
	"errors"
	"testing"

	craftingconfig "github.com/niflaot/pixels/internal/realm/crafting/config"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// craftingStore executes fixture transactions directly.
type craftingStore struct{ craftingrecord.Store }

// WithinTransaction executes one fixture transaction directly.
func (*craftingStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// craftingFurniture supplies and consumes one inventory ingredient.
type craftingFurniture struct {
	furnitureservice.TradingManager
	// deleted stores consumed item identifiers.
	deleted []int64
}

// ListInventory returns one deterministic ingredient.
func (*craftingFurniture) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return []furnituremodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 4}}, DefinitionID: 10, OwnerPlayerID: 7}}, nil
}

// DeleteInventoryItem records one consumed ingredient.
func (furniture *craftingFurniture) DeleteInventoryItem(_ context.Context, itemID int64, _ int64) error {
	furniture.deleted = append(furniture.deleted, itemID)
	return nil
}

// FindDefinitionByID returns the requested reward definition.
func (*craftingFurniture) FindDefinitionByID(_ context.Context, definitionID int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: definitionID}}}, true, nil
}

// craftingGranter records the granted reward definition.
type craftingGranter struct {
	// definitionID stores the granted reward definition.
	definitionID int64
}

// Grant records and returns one reward.
func (granter *craftingGranter) Grant(_ context.Context, params furnitureservice.GrantParams) ([]furnituremodel.Item, error) {
	granter.definitionID = params.DefinitionID
	return []furnituremodel.Item{{DefinitionID: params.DefinitionID}}, nil
}

// craftingEvents returns a deterministic reward and veto.
type craftingEvents struct {
	// reward stores the replacement definition.
	reward int64
	// cancelled vetoes crafting.
	cancelled bool
}

// DispatchCraftingCraft returns the configured crafting interception.
func (events craftingEvents) DispatchCraftingCraft(context.Context, int64, int64, int64) (int64, bool) {
	return events.reward, events.cancelled
}

// TestPluginCraftingMutationAndCancellation verifies the transactional crafting gate.
func TestPluginCraftingMutationAndCancellation(t *testing.T) {
	furniture := &craftingFurniture{}
	granter := &craftingGranter{}
	service := New(craftingconfig.Config{}, &craftingStore{}, furniture, granter, nil, nil)
	recipe := craftingrecord.Recipe{ID: 3, RewardDefinitionID: 20, Ingredients: []craftingrecord.Ingredient{{DefinitionID: 10, Amount: 1}}}
	service.SetPluginRuntime(craftingEvents{reward: 99})
	result, err := service.commit(context.Background(), 7, recipe, nil)
	if err != nil || granter.definitionID != 99 || result.Recipe.RewardDefinitionID != 99 {
		t.Fatalf("result=%#v grant=%d err=%v", result, granter.definitionID, err)
	}
	service.SetPluginRuntime(craftingEvents{reward: 99, cancelled: true})
	if _, err = service.commit(context.Background(), 7, recipe, nil); !errors.Is(err, ErrCancelledByPlugin) {
		t.Fatalf("expected plugin cancellation, got %v", err)
	}
	if len(furniture.deleted) != 1 {
		t.Fatalf("cancelled craft consumed ingredients: %v", furniture.deleted)
	}
}
