package config

import (
	"os"
	"strings"
)

// Config holds process configuration. Nothing in the domain reads the environment.
type Config struct {
	Port        string
	CORSOrigins string
}

func FromEnv() Config {
	port := strings.TrimPrefix(os.Getenv("PORT"), ":")
	if port == "" {
		port = "8080"
	}
	origins := os.Getenv("CORS_ORIGINS")
	if origins == "" {
		origins = "*"
	}
	return Config{Port: port, CORSOrigins: origins}
}

func (c Config) Address() string {
	return ":" + c.Port
}
