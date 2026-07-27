package catalog

import "github.com/caarlos0/env/v11"

// Config controls catalog-wide runtime compatibility behavior.
type Config struct {
	// FreeItems starts the process with every catalog offer priced at zero.
	FreeItems bool `env:"PIXELS_CATALOG_FREE_ITEMS" envDefault:"false"`
}

// LoadConfig reads catalog configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}
