package config

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Env                   string        `env:"ENV" envDefault:"development"`
	HTTPAddr              string        `env:"HTTP_ADDR" envDefault:":8080"`
	MetricsAddr           string        `env:"METRICS_ADDR"`
	DatabaseURL           string        `env:"DATABASE_URL,required"`
	MigrationsDatabaseURL string        `env:"MIGRATIONS_DATABASE_URL"`
	JWTSecret             string        `env:"JWT_SECRET,required"`
	JWTIssuer             string        `env:"JWT_ISSUER" envDefault:"finance-backend"`
	JWTAudience           string        `env:"JWT_AUDIENCE" envDefault:"finance-frontend"`
	JWTAccessTTL          time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	JWTRefreshTTL         time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`
	RefreshCookieName     string        `env:"REFRESH_COOKIE_NAME" envDefault:"refresh_token"`
	RefreshCookiePath     string        `env:"REFRESH_COOKIE_PATH" envDefault:"/api/auth"`
	RefreshCookieDomain   string        `env:"REFRESH_COOKIE_DOMAIN"`
	RefreshCookieSameSite string        `env:"REFRESH_COOKIE_SAME_SITE" envDefault:"Lax"`
	RefreshCookieSecure   bool          `env:"REFRESH_COOKIE_SECURE" envDefault:"false"`
	Argon2Time            uint32        `env:"ARGON2_TIME" envDefault:"1"`
	Argon2Memory          uint32        `env:"ARGON2_MEMORY" envDefault:"65536"`
	Argon2Threads         uint8         `env:"ARGON2_THREADS" envDefault:"4"`
	Argon2KeyLen          uint32        `env:"ARGON2_KEY_LEN" envDefault:"32"`
	RedisEnabled          bool          `env:"REDIS_ENABLED" envDefault:"false"`
	RedisURL              string        `env:"REDIS_URL" envDefault:"redis://localhost:6379/0"`
	BRAPIToken            string        `env:"BRAPI_TOKEN"`
	FinnhubKey            string        `env:"FINNHUB_KEY"`
	CoinGeckoKey          string        `env:"COINGECKO_KEY"`
	FXProviderKey         string        `env:"FX_PROVIDER_KEY"`
	CORSOrigins           string        `env:"CORS_ORIGINS" envDefault:"http://localhost:5173"`
}

func Load() (Config, error) {
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	if cfg.MigrationsDatabaseURL == "" {
		cfg.MigrationsDatabaseURL = cfg.DatabaseURL
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if len(c.JWTSecret) < 32 {
		return errors.New("JWT_SECRET must be at least 32 characters")
	}
	if strings.TrimSpace(c.RefreshCookieName) == "" {
		return errors.New("REFRESH_COOKIE_NAME is required")
	}
	if strings.TrimSpace(c.RefreshCookiePath) == "" {
		return errors.New("REFRESH_COOKIE_PATH is required")
	}

	if strings.TrimSpace(c.DatabaseURL) == "" {
		return errors.New("DATABASE_URL is required")
	}

	if c.Env != "development" && !strings.Contains(c.DatabaseURL, "sslmode=require") && !strings.Contains(c.DatabaseURL, "sslmode=verify") {
		return fmt.Errorf("DATABASE_URL must enable sslmode outside development")
	}

	return nil
}

func (c Config) RefreshCookieSameSiteMode() string {
	mode := strings.ToLower(strings.TrimSpace(c.RefreshCookieSameSite))
	switch mode {
	case "strict", "none":
		return mode
	default:
		return "lax"
	}
}

func (c Config) AllowedOrigins() []string {
	parts := strings.Split(c.CORSOrigins, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			origins = append(origins, part)
		}
	}
	return origins
}
