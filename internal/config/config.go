package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config is the root configuration loaded from the environment.
type Config struct {
	Server ServerConfig
	DB     DBConfig
	Log    LogConfig
}

// ServerConfig holds values for the HTTP server and its middleware.
type ServerConfig struct {
	Port           int
	StaticDir      string
	AllowedOrigins []string
	GinMode        string
}

// DBConfig holds database connection values.
type DBConfig struct {
	URL string
}

// LogConfig holds logger formatting values.
type LogConfig struct {
	Format string
}

const (
	defaultPort          = 8080
	defaultDBURL         = "./avms.db"
	defaultStaticDir     = "./frontend/dist"
	defaultAllowedOrigin = "http://localhost:5173"
)

// Load reads the environment (including a .env file if present) and returns
// a validated Config. It fails fast on invalid values so that configuration
// bugs surface at process startup.
func Load() (Config, error) {
	// Load .env if present; ignore missing file.
	_ = godotenv.Load()

	var cfg Config
	var err error

	cfg.Server.Port, err = parsePort(os.Getenv("AVMS_PORT"))
	if err != nil {
		return Config{}, fmt.Errorf("config: AVMS_PORT: %w", err)
	}

	cfg.Server.StaticDir = firstNonEmpty(os.Getenv("AVMS_STATIC_DIR"), defaultStaticDir)
	cfg.Server.AllowedOrigins = parseAllowedOrigins(os.Getenv("AVMS_CORS_ORIGINS"))
	cfg.Server.GinMode = os.Getenv("GIN_MODE")

	cfg.DB.URL = firstNonEmpty(os.Getenv("AVMS_DB_URL"), defaultDBURL)

	cfg.Log.Format = os.Getenv("AVMS_LOG_FORMAT")

	return cfg, nil
}

// Default returns a Config populated with zero-override defaults.
// Tests construct Config directly, using Default as the baseline.
func Default() Config {
	return Config{
		Server: ServerConfig{
			Port:           defaultPort,
			StaticDir:      defaultStaticDir,
			AllowedOrigins: []string{defaultAllowedOrigin},
		},
		DB: DBConfig{
			URL: defaultDBURL,
		},
	}
}

func parsePort(raw string) (int, error) {
	if raw == "" {
		return defaultPort, nil
	}
	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid port %q: %w", raw, err)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("port out of range: %d", port)
	}
	return port, nil
}

func parseAllowedOrigins(raw string) []string {
	if raw == "" {
		return []string{defaultAllowedOrigin}
	}

	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))

	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" || origin == "*" {
			continue
		}
		if _, exists := seen[origin]; exists {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}

	if len(origins) == 0 {
		return []string{defaultAllowedOrigin}
	}

	return origins
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
