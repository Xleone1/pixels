package identity

import (
	"context"
	"time"
)

// Store commits atomic identity replacements and denormalized room owners.
type Store interface {
	// Rename commits one cooldown-governed username replacement and audit row.
	Rename(context.Context, int64, string, time.Time, time.Duration) (RenameResult, error)
	// RenameAdmin commits one attributed administrative username replacement.
	RenameAdmin(context.Context, int64, string, int64, string, time.Time, time.Duration) (RenameResult, error)
	// NameChanges returns recent username audit entries for one player.
	NameChanges(context.Context, int64, int) ([]NameChange, error)
}
