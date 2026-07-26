package buy

import (
	"context"
	"testing"

	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/catalog/offer"
	outok "github.com/niflaot/pixels/networking/outbound/catalog/purchase/ok"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outbadgeadd "github.com/niflaot/pixels/networking/outbound/progression/achievement/badgeadd"
)

// TestSendPurchaseMarksBadgeUnseenBeforeDelivery verifies Nitro can show the badge counter immediately.
func TestSendPurchaseMarksBadgeUnseenBeforeDelivery(t *testing.T) {
	connection, sent := buyConnection(t)
	result := catalogservice.PurchaseResult{GrantedBadge: &catalogservice.BadgeReward{ID: 24, Code: "QA_BADGE"}}
	if err := (Handler{}).sendPurchase(context.Background(), connection, result, offer.Offer{}); err != nil {
		t.Fatalf("send purchase: %v", err)
	}
	if len(*sent) != 3 || (*sent)[0].Header != outunseen.Header || (*sent)[1].Header != outbadgeadd.Header || (*sent)[2].Header != outok.Header {
		t.Fatalf("unexpected packets %#v", *sent)
	}
	values, err := codec.DecodePacketExact((*sent)[0], codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
	})
	if err != nil || values[1].Int32 != 4 || values[3].Int32 != 24 {
		t.Fatalf("unexpected unseen payload %#v err=%v", values, err)
	}
}
