package catalog

import (
	"context"
	"errors"
	"testing"

	botpolicy "github.com/niflaot/pixels/internal/realm/bot/policy"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
)

// catalogStore stores catalog-granted bots in memory.
type catalogStore struct {
	botrecord.Store
	// count stores the configured inventory size.
	count int
	// created stores the last created bot.
	created botrecord.Bot
}

// CountInventory returns the configured inventory size.
func (store *catalogStore) CountInventory(context.Context, int64) (int, error) {
	return store.count, nil
}

// Create records and identifies one catalog bot.
func (store *catalogStore) Create(_ context.Context, bot botrecord.Bot) (botrecord.Bot, error) {
	bot.ID = 91
	store.created = bot
	return bot, nil
}

// TestGrantCatalogCreatesEverySupportedBotBehavior verifies trusted template conversion.
func TestGrantCatalogCreatesEverySupportedBotBehavior(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"bot_generic":              botrecord.BehaviorGeneric,
		"bot_bartender":            botrecord.BehaviorBartender,
		"rentable_bot_visitor_log": botrecord.BehaviorVisitorLog,
	}
	for code, expected := range cases {
		store := &catalogStore{}
		service := New(botpolicy.Config{MaxInventory: 25}, store, nil)
		reward, err := service.GrantCatalog(context.Background(), catalogservice.BotGrantParams{
			OwnerPlayerID: 7,
			ProductCode:   code,
			ExtraData:     "name:Robbie;motto:Ayudante;figure:hd-180-1.hr-100-61;gender:m",
		})
		if err != nil || reward.ID != 91 || store.created.BehaviorType != expected || store.created.Gender != "M" {
			t.Fatalf("code=%q reward=%#v bot=%#v error=%v", code, reward, store.created, err)
		}
	}
}

// TestGrantCatalogRejectsInvalidTemplatesAndInventoryOverflow verifies grant boundaries.
func TestGrantCatalogRejectsInvalidTemplatesAndInventoryOverflow(t *testing.T) {
	t.Parallel()
	store := &catalogStore{count: 25}
	service := New(botpolicy.Config{MaxInventory: 25}, store, nil)
	if _, err := service.GrantCatalog(context.Background(), catalogservice.BotGrantParams{OwnerPlayerID: 7}); !errors.Is(err, botrecord.ErrInventoryLimit) {
		t.Fatalf("inventory limit error=%v", err)
	}
	store.count = 0
	if _, err := service.GrantCatalog(context.Background(), catalogservice.BotGrantParams{OwnerPlayerID: 7, ExtraData: "name:;gender:x"}); !errors.Is(err, botrecord.ErrInvalidSkill) {
		t.Fatalf("invalid template error=%v", err)
	}
}
