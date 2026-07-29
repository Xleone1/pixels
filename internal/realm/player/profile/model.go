// Package profile owns public player presentation, tags, and user respect.
package profile

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/permission"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerwardrobe "github.com/niflaot/pixels/internal/realm/player/wardrobe"
)

const (
	// DefaultDailyRespectLimit stores the ordinary daily user-respect allowance.
	DefaultDailyRespectLimit = 3
	// DefaultPetRespectLimit stores the ordinary daily pet-respect allowance.
	DefaultPetRespectLimit = 3
	// MaxTags stores the public tag capacity.
	MaxTags = 5
	// MaxTagLength stores the maximum tag length in runes.
	MaxTagLength = 32
)

var (
	// RespectUnlimited bypasses the ordinary daily user-respect quota.
	RespectUnlimited = permission.RegisterNode("profile.respect.unlimited", "")
)

// figureAllowed resolves entitlement snapshots before immutable catalog validation.
func (service *Service) figureAllowed(ctx context.Context, playerID int64, figure string, gender playermodel.Gender) (bool, error) {
	club := playermodel.ClubLevelNone
	previous := ""
	if service.live != nil {
		if player, found := service.live.Find(playerID); found {
			snapshot := player.Snapshot()
			club = snapshot.ClubLevelAt(service.now())
			if snapshot.Gender == gender {
				previous = snapshot.Look
			}
		}
	}
	unlocked := playerwardrobe.ClothingSnapshot{}
	if service.unlocks != nil {
		var err error
		unlocked, err = service.unlocks.Clothing(ctx, playerID)
		if err != nil {
			return false, err
		}
	}
	if previous != "" {
		return service.figures.AllowedTransition(previous, figure, gender, club, unlocked.FigureSetIDs), nil
	}
	return service.figures.Allowed(figure, gender, club, unlocked.FigureSetIDs), nil
}

var (
	// ErrInvalidFigure reports a malformed avatar figure or gender.
	ErrInvalidFigure = errors.New("invalid player figure")
	// ErrInvalidMotto reports a motto outside the public profile limit.
	ErrInvalidMotto = errors.New("invalid player motto")
	// ErrInvalidTags reports an invalid public tag replacement.
	ErrInvalidTags = errors.New("invalid player tags")
	// ErrCancelledByPlugin reports a public profile mutation vetoed before persistence.
	ErrCancelledByPlugin = errors.New("player profile mutation cancelled by plugin")
	// ErrRespectNotAllowed reports an ineligible respect attempt.
	ErrRespectNotAllowed = errors.New("player respect not allowed")
	// ErrRespectExhausted reports a consumed daily respect allowance.
	ErrRespectExhausted = errors.New("player respect exhausted")
	// ErrRespectAlreadyGranted reports a duplicate actor-target grant for the hotel day.
	ErrRespectAlreadyGranted = errors.New("player respect already granted")
	// ErrRespectThrottled reports a repeated respect request inside the abuse window.
	ErrRespectThrottled = errors.New("player respect throttled")
)

// RespectState contains durable user and pet respect counters.
type RespectState struct {
	// Received stores total respect received by the player.
	Received int32
	// UserRemaining stores remaining daily user respect grants.
	UserRemaining int32
	// PetRemaining stores remaining daily pet respect grants.
	PetRemaining int32
}

// RespectResult contains one committed user-respect outcome.
type RespectResult struct {
	// Applied reports whether the grant was newly committed.
	Applied bool
	// Duplicate reports whether the actor already respected this target today.
	Duplicate bool
	// TotalReceived stores the target's new durable total.
	TotalReceived int32
	// Remaining stores the actor's remaining daily grants.
	Remaining int32
}
