package config

import (
	"reflect"
	"testing"
)

func TestParsePort(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "empty uses default", input: "", want: defaultPort},
		{name: "whitespace uses default", input: "   ", wantErr: true},
		{name: "valid port", input: "8081", want: 8081},
		{name: "min allowed port", input: "1", want: 1},
		{name: "max allowed port", input: "65535", want: 65535},
		{name: "invalid text", input: "abc", wantErr: true},
		{name: "zero out of range", input: "0", wantErr: true},
		{name: "negative out of range", input: "-1", wantErr: true},
		{name: "too large out of range", input: "65536", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePort(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (port=%d)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("unexpected port: got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseAllowedOrigins(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty uses default",
			input: "",
			want:  []string{defaultAllowedOrigin},
		},
		{
			name:  "parses and deduplicates origins",
			input: " http://localhost:3000, http://localhost:5173, http://localhost:3000, *, ",
			want:  []string{"http://localhost:3000", "http://localhost:5173"},
		},
		{
			name:  "only invalid entries falls back",
			input: " , *, ",
			want:  []string{defaultAllowedOrigin},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAllowedOrigins(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("unexpected origins: got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Server.Port != defaultPort {
		t.Fatalf("unexpected port: got %d, want %d", cfg.Server.Port, defaultPort)
	}
	if cfg.Server.StaticDir != defaultStaticDir {
		t.Fatalf("unexpected static dir: got %q, want %q", cfg.Server.StaticDir, defaultStaticDir)
	}
	if len(cfg.Server.AllowedOrigins) != 1 || cfg.Server.AllowedOrigins[0] != defaultAllowedOrigin {
		t.Fatalf("unexpected origins: got %v", cfg.Server.AllowedOrigins)
	}
	if cfg.DB.URL != defaultDBURL {
		t.Fatalf("unexpected db url: got %q, want %q", cfg.DB.URL, defaultDBURL)
	}
}

func TestLoad(t *testing.T) {
	// Override any values that may have been loaded from a .env file so the
	// test is deterministic.
	t.Setenv("AVMS_PORT", "")
	t.Setenv("AVMS_DB_URL", "")
	t.Setenv("AVMS_STATIC_DIR", "")
	t.Setenv("AVMS_CORS_ORIGINS", "")
	t.Setenv("GIN_MODE", "")
	t.Setenv("AVMS_LOG_FORMAT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != defaultPort {
		t.Fatalf("unexpected port: got %d, want %d", cfg.Server.Port, defaultPort)
	}
	if cfg.DB.URL != defaultDBURL {
		t.Fatalf("unexpected db url: got %q, want %q", cfg.DB.URL, defaultDBURL)
	}
}
