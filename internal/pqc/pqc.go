package pqc

import (
	"bytes"
	"fmt"
)

// KEM is the interface for key encapsulation mechanism operations.
type KEM interface {
	GenerateKeyPair() ([]byte, error)
	EncapSecret(publicKey []byte) (ciphertext []byte, sharedSecret []byte, err error)
	DecapSecret(ciphertext []byte) ([]byte, error)
	Details() string
	Clean()
}

// KEMCheckResult holds the outcome of a KEM compatibility check.
type KEMCheckResult struct {
	LiboqsVersion         string   `json:"liboqsVersion"`
	EnabledKEMs           []string `json:"enabledKEMs"`
	KEMName               string   `json:"kemName"`
	Details               string   `json:"details"`
	SharedSecretsCoincide bool     `json:"sharedSecretsCoincide"`
}

// RunKEMCheck creates two KEM instances and verifies that the shared secrets
// produced by client and server coincide.
func RunKEMCheck(kemName string) (KEMCheckResult, error) {
	client, err := NewKEM(kemName)
	if err != nil {
		return KEMCheckResult{}, err
	}
	defer client.Clean()

	server, err := NewKEM(kemName)
	if err != nil {
		return KEMCheckResult{}, err
	}
	defer server.Clean()

	result, err := runKEMCheck(client, server, kemName)
	if err != nil {
		return KEMCheckResult{}, err
	}

	result.LiboqsVersion = LiboqsVersion()
	result.EnabledKEMs = EnabledKEMs()

	return result, nil
}

func runKEMCheck(client, server KEM, kemName string) (KEMCheckResult, error) {
	result := KEMCheckResult{
		KEMName: kemName,
		Details: client.Details(),
	}

	clientPublicKey, err := client.GenerateKeyPair()
	if err != nil {
		return KEMCheckResult{}, fmt.Errorf("client keypair generation failed: %w", err)
	}

	ciphertext, sharedSecretServer, err := server.EncapSecret(clientPublicKey)
	if err != nil {
		return KEMCheckResult{}, fmt.Errorf("server encapsulation failed: %w", err)
	}

	sharedSecretClient, err := client.DecapSecret(ciphertext)
	if err != nil {
		return KEMCheckResult{}, fmt.Errorf("client decapsulation failed: %w", err)
	}

	result.SharedSecretsCoincide = bytes.Equal(sharedSecretClient, sharedSecretServer)

	return result, nil
}
