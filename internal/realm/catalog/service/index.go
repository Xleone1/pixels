package service

// IndexReader exposes offer identifiers needed by the catalog index packet.
type IndexReader interface {
	// PageOfferIDs returns direct offers visible on one catalog page.
	PageOfferIDs(pageID int64, hasClub bool) []int64
}

// PageOfferIDs returns direct offers visible on one catalog page.
func (service *Service) PageOfferIDs(pageID int64, hasClub bool) []int64 {
	items := service.cache.pageItems(pageID)
	ids := make([]int64, 0, len(items))
	for _, item := range items {
		if item.Enabled && (!item.ClubOnly || hasClub) {
			ids = append(ids, item.ID)
		}
	}

	return ids
}
