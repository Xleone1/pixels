// Package catalog grants persistent bots purchased through the hotel catalog.
package catalog

import (
	"context"
	"strings"

	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botpolicy "github.com/niflaot/pixels/internal/realm/bot/policy"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
)

// Service coordinates transactional bot grants and post-commit projection.
type Service struct {
	// config stores inventory limits.
	config botpolicy.Config
	// store persists bot records.
	store botrecord.Store
	// runtime projects bot inventory updates.
	runtime *botcore.Service
}

// New creates catalog bot reward behavior.
func New(config botpolicy.Config, store botrecord.Store, runtime *botcore.Service) *Service {
	return &Service{config: config.Normalize(), store: store, runtime: runtime}
}

// GrantCatalog creates one bot inside the caller transaction.
func (service *Service) GrantCatalog(ctx context.Context, params catalogservice.BotGrantParams) (catalogservice.BotReward, error) {
	count, err := service.store.CountInventory(ctx, params.OwnerPlayerID)
	if err != nil {
		return catalogservice.BotReward{}, err
	}
	if count >= service.config.MaxInventory {
		return catalogservice.BotReward{}, botrecord.ErrInventoryLimit
	}
	template, valid := parseTemplate(params.ProductCode, params.ExtraData)
	if !valid {
		return catalogservice.BotReward{}, botrecord.ErrInvalidSkill
	}
	template.OwnerPlayerID = params.OwnerPlayerID
	created, err := service.store.Create(ctx, template)
	if err != nil {
		return catalogservice.BotReward{}, err
	}

	return catalogservice.BotReward{ID: created.ID, OwnerPlayerID: created.OwnerPlayerID}, nil
}

// ProjectCatalog sends one committed bot to its online owner's inventory.
func (service *Service) ProjectCatalog(ctx context.Context, reward catalogservice.BotReward) {
	if service.runtime == nil {
		return
	}
	bot, found, err := service.store.Find(ctx, reward.ID)
	if err == nil && found {
		service.runtime.SendInventoryAdd(ctx, reward.OwnerPlayerID, bot, true)
	}
}

// parseTemplate converts one trusted Hubbly bot template into a durable record.
func parseTemplate(productCode string, extraData string) (botrecord.Bot, bool) {
	values := make(map[string]string)
	for _, field := range strings.Split(extraData, ";") {
		key, value, found := strings.Cut(field, ":")
		if found {
			values[strings.ToLower(strings.TrimSpace(key))] = strings.TrimSpace(value)
		}
	}
	name := values["name"]
	figure := values["figure"]
	gender := strings.ToUpper(values["gender"])
	if len(name) == 0 || len(name) > 15 || figure == "" || gender != "M" && gender != "F" {
		return botrecord.Bot{}, false
	}
	behavior := botrecord.BehaviorGeneric
	canWalk := true
	switch strings.ToLower(strings.TrimSpace(productCode)) {
	case "bot_bartender":
		behavior = botrecord.BehaviorBartender
	case "rentable_bot_visitor_log":
		behavior = botrecord.BehaviorVisitorLog
		canWalk = false
	}

	return botrecord.Bot{
		BehaviorType:     behavior,
		Name:             name,
		Motto:            values["motto"],
		Figure:           figure,
		Gender:           gender,
		CanWalk:          canWalk,
		ChatDelaySeconds: 10,
	}, true
}
