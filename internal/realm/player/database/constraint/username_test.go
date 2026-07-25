package constraint

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// TestIsActiveUsernameConflict classifies only the active username constraint.
func TestIsActiveUsernameConflict(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		conflicted bool
	}{
		{name: "username", err: &pgconn.PgError{Code: "23505", ConstraintName: activeUsernameUnique}, conflicted: true},
		{name: "wrapped username", err: fmt.Errorf("insert player: %w", &pgconn.PgError{Code: "23505", ConstraintName: activeUsernameUnique}), conflicted: true},
		{name: "primary key", err: &pgconn.PgError{Code: "23505", ConstraintName: "players_pkey"}},
		{name: "different code", err: &pgconn.PgError{Code: "23503", ConstraintName: activeUsernameUnique}},
		{name: "generic error", err: errors.New("failed")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if conflicted := IsActiveUsernameConflict(test.err); conflicted != test.conflicted {
				t.Fatalf("expected conflicted=%t, got %t", test.conflicted, conflicted)
			}
		})
	}
}
