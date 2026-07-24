package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port int

	DatabaseDSN string

	SessionTTL time.Duration

	LinkCodeTTL time.Duration

	InternalSharedSecret string

	SessionCookieSecure bool
}

const (
	DefaultPort        = 8081
	DefaultSessionTTL  = 30 * 24 * time.Hour
	DefaultLinkCodeTTL = 10 * time.Minute
)

func Load() (*Config, error) {
	cfg := &Config{
		Port:                DefaultPort,
		SessionTTL:          DefaultSessionTTL,
		LinkCodeTTL:         DefaultLinkCodeTTL,
		SessionCookieSecure: true,
	}

	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.Port = n
		}
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.DatabaseDSN = v
	}
	if v := os.Getenv("SESSION_TTL_HOURS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.SessionTTL = time.Duration(n) * time.Hour
		}
	}
	if v := os.Getenv("LINK_CODE_TTL_MINUTES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.LinkCodeTTL = time.Duration(n) * time.Minute
		}
	}
	if v := os.Getenv("INTERNAL_SHARED_SECRET"); v != "" {
		cfg.InternalSharedSecret = v
	}
	if v := os.Getenv("SESSION_COOKIE_SECURE"); v != "" {
		cfg.SessionCookieSecure = v != "false"
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.DatabaseDSN == "" {
		return fmt.Errorf("DATABASE_DSN is required")
	}
	if c.InternalSharedSecret == "" {
		return fmt.Errorf("INTERNAL_SHARED_SECRET is required")
	}
	if len(c.InternalSharedSecret) < 16 {
		return fmt.Errorf("INTERNAL_SHARED_SECRET is too short to be a real secret (got %d chars, want at least 16)", len(c.InternalSharedSecret))
	}
	return nil
}
