package admin

import "github.com/caarlos0/env/v11"

// Config controls first-party administrative command compatibility.
type Config struct {
	// AllowUnpermittedEffectClear lets players without admin.effect use :effect 0.
	AllowUnpermittedEffectClear bool `env:"PIXELS_EFFECT_ALLOW_UNPERMITTED_CLEAR" envDefault:"true"`
}

// LoadConfig reads administrative command configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}
