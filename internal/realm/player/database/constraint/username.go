// Package constraint classifies PostgreSQL player persistence constraints.
package constraint

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const activeUsernameUnique = "players_username_active_uidx"

// IsActiveUsernameConflict reports whether err violates the active username constraint.
func IsActiveUsernameConflict(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		postgresError.ConstraintName == activeUsernameUnique
}
