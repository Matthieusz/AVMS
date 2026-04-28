//go:build !liboqs

package pqc

import "fmt"

func RunKEMCheck(kemName string) (KEMCheckResult, error) {
	return KEMCheckResult{}, fmt.Errorf("liboqs support is disabled: rebuild with -tags liboqs and ensure liboqs is installed")
}
