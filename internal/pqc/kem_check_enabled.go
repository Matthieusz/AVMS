//go:build liboqs

package pqc

import (
	"bytes"
	"fmt"

	"github.com/open-quantum-safe/liboqs-go/oqs"
)

func RunKEMCheck(kemName string) (KEMCheckResult, error) {
	result := KEMCheckResult{
		LiboqsVersion: oqs.LiboqsVersion(),
		EnabledKEMs:   oqs.EnabledKEMs(),
		KEMName:       kemName,
	}

	client := oqs.KeyEncapsulation{}
	defer client.Clean()

	if err := client.Init(kemName, nil); err != nil {
		return KEMCheckResult{}, fmt.Errorf("client init failed: %w", err)
	}

	clientPublicKey, err := client.GenerateKeyPair()
	if err != nil {
		return KEMCheckResult{}, fmt.Errorf("client keypair generation failed: %w", err)
	}

	result.Details = fmt.Sprint(client.Details())

	server := oqs.KeyEncapsulation{}
	defer server.Clean()

	if err := server.Init(kemName, nil); err != nil {
		return KEMCheckResult{}, fmt.Errorf("server init failed: %w", err)
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
