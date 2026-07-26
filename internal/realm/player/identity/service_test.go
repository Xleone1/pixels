package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	redispkg "github.com/niflaot/pixels/pkg/redis"
)

// identityFinder stores exact usernames and rename policy for focused tests.
type identityFinder struct {
	// cooldown reports whether the actor recently changed names.
	cooldown bool
	// taken stores case-insensitive existing names.
	taken map[string]bool
	// playerID identifies the returned player.
	playerID int64
	// username stores the current visible name.
	username string
	// takenID identifies a colliding player.
	takenID int64
}

// TestCheckReusesTheSamePlayersReservation verifies repeated browser checks.
func TestCheckReusesTheSamePlayersReservation(t *testing.T) {
	server := miniredis.RunT(t)
	client := redispkg.New(redispkg.Config{Address: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	service := New(
		&renameStore{},
		identityFinder{taken: map[string]bool{}},
		client,
	)
	for attempt := 0; attempt < 2; attempt++ {
		result, err := service.Check(context.Background(), 1, "Browser")
		if err != nil || result.Code != ResultAvailable {
			t.Fatalf("attempt=%d result=%#v err=%v", attempt, result, err)
		}
	}
}

// FindByID returns the actor policy.
func (finder identityFinder) FindByID(_ context.Context, playerID int64) (playerservice.Record, bool, error) {
	if finder.playerID == 0 {
		finder.playerID = playerID
	}
	var lastChangedAt *time.Time
	if finder.cooldown {
		value := time.Now().UTC()
		lastChangedAt = &value
	}
	return playerservice.Record{Player: playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: finder.playerID}}, Username: finder.username}, Profile: playermodel.Profile{LastNameChangeAt: lastChangedAt}}, true, nil
}

// FindByUsername reports configured collisions.
func (finder identityFinder) FindByUsername(_ context.Context, username string) (playerservice.Record, bool, error) {
	return playerservice.Record{Player: playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: finder.takenID}}}}, finder.taken[username], nil
}

// renameStore records committed candidates.
type renameStore struct {
	// candidate stores the committed username.
	candidate string
	// actorPlayerID stores the administrative actor.
	actorPlayerID int64
	// reason stores the administrative reason.
	reason string
	// historyLimit stores the requested audit limit.
	historyLimit int
}

// identityFilter reports a deterministic protected-word match.
type identityFilter struct{}

// Censor rejects candidates containing the test protected word.
func (identityFilter) Censor(value string) (string, bool) { return value, value == "blocked" }

// Rename stores one candidate.
func (store *renameStore) Rename(_ context.Context, _ int64, candidate string, changedAt time.Time, cooldown time.Duration) (RenameResult, error) {
	store.candidate = candidate
	return RenameResult{OldUsername: "old", NewUsername: candidate, ChangedAt: changedAt, AvailableAt: changedAt.Add(cooldown)}, nil
}

// RenameAdmin stores one attributed candidate.
func (store *renameStore) RenameAdmin(_ context.Context, _ int64, candidate string, actorPlayerID int64, reason string, changedAt time.Time, cooldown time.Duration) (RenameResult, error) {
	store.candidate = candidate
	store.actorPlayerID = actorPlayerID
	store.reason = reason
	return RenameResult{OldUsername: "old", NewUsername: candidate, ChangedAt: changedAt, AvailableAt: changedAt.Add(cooldown)}, nil
}

// NameChanges returns no configured history.
func (store *renameStore) NameChanges(_ context.Context, _ int64, limit int) ([]NameChange, error) {
	store.historyLimit = limit
	return []NameChange{{ID: 1}}, nil
}

// TestCheckRejectsReservedAndFilteredNames verifies server-owned name policy.
func TestCheckRejectsReservedAndFilteredNames(t *testing.T) {
	service := NewConfigured(&renameStore{}, identityFinder{taken: map[string]bool{}}, nil, identityFilter{}, DefaultConfig())
	for _, candidate := range []string{"admin", "blocked"} {
		result, err := service.Check(context.Background(), 1, candidate)
		if err != nil || result.Code != ResultInvalid {
			t.Fatalf("candidate=%q result=%#v err=%v", candidate, result, err)
		}
	}
}

// TestValidateUsernameCodes verifies Nitro's complete validation matrix.
func TestValidateUsernameCodes(t *testing.T) {
	service := New(&renameStore{}, identityFinder{}, nil)
	cases := map[string]int32{"ab": ResultTooShort, "abcdefghijklmnop": ResultTooLong, "bad name": ResultInvalid, "Valid_1": ResultAvailable}
	for value, expected := range cases {
		if actual := service.validate(value); actual != expected {
			t.Fatalf("validate %q=%d expected %d", value, actual, expected)
		}
	}
}

// TestCheckHonorsPolicyAndProducesDeterministicSuggestions verifies availability behavior.
func TestCheckHonorsPolicyAndProducesDeterministicSuggestions(t *testing.T) {
	service := New(&renameStore{}, identityFinder{cooldown: true, taken: map[string]bool{}}, nil)
	result, err := service.Check(context.Background(), 1, "Valid")
	if err != nil || result.Code != ResultCooldown || result.AvailableAt == nil {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	service = New(&renameStore{}, identityFinder{taken: map[string]bool{"Valid": true}}, nil)
	result, err = service.Check(context.Background(), 1, "Valid")
	if err != nil || result.Code != ResultTaken || len(result.Suggestions) != 4 || result.Suggestions[0] != "Valid1" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

// TestRenameValidatesAndCommits verifies the cold-path rename boundary.
func TestRenameValidatesAndCommits(t *testing.T) {
	store := &renameStore{}
	service := New(store, identityFinder{taken: map[string]bool{}}, nil)
	if _, err := service.Rename(context.Background(), 1, "bad name"); !errors.Is(err, ErrReservationMissing) {
		t.Fatalf("expected invalid rename, got %v", err)
	}
	if _, err := service.Rename(context.Background(), 1, "admin"); !errors.Is(err, ErrReservationMissing) {
		t.Fatalf("expected reserved rename rejection, got %v", err)
	}
	result, err := service.Rename(context.Background(), 1, "Valid")
	if err != nil || result.NewUsername != "Valid" || store.candidate != "Valid" {
		t.Fatalf("result=%#v candidate=%q err=%v", result, store.candidate, err)
	}
	if reservationKey("VaLiD") != "player:name:reservation:valid" {
		t.Fatal("unexpected reservation key")
	}
}

// TestRenameRejectsKeepingCurrentName verifies no-op changes do not consume cooldown.
func TestRenameRejectsKeepingCurrentName(t *testing.T) {
	store := &renameStore{}
	finder := identityFinder{username: "Current", taken: map[string]bool{"Current": true}, takenID: 1}
	service := New(store, finder, nil)
	check, err := service.Check(context.Background(), 1, "Current")
	if err != nil || check.Code != ResultAvailable {
		t.Fatalf("check=%#v err=%v", check, err)
	}
	if _, err = service.Rename(context.Background(), 1, "Current"); !errors.Is(err, ErrUsernameTaken) {
		t.Fatalf("expected current-name rejection, got %v", err)
	}
}

// TestAdministrativeRenameAndHistory verifies attributed mutation and bounded history behavior.
func TestAdministrativeRenameAndHistory(t *testing.T) {
	store := &renameStore{}
	service := New(store, identityFinder{playerID: 1, taken: map[string]bool{}}, nil)
	result, err := service.RenameAdmin(context.Background(), 1, "Renamed", 9, "support request")
	if err != nil || result.NewUsername != "Renamed" || store.actorPlayerID != 9 || store.reason != "support request" {
		t.Fatalf("result=%#v store=%#v err=%v", result, store, err)
	}
	changes, err := service.NameChanges(context.Background(), 1, 500)
	if err != nil || len(changes) != 1 || store.historyLimit != 200 {
		t.Fatalf("changes=%#v limit=%d err=%v", changes, store.historyLimit, err)
	}
	if _, err = service.RenameAdmin(context.Background(), 1, "Renamed", 0, ""); !errors.Is(err, ErrInvalidAttribution) {
		t.Fatalf("expected invalid attribution, got %v", err)
	}
}

// TestStatusUsesConfiguredCooldown verifies exact automatic eligibility.
func TestStatusUsesConfiguredCooldown(t *testing.T) {
	changedAt := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	finder := identityFinder{taken: map[string]bool{}}
	service := NewConfigured(&renameStore{}, finder, nil, nil, Config{
		MinimumLength: 3, MaximumLength: 15, AllowedSymbols: "_",
		ReservationTTL: time.Minute, ChangeCooldownDays: 30,
	})
	service.now = func() time.Time { return changedAt.Add(29 * 24 * time.Hour) }
	status := service.status(&changedAt, service.now())
	if status.Available || status.AvailableAt == nil || status.RemainingSeconds != 24*60*60 {
		t.Fatalf("status=%#v", status)
	}
	service.now = func() time.Time { return changedAt.Add(30 * 24 * time.Hour) }
	if status = service.status(&changedAt, service.now()); !status.Available || status.AvailableAt != nil {
		t.Fatalf("status=%#v", status)
	}
}

// TestConfiguredPolicyFallsBackAndLoadsEnvironment verifies config boundaries.
func TestConfiguredPolicyFallsBackAndLoadsEnvironment(t *testing.T) {
	service := NewConfigured(&renameStore{}, identityFinder{}, nil, nil, Config{})
	if service.config.MinimumLength != DefaultConfig().MinimumLength {
		t.Fatalf("config=%#v", service.config)
	}
	t.Setenv("PIXELS_PLAYER_USERNAME_MIN_LENGTH", "4")
	t.Setenv("PIXELS_PLAYER_USERNAME_CHANGE_COOLDOWN_DAYS", "45")
	config, err := LoadConfig()
	if err != nil || config.MinimumLength != 4 || config.ChangeCooldownDays != 45 {
		t.Fatalf("config=%#v err=%v", config, err)
	}
}
