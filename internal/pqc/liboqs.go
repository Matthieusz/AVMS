//go:build liboqs

package pqc

import (
	"fmt"

	"github.com/open-quantum-safe/liboqs-go/oqs"
)

// NewKEM creates a new KEM instance backed by liboqs.
func NewKEM(kemName string) (KEM, error) {
	kem := &liboqsKEM{}
	if err := kem.kem.Init(kemName, nil); err != nil {
		return nil, fmt.Errorf("kem init failed: %w", err)
	}
	return kem, nil
}

// LiboqsVersion returns the liboqs version.
func LiboqsVersion() string {
	return oqs.LiboqsVersion()
}

// EnabledKEMs returns the list of enabled KEMs.
func EnabledKEMs() []string {
	return oqs.EnabledKEMs()
}

type liboqsKEM struct {
	kem oqs.KeyEncapsulation
}

func (k *liboqsKEM) GenerateKeyPair() ([]byte, error) {
	return k.kem.GenerateKeyPair()
}

func (k *liboqsKEM) EncapSecret(publicKey []byte) (ciphertext, sharedSecret []byte, err error) {
	return k.kem.EncapSecret(publicKey)
}

func (k *liboqsKEM) DecapSecret(ciphertext []byte) ([]byte, error) {
	return k.kem.DecapSecret(ciphertext)
}

func (k *liboqsKEM) Details() string {
	return fmt.Sprint(k.kem.Details())
}

func (k *liboqsKEM) Clean() {
	k.kem.Clean()
}
