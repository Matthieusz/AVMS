//go:build !liboqs

package pqc

import "fmt"

// NewKEM returns an error when liboqs support is not compiled in.
func NewKEM(kemName string) (KEM, error) {
	return nil, fmt.Errorf("liboqs support is disabled: rebuild with -tags liboqs and ensure liboqs is installed")
}

// LiboqsVersion returns an empty string when liboqs is disabled.
func LiboqsVersion() string {
	return ""
}

// EnabledKEMs returns nil when liboqs is disabled.
func EnabledKEMs() []string {
	return nil
}
