package routes

import (
	"context"
	"strings"
	"time"

	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/pkg/http/adminaction"
)

// tradeLockSnapshot resolves the active administrative trade-lock state.
func (dependencies Dependencies) tradeLockSnapshot(playerID int64) adminaction.Snapshot {
	return func(ctx context.Context) (any, error) {
		history, err := dependencies.Sanctions.History(ctx, playerID, 500)
		if err != nil {
			return nil, err
		}
		for _, punishment := range history {
			administrative := punishment.Source == "admin_http" ||
				strings.HasPrefix(punishment.Reason, "Migrated legacy")
			if punishment.Kind == sanctionrecord.KindTradeLock &&
				punishment.ActiveAt(time.Now()) && administrative {
				return map[string]any{
					"locked": true, "punishmentId": punishment.ID,
					"expiresAt": punishment.ExpiresAt,
				}, nil
			}
		}
		return map[string]any{"locked": false}, nil
	}
}
