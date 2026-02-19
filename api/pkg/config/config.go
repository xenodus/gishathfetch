package config

import "os"

const (
	UtmSource        = "gishathfetch"
	MaxPagesToSearch = 3
	EnvProd          = "prod"
	EnvStaging       = "staging"
	EnvLocal         = "local"
)

func GetAllowedOrigins() []string {
	if os.Getenv("ENV") == EnvProd {
		return []string{
			"https://gishathfetch.com",
		}
	}

	return []string{
		"https://gishathfetch.com",
		"https://staging.gishathfetch.com",
		"http://localhost:5173",
		"http://localhost:63342", // JetBrains IDE built-in HTTP server (local dev only)
	}
}
