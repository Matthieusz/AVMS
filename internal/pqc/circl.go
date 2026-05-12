package pqc

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/cloudflare/circl/kem"
	"github.com/cloudflare/circl/kem/mlkem/mlkem1024"
	"github.com/cloudflare/circl/kem/mlkem/mlkem512"
	"github.com/cloudflare/circl/kem/mlkem/mlkem768"
	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
	"github.com/cloudflare/circl/sign/slhdsa"
)

func SupportedKEMAlgorithms() []string {
	return []string{"ML-KEM-512", "ML-KEM-768", "ML-KEM-1024"}
}

func SupportedSignatureAlgorithms() []string {
	return []string{"ML-DSA", "SLH-DSA"}
}

func ValidateKEMPublicKey(algorithm, encoded string) error {
	scheme, err := kemSchemeByName(algorithm)
	if err != nil {
		return err
	}

	raw, err := decodeBase64Value(encoded, "public key")
	if err != nil {
		return err
	}

	if _, err := scheme.UnmarshalBinaryPublicKey(raw); err != nil {
		return fmt.Errorf("parse %s public key: %w", strings.TrimSpace(algorithm), err)
	}

	return nil
}

func ValidateSignaturePublicKey(algorithm, encoded string) error {
	publicKey, err := decodeBase64Value(encoded, "public key")
	if err != nil {
		return err
	}

	switch normalizeSignatureAlgorithm(algorithm) {
	case "ML-DSA":
		var key mldsa65.PublicKey
		if err := key.UnmarshalBinary(publicKey); err != nil {
			return fmt.Errorf("parse ML-DSA public key: %w", err)
		}
	case "SLH-DSA":
		var key slhdsa.PublicKey
		if err := key.UnmarshalBinary(publicKey); err != nil {
			return fmt.Errorf("parse SLH-DSA public key: %w", err)
		}
	default:
		return fmt.Errorf("unsupported signature algorithm %q", strings.TrimSpace(algorithm))
	}

	return nil
}

func VerifySignature(algorithm, encodedPublicKey string, message []byte, encodedSignature string) error {
	publicKey, err := decodeBase64Value(encodedPublicKey, "public key")
	if err != nil {
		return err
	}

	signature, err := decodeBase64Value(encodedSignature, "signature")
	if err != nil {
		return err
	}

	switch normalizeSignatureAlgorithm(algorithm) {
	case "ML-DSA":
		var key mldsa65.PublicKey
		if err := key.UnmarshalBinary(publicKey); err != nil {
			return fmt.Errorf("parse ML-DSA public key: %w", err)
		}
		if !mldsa65.Verify(&key, message, nil, signature) {
			return errors.New("invalid ML-DSA signature")
		}
	case "SLH-DSA":
		var key slhdsa.PublicKey
		if err := key.UnmarshalBinary(publicKey); err != nil {
			return fmt.Errorf("parse SLH-DSA public key: %w", err)
		}
		if !slhdsa.Verify(&key, slhdsa.NewMessage(message), signature, nil) {
			return errors.New("invalid SLH-DSA signature")
		}
	default:
		return fmt.Errorf("unsupported signature algorithm %q", strings.TrimSpace(algorithm))
	}

	return nil
}

func GetRSUBeacon(rsuID, kemName string) (RSUBeacon, error) {
	rsuID = strings.TrimSpace(rsuID)
	if rsuID == "" {
		return RSUBeacon{}, fmt.Errorf("rsuId is required")
	}

	scheme, err := kemSchemeByName(kemName)
	if err != nil {
		return RSUBeacon{}, err
	}

	publicKey, _, err := deriveRSUKeyPair(rsuID, scheme)
	if err != nil {
		return RSUBeacon{}, err
	}

	rawPublicKey, err := publicKey.MarshalBinary()
	if err != nil {
		return RSUBeacon{}, fmt.Errorf("marshal %s public key: %w", strings.TrimSpace(kemName), err)
	}

	fingerprint := sha256.Sum256(rawPublicKey)

	return RSUBeacon{
		RSUID:        rsuID,
		KEMAlgorithm: strings.TrimSpace(kemName),
		KEMPublicKey: base64.StdEncoding.EncodeToString(rawPublicKey),
		KeyVersion:   hex.EncodeToString(fingerprint[:6]),
		Details:      fmt.Sprintf("Deterministic demo beacon for %s using %s via CIRCL pure-Go.", rsuID, strings.TrimSpace(kemName)),
	}, nil
}

func DecapsulateRSU(rsuID, kemName, encodedCiphertext string) ([]byte, error) {
	scheme, err := kemSchemeByName(kemName)
	if err != nil {
		return nil, err
	}

	_, privateKey, err := deriveRSUKeyPair(rsuID, scheme)
	if err != nil {
		return nil, err
	}

	ciphertext, err := decodeBase64Value(encodedCiphertext, "ciphertext")
	if err != nil {
		return nil, err
	}

	sharedSecret, err := scheme.Decapsulate(privateKey, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("decapsulate %s ciphertext: %w", strings.TrimSpace(kemName), err)
	}

	return sharedSecret, nil
}

func SessionKeyReference(sharedSecret []byte) string {
	fingerprint := sha256.Sum256(sharedSecret)
	return "sess_" + hex.EncodeToString(fingerprint[:8])
}

func kemSchemeByName(name string) (kem.Scheme, error) {
	switch strings.TrimSpace(name) {
	case "ML-KEM-512":
		return mlkem512.Scheme(), nil
	case "ML-KEM-768":
		return mlkem768.Scheme(), nil
	case "ML-KEM-1024":
		return mlkem1024.Scheme(), nil
	default:
		return nil, fmt.Errorf("unsupported KEM algorithm %q", strings.TrimSpace(name))
	}
}

func deriveRSUKeyPair(rsuID string, scheme kem.Scheme) (kem.PublicKey, kem.PrivateKey, error) {
	seed := deterministicSeed("rsu:"+strings.TrimSpace(rsuID)+":"+scheme.Name(), scheme.SeedSize())
	publicKey, privateKey := scheme.DeriveKeyPair(seed)
	return publicKey, privateKey, nil
}

func deterministicSeed(label string, size int) []byte {
	seed := make([]byte, 0, size)
	for counter := 0; len(seed) < size; counter++ {
		blockInput := []byte(fmt.Sprintf("%s:%d", label, counter))
		block := sha256.Sum256(blockInput)
		seed = append(seed, block[:]...)
	}
	return seed[:size]
}

func decodeBase64Value(value, fieldName string) ([]byte, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, fmt.Errorf("%s is required", fieldName)
	}

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", fieldName, err)
	}

	return decoded, nil
}

func normalizeSignatureAlgorithm(name string) string {
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "ML-DSA", "ML-DSA-65":
		return "ML-DSA"
	case "SLH-DSA", "SLH-DSA-SHAKE-128S":
		return "SLH-DSA"
	default:
		return strings.TrimSpace(name)
	}
}

func kemDetails(scheme kem.Scheme) string {
	return fmt.Sprintf("CIRCL pure-Go %s ciphertext=%d sharedSecret=%d publicKey=%d", scheme.Name(), scheme.CiphertextSize(), scheme.SharedKeySize(), scheme.PublicKeySize())
}

func equalBytes(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}