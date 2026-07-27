package openapi

import "time"

// SubscriptionMembershipResponse documents durable membership state.
type SubscriptionMembershipResponse struct {
	// PlayerID identifies the member.
	PlayerID int64 `json:"playerId"`
	// Level stores the HC or VIP tier.
	Level int16 `json:"level"`
	// ExpiresAt stores the exclusive entitlement expiration.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// LifetimeActiveSeconds stores accumulated active club time.
	LifetimeActiveSeconds int64 `json:"lifetimeActiveSeconds"`
	// LifetimeVIPSeconds stores accumulated VIP time.
	LifetimeVIPSeconds int64 `json:"lifetimeVipSeconds"`
	// GiftsEarned stores materialized monthly gifts.
	GiftsEarned int32 `json:"giftsEarned"`
	// GiftsClaimed stores claimed monthly gifts.
	GiftsClaimed int32 `json:"giftsClaimed"`
	// Version stores the optimistic membership version.
	Version int64 `json:"version"`
}
