package buy

import (
	"context"
	"errors"
	"strconv"

	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
	netconn "github.com/niflaot/pixels/networking/connection"
	outsoldout "github.com/niflaot/pixels/networking/outbound/catalog/limited/soldout"
	"github.com/niflaot/pixels/networking/outbound/catalog/offer"
	outfailed "github.com/niflaot/pixels/networking/outbound/catalog/purchase/failed"
	outok "github.com/niflaot/pixels/networking/outbound/catalog/purchase/ok"
	outunavailable "github.com/niflaot/pixels/networking/outbound/catalog/purchase/unavailable"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outbadgeadd "github.com/niflaot/pixels/networking/outbound/progression/achievement/badgeadd"
	"github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// sendPurchase projects one committed purchase through its native inventory packets.
func (handler Handler) sendPurchase(ctx context.Context, connection netconn.Context, result catalogservice.PurchaseResult, mapped offer.Offer) error {
	if result.CreatedRoomID != nil {
		if err := sendPurchaseOK(ctx, connection, mapped); err != nil {
			return err
		}
		message := "Room bundle purchased."
		if handler.Translations != nil {
			message = handler.Translations.Default("catalog.room_bundle.purchased", i18n.Params{"room": result.CreatedRoomName})
		}
		packet, err := bubblealert.Encode("catalog.room_bundle.purchased", message, bubblealert.WithDisplayBubble(), bubblealert.WithParam("roomId", strconv.FormatInt(*result.CreatedRoomID, 10)))
		if err != nil {
			return err
		}
		return connection.Send(ctx, packet)
	}
	if result.GrantedPet != nil {
		return sendPurchaseOK(ctx, connection, mapped)
	}
	if result.GrantedBadge != nil {
		packet, err := outunseen.EncodeBadges([]int64{result.GrantedBadge.ID})
		if err != nil {
			return err
		}
		if err = connection.Send(ctx, packet); err != nil {
			return err
		}
		packet, err = outbadgeadd.Encode(int32(result.GrantedBadge.ID), result.GrantedBadge.Code)
		if err != nil {
			return err
		}
		if err = connection.Send(ctx, packet); err != nil {
			return err
		}
		return sendPurchaseOK(ctx, connection, mapped)
	}
	itemIDs := make([]int64, 0, len(result.GrantedItems))
	for _, item := range result.GrantedItems {
		itemIDs = append(itemIDs, item.ID)
	}
	packet, err := outunseen.EncodeOwned(itemIDs)
	if err != nil {
		return err
	}
	if err = connection.Send(ctx, packet); err != nil {
		return err
	}
	if err = sendPurchaseOK(ctx, connection, mapped); err != nil {
		return err
	}
	refresh, err := outrefresh.Encode()
	if err != nil {
		return err
	}
	return connection.Send(ctx, refresh)
}

// sendPurchaseOK sends one successful catalog purchase result.
func sendPurchaseOK(ctx context.Context, connection netconn.Context, mapped offer.Offer) error {
	packet, err := outok.Encode(mapped)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// sendError maps a catalog service failure to its protocol result.
func (handler Handler) sendError(ctx context.Context, connection netconn.Context, offerID int64, err error) error {
	if errors.Is(err, catalogservice.ErrLimitedSoldOut) {
		packet, encodeErr := outsoldout.Encode()
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, packet)
	}
	if errors.Is(err, roombundle.ErrRoomLimitReached) {
		packet, encodeErr := outfailed.Encode(outfailed.CodeRoomLimit)
		if encodeErr != nil {
			return encodeErr
		}
		if sendErr := connection.Send(ctx, packet); sendErr != nil {
			return sendErr
		}
		message := "You have reached the room limit."
		if handler.Translations != nil {
			message = handler.Translations.Default("catalog.room_bundle.error.room_limit")
		}
		packet, encodeErr = bubblealert.Encode("catalog.room_bundle.error.room_limit", message, bubblealert.WithDisplayBubble())
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, packet)
	}
	if errors.Is(err, catalogservice.ErrOfferNotFound) || errors.Is(err, catalogservice.ErrOfferNotVisible) ||
		errors.Is(err, catalogservice.ErrOfferDisabled) || errors.Is(err, catalogservice.ErrPageNotFound) ||
		errors.Is(err, catalogservice.ErrInvalidAmount) || errors.Is(err, currencyservice.ErrInsufficientBalance) {
		return handler.sendUnavailable(ctx, connection)
	}
	if handler.Log != nil {
		handler.Log.Error("catalog purchase failed", zap.Int64("offer_id", offerID), zap.Error(err))
	}
	packet, encodeErr := outfailed.Encode(outfailed.CodeServer)
	if encodeErr != nil {
		return encodeErr
	}
	return connection.Send(ctx, packet)
}

// sendUnavailable sends an illegal purchase response.
func (handler Handler) sendUnavailable(ctx context.Context, connection netconn.Context) error {
	packet, err := outunavailable.Encode(outunavailable.CodeIllegal)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}
