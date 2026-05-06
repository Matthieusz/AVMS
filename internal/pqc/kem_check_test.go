package pqc

import (
	"errors"
	"testing"
)

type stubKEM struct {
	publicKey    []byte
	ciphertext   []byte
	sharedSecret []byte
	details      string

	generateErr error
	encapErr    error
	decapErr    error
}

func (s *stubKEM) GenerateKeyPair() ([]byte, error) {
	if s.generateErr != nil {
		return nil, s.generateErr
	}
	return s.publicKey, nil
}

func (s *stubKEM) EncapSecret(publicKey []byte) (ciphertext, sharedSecret []byte, err error) {
	if s.encapErr != nil {
		return nil, nil, s.encapErr
	}
	return s.ciphertext, s.sharedSecret, nil
}

func (s *stubKEM) DecapSecret(ciphertext []byte) ([]byte, error) {
	if s.decapErr != nil {
		return nil, s.decapErr
	}
	return s.sharedSecret, nil
}

func (s *stubKEM) Details() string {
	return s.details
}

func (s *stubKEM) Clean() {}

func TestRunKEMCheckSuccess(t *testing.T) {
	client := &stubKEM{
		publicKey:    []byte("pk"),
		ciphertext:   []byte("ct"),
		sharedSecret: []byte("ss"),
		details:      "test-details",
	}
	server := &stubKEM{
		publicKey:    []byte("pk"),
		ciphertext:   []byte("ct"),
		sharedSecret: []byte("ss"),
	}

	result, err := runKEMCheck(client, server, "TEST-KEM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.KEMName != "TEST-KEM" {
		t.Fatalf("unexpected kem name: got %q, want %q", result.KEMName, "TEST-KEM")
	}
	if result.Details != "test-details" {
		t.Fatalf("unexpected details: got %q, want %q", result.Details, "test-details")
	}
	if !result.SharedSecretsCoincide {
		t.Fatal("expected shared secrets to coincide")
	}
}

func TestRunKEMCheckMismatchedSecrets(t *testing.T) {
	client := &stubKEM{
		publicKey:    []byte("pk"),
		ciphertext:   []byte("ct"),
		sharedSecret: []byte("ss-client"),
	}
	server := &stubKEM{
		publicKey:    []byte("pk"),
		ciphertext:   []byte("ct"),
		sharedSecret: []byte("ss-server"),
	}

	result, err := runKEMCheck(client, server, "TEST-KEM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.SharedSecretsCoincide {
		t.Fatal("expected shared secrets not to coincide")
	}
}

func TestRunKEMCheckGenerateKeyPairFailure(t *testing.T) {
	client := &stubKEM{
		generateErr: errors.New("keypair failed"),
	}
	server := &stubKEM{}

	_, err := runKEMCheck(client, server, "TEST-KEM")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunKEMCheckEncapFailure(t *testing.T) {
	client := &stubKEM{
		publicKey: []byte("pk"),
	}
	server := &stubKEM{
		encapErr: errors.New("encap failed"),
	}

	_, err := runKEMCheck(client, server, "TEST-KEM")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRunKEMCheckDecapFailure(t *testing.T) {
	client := &stubKEM{
		publicKey:  []byte("pk"),
		ciphertext: []byte("ct"),
		decapErr:   errors.New("decap failed"),
	}
	server := &stubKEM{
		publicKey:    []byte("pk"),
		ciphertext:   []byte("ct"),
		sharedSecret: []byte("ss"),
	}

	_, err := runKEMCheck(client, server, "TEST-KEM")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewKEMDisabled(t *testing.T) {
	_, err := NewKEM("TEST-KEM")
	if err == nil {
		t.Fatal("expected error when liboqs is disabled")
	}
}

func TestLiboqsVersionDisabled(t *testing.T) {
	if v := LiboqsVersion(); v != "" {
		t.Fatalf("expected empty version, got %q", v)
	}
}

func TestEnabledKEMsDisabled(t *testing.T) {
	if kems := EnabledKEMs(); kems != nil {
		t.Fatalf("expected nil, got %v", kems)
	}
}
