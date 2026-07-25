package identity

import "context"

// Store commits atomic identity replacements and denormalized room owners.
type Store interface {
	// Rename commits one one-shot username replacement and audit row.
	Rename(context.Context, int64, string) (RenameResult, error)
	// RenameAdmin commits one attributed administrative username replacement.
	RenameAdmin(context.Context, int64, string, int64, string) (RenameResult, error)
	// NameChanges returns recent username audit entries for one player.
	NameChanges(context.Context, int64, int) ([]NameChange, error)
	// SetAuthorization replaces and audits one self-service rename authorization.
	SetAuthorization(context.Context, int64, bool, int64, string) error
}
