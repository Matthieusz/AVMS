//go:build !liboqs

package pqc

func RunKEMCheck(kemName string) (KEMCheckResult, error) {
	scheme, err := kemSchemeByName(kemName)
	if err != nil {
		return KEMCheckResult{}, err
	}

	publicKey, privateKey, err := scheme.GenerateKeyPair()
	if err != nil {
		return KEMCheckResult{}, err
	}

	ciphertext, sharedSecretEncapsulated, err := scheme.Encapsulate(publicKey)
	if err != nil {
		return KEMCheckResult{}, err
	}

	sharedSecretDecapsulated, err := scheme.Decapsulate(privateKey, ciphertext)
	if err != nil {
		return KEMCheckResult{}, err
	}

	return KEMCheckResult{
		LiboqsVersion:         "circl v1.6.3 (pure-go)",
		EnabledKEMs:           SupportedKEMAlgorithms(),
		KEMName:               kemName,
		Details:               kemDetails(scheme),
		SharedSecretsCoincide: equalBytes(sharedSecretEncapsulated, sharedSecretDecapsulated),
	}, nil
}
