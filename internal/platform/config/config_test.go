package config

import "testing"

func TestFromEnvDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("CORS_ORIGINS", "")
	cfg := FromEnv()
	if cfg.Port != "8080" || cfg.Address() != ":8080" || cfg.CORSOrigins != "*" {
		t.Fatalf("%#v", cfg)
	}
}

func TestFromEnvStripsColon(t *testing.T) {
	t.Setenv("PORT", ":9090")
	t.Setenv("CORS_ORIGINS", "http://localhost:5173")
	cfg := FromEnv()
	if cfg.Address() != ":9090" || cfg.CORSOrigins != "http://localhost:5173" {
		t.Fatalf("%#v", cfg)
	}
}
