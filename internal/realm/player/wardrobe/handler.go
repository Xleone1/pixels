package wardrobe

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inredeem "github.com/niflaot/pixels/networking/inbound/user/clothing/redeem"
	inget "github.com/niflaot/pixels/networking/inbound/user/wardrobe/get"
	insave "github.com/niflaot/pixels/networking/inbound/user/wardrobe/save"
	outremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outroomremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outclothing "github.com/niflaot/pixels/networking/outbound/user/clothing/sets"
	outoutfits "github.com/niflaot/pixels/networking/outbound/user/wardrobe/outfits"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler adapts Nitro wardrobe packets to durable behavior.
type Handler struct {
	// Service validates and persists outfits.
	Service *Service
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
	// Rooms stores active room furniture projections.
	Rooms *roomlive.Registry
	// Connections stores active room audiences.
	Connections *netconn.Registry
	// Translations resolves expected validation feedback.
	Translations i18n.Translator
}

// RegisterHandlers installs wardrobe packet adapters.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(insave.Header, handler.save)
	_ = registry.Register(inget.Header, handler.get)
	_ = registry.Register(inredeem.Header, handler.redeem)
}

// redeem consumes one eligible clothing furniture and projects current unlocks.
func (handler Handler) redeem(connection netconn.Context, packet codec.Packet) error {
	payload, err := inredeem.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Service == nil {
		return nil
	}
	result, err := handler.Service.Redeem(context.Background(), playerID, int64(payload.ItemID))
	if err != nil {
		if errors.Is(err, ErrInvalidClothingItem) {
			return handler.alert(connection, "user.clothing.invalid_item")
		}
		return err
	}
	if result.Applied {
		if result.RoomID > 0 {
			if projectionErr := handler.removePlaced(context.Background(), result.RoomID, int64(payload.ItemID), playerID); projectionErr != nil {
				return projectionErr
			}
		} else {
			removed, encodeErr := outremove.Encode(int64(payload.ItemID))
			if encodeErr != nil {
				return encodeErr
			}
			if sendErr := connection.Send(context.Background(), removed); sendErr != nil {
				return sendErr
			}
		}
	}
	response, err := outclothing.Encode(result.Snapshot.FigureSetIDs, result.Snapshot.ProductCodes)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// removePlaced removes one redeemed clothing item from its active room.
func (handler Handler) removePlaced(ctx context.Context, roomID int64, itemID int64, ownerPlayerID int64) error {
	if handler.Rooms == nil {
		return nil
	}
	active, found := handler.Rooms.Find(roomID)
	if !found {
		return nil
	}
	if _, err := active.ReloadFurniture(itemID, nil); err != nil {
		return err
	}
	if handler.Connections == nil {
		return nil
	}
	packet, err := outroomremove.Encode(itemID, ownerPlayerID)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
}

// save persists one wardrobe slot without inventing a success response.
func (handler Handler) save(connection netconn.Context, packet codec.Packet) error {
	payload, err := insave.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Service == nil {
		return nil
	}
	err = handler.Service.Save(context.Background(), playerID, Outfit{SlotID: payload.SlotID, Figure: payload.Figure, Gender: payload.Gender})
	if errors.Is(err, ErrInvalidOutfit) {
		return handler.alert(connection, "user.wardrobe.slot_invalid")
	}
	return err
}

// alert sends localized expected wardrobe feedback without disconnecting the session.
func (handler Handler) alert(connection netconn.Context, key i18n.Key) error {
	message := string(key)
	if handler.Translations != nil {
		message = handler.Translations.Default(key)
	}
	packet, err := outalert.Encode(message)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), packet)
}

// get sends one bounded wardrobe page.
func (handler Handler) get(connection netconn.Context, packet codec.Packet) error {
	payload, err := inget.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Service == nil {
		return nil
	}
	outfits, err := handler.Service.Outfits(context.Background(), playerID)
	if err != nil {
		return err
	}
	slots := make([]int32, len(outfits))
	figures := make([]string, len(outfits))
	genders := make([]string, len(outfits))
	for index, outfit := range outfits {
		slots[index], figures[index], genders[index] = outfit.SlotID, outfit.Figure, outfit.Gender
	}
	response, err := outoutfits.Encode(payload.PageID, slots, figures, genders)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// playerID resolves one authenticated connection without allocations.
func (handler Handler) playerID(connection netconn.Context) (int64, bool) {
	if handler.Bindings == nil {
		return 0, false
	}
	current, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	return current.PlayerID, found
}
