//go:build !liboqs

package pqc

import "fmt"

type KEMCheckResult struct {
	LiboqsVersion         string   `json:"liboqsVersion"`
	EnabledKEMs           []string `json:"enabledKEMs"`
	KEMName               string   `json:"kemName"`
	Details               string   `json:"details"`
	SharedSecretsCoincide bool     `json:"sharedSecretsCoincide"`
}

func RunKEMCheck(kemName string) (KEMCheckResult, error) {
	return KEMCheckResult{}, fmt.Errorf("liboqs support is disabled: rebuild with -tags liboqs and ensure liboqs is installed")
}
