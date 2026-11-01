// Package config loads the API configuration from environment variables.
package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Defaults used when the corresponding variable is unset.
const (
	DefaultPort              = "8080"
	DefaultRateLimitRequests = 20
	DefaultRateLimitWindow   = time.Minute
)

// DefaultTrustedProxies are the networks whose X-Forwarded-For headers are
// trusted to find the real client IP: loopback and private ranges, where the
// Vite proxy, nginx, Caddy and Docker networks live.
var DefaultTrustedProxies = []string{
	"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "::1/128", "fc00::/7",
}

// Config is the runtime configuration of the API.
type Config struct {
	Port              string
	RateLimitRequests int           // per client per window; 0 disables rate limiting
	RateLimitWindow   time.Duration // refill period of the per-client budget
	TrustedProxies    []string      // CIDRs/IPs; empty trusts no proxy
}

// Load reads the configuration using getenv (os.Getenv in production).
//
//	PORT                 default 8080
//	RATE_LIMIT_REQUESTS  default 20 (0 disables)
//	RATE_LIMIT_WINDOW    default 1m (Go duration, e.g. 30s, 1m, 1h)
//	TRUSTED_PROXIES      default private networks; comma-separated, or "none"
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Port:              DefaultPort,
		RateLimitRequests: DefaultRateLimitRequests,
		RateLimitWindow:   DefaultRateLimitWindow,
		TrustedProxies:    DefaultTrustedProxies,
	}

	if v := getenv("PORT"); v != "" {
		cfg.Port = v
	}

	if v := getenv("RATE_LIMIT_REQUESTS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return Config{}, fmt.Errorf("RATE_LIMIT_REQUESTS must be a non-negative integer, got %q", v)
		}
		cfg.RateLimitRequests = n
	}

	if v := getenv("RATE_LIMIT_WINDOW"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d <= 0 {
			return Config{}, fmt.Errorf("RATE_LIMIT_WINDOW must be a positive duration like 1m, got %q", v)
		}
		cfg.RateLimitWindow = d
	}

	switch v := strings.TrimSpace(getenv("TRUSTED_PROXIES")); v {
	case "":
	case "none":
		cfg.TrustedProxies = nil
	default:
		cfg.TrustedProxies = nil
		for _, p := range strings.Split(v, ",") {
			if p = strings.TrimSpace(p); p != "" {
				cfg.TrustedProxies = append(cfg.TrustedProxies, p)
			}
		}
	}

	return cfg, nil
}
