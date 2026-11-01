package config

import (
	"slices"
	"testing"
	"time"
)

func env(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.RateLimitRequests != 20 || cfg.RateLimitWindow != time.Minute {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
	if !slices.Equal(cfg.TrustedProxies, DefaultTrustedProxies) {
		t.Errorf("trusted proxies = %v", cfg.TrustedProxies)
	}
}

func TestLoadOverrides(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"PORT":                "9000",
		"RATE_LIMIT_REQUESTS": "0",
		"RATE_LIMIT_WINDOW":   "30s",
		"TRUSTED_PROXIES":     " 10.0.0.1 , 172.16.0.0/12 ,",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "9000" || cfg.RateLimitRequests != 0 || cfg.RateLimitWindow != 30*time.Second {
		t.Errorf("overrides not applied: %+v", cfg)
	}
	if !slices.Equal(cfg.TrustedProxies, []string{"10.0.0.1", "172.16.0.0/12"}) {
		t.Errorf("trusted proxies = %v", cfg.TrustedProxies)
	}
}

func TestLoadTrustNoProxy(t *testing.T) {
	cfg, err := Load(env(map[string]string{"TRUSTED_PROXIES": "none"}))
	if err != nil || cfg.TrustedProxies != nil {
		t.Errorf("got %v, %v", cfg.TrustedProxies, err)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	for name, vars := range map[string]map[string]string{
		"non-numeric requests": {"RATE_LIMIT_REQUESTS": "many"},
		"negative requests":    {"RATE_LIMIT_REQUESTS": "-1"},
		"bad window":           {"RATE_LIMIT_WINDOW": "soon"},
		"zero window":          {"RATE_LIMIT_WINDOW": "0s"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Load(env(vars)); err == nil {
				t.Error("expected an error")
			}
		})
	}
}
