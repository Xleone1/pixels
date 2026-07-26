package identity

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config contains username policy and reservation settings.
type Config struct {
	// MinimumLength is the shortest accepted username in runes.
	MinimumLength int `env:"PIXELS_PLAYER_USERNAME_MIN_LENGTH" envDefault:"3"`
	// MaximumLength is the longest accepted username in runes.
	MaximumLength int `env:"PIXELS_PLAYER_USERNAME_MAX_LENGTH" envDefault:"15"`
	// AllowedSymbols contains accepted non-alphanumeric ASCII characters.
	AllowedSymbols string `env:"PIXELS_PLAYER_USERNAME_ALLOWED_SYMBOLS" envDefault:"_-=!?@:,.'"`
	// ReservationTTL bounds an available username claim.
	ReservationTTL time.Duration `env:"PIXELS_PLAYER_USERNAME_RESERVATION_TTL" envDefault:"2m"`
	// ReservedNames contains case-insensitive exact names unavailable to players.
	ReservedNames []string `env:"PIXELS_PLAYER_USERNAME_RESERVED" envDefault:"admin,moderator,staff,system" envSeparator:","`
	// ChangeCooldownDays is the number of full days between username changes.
	ChangeCooldownDays int `env:"PIXELS_PLAYER_USERNAME_CHANGE_COOLDOWN_DAYS" envDefault:"30"`
}

// LoadConfig loads username policy from environment variables and defaults.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// DefaultConfig returns the documented username policy defaults.
func DefaultConfig() Config {
	return Config{MinimumLength: 3, MaximumLength: 15, AllowedSymbols: "_-=!?@:,.'", ReservationTTL: 2 * time.Minute, ReservedNames: []string{"admin", "moderator", "staff", "system"}, ChangeCooldownDays: 30}
}

// changeCooldown returns the configured interval as a duration.
func (config Config) changeCooldown() time.Duration {
	return time.Duration(config.ChangeCooldownDays) * 24 * time.Hour
}
