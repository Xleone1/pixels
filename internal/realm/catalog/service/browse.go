package service

import (
	"context"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// Pages returns pages visible to one player capability set.
func (service *Service) Pages(ctx context.Context, playerID int64, hasClub bool) ([]catalogmodel.Page, error) {
	pages := service.cache.pages()
	visible := make([]catalogmodel.Page, 0, len(pages))
	for _, page := range pages {
		accessible, err := service.pageAccessible(ctx, page, playerID, hasClub)
		if err != nil {
			return nil, err
		}
		if accessible {
			visible = append(visible, page)
		}
	}

	return visible, nil
}

// Page returns one visible page and its enabled offers.
func (service *Service) Page(ctx context.Context, pageID int64, playerID int64, hasClub bool) (catalogmodel.Page, []catalogmodel.Item, error) {
	page, found := service.cache.page(pageID)
	if !found {
		return catalogmodel.Page{}, nil, ErrPageNotFound
	}
	accessible, err := service.pageAccessible(ctx, page, playerID, hasClub)
	if err != nil {
		return catalogmodel.Page{}, nil, err
	}
	if !accessible {
		return catalogmodel.Page{}, nil, ErrOfferNotVisible
	}

	items, err := service.pageItems(ctx, page, playerID)
	if err != nil {
		return catalogmodel.Page{}, nil, err
	}
	visible := make([]catalogmodel.Item, 0, len(items))
	for _, item := range items {
		if (item.Enabled || page.Layout == catalogmodel.SoldLimitedItemsLayout) && (!item.ClubOnly || hasClub) {
			visible = append(visible, item)
		}
	}

	return page, visible, nil
}

// pageItems resolves static and player-specific catalog page contents.
func (service *Service) pageItems(ctx context.Context, page catalogmodel.Page, playerID int64) ([]catalogmodel.Item, error) {
	switch page.Layout {
	case catalogmodel.RecentPurchasesLayout:
		if service.purchaseHistory == nil {
			return nil, nil
		}
		ids, err := service.purchaseHistory.ListRecentPurchaseItemIDs(ctx, playerID, 100)
		if err != nil {
			return nil, err
		}
		return service.cache.itemsByID(ids), nil
	case catalogmodel.SoldLimitedItemsLayout:
		return service.cache.soldLimitedItems(), nil
	default:
		return service.cache.pageItems(page.ID), nil
	}
}

// SanitizeList returns definitions without an enabled active offer.
func (service *Service) SanitizeList(ctx context.Context) ([]furnituremodel.Definition, error) {
	return service.store.SanitizeList(ctx)
}
