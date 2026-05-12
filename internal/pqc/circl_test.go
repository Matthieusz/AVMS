package pqc

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/cloudflare/circl/sign/mldsa/mldsa65"
)

func TestRunKEMCheckPureGo(t *testing.T) {
	result, err := RunKEMCheck("ML-KEM-768")
	if err != nil {
		t.Fatalf("unexpected KEM check error: %v", err)
	}
	if !result.SharedSecretsCoincide {
		t.Fatal("expected shared secrets to coincide")
	}
	if result.KEMName != "ML-KEM-768" {
		t.Fatalf("unexpected KEM name: %q", result.KEMName)
	}
	if len(result.EnabledKEMs) == 0 {
		t.Fatal("expected enabled KEM list to be populated")
	}
}

func TestGetRSUBeaconAndDecapsulate(t *testing.T) {
	beacon, err := GetRSUBeacon("rsu-42", "ML-KEM-768")
	if err != nil {
		t.Fatalf("unexpected beacon error: %v", err)
	}
	if beacon.KEMPublicKey == "" {
		t.Fatal("expected RSU beacon public key")
	}
}

func TestValidateAndVerifyMLDSASignature(t *testing.T) {
	publicKey, privateKey, err := mldsa65.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ML-DSA key pair: %v", err)
	}

	rawPublicKey, err := publicKey.MarshalBinary()
	if err != nil {
		t.Fatalf("failed to marshal public key: %v", err)
	}

	message := []byte("join-payload")
	signature, err := privateKey.Sign(rand.Reader, message, nil)
	if err != nil {
		t.Fatalf("failed to sign message: %v", err)
	}

	encodedPublicKey := base64.StdEncoding.EncodeToString(rawPublicKey)
	encodedSignature := base64.StdEncoding.EncodeToString(signature)

	if err := ValidateSignaturePublicKey("ML-DSA", encodedPublicKey); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if err := VerifySignature("ML-DSA", encodedPublicKey, message, encodedSignature); err != nil {
		t.Fatalf("unexpected verification error: %v", err)
	}
	if err := VerifySignature("ML-DSA", encodedPublicKey, []byte("tampered"), encodedSignature); err == nil {
		t.Fatal("expected tampered verification to fail")
	}
}