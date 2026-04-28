package server

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestResolvePort(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "empty uses default", input: "", want: defaultPort},
		{name: "whitespace uses default", input: "   ", want: defaultPort},
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
			got, err := resolvePort(tt.input)
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
