package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
)

type Keypair struct {
	Public  []byte
	Private *rsa.PrivateKey
}

type PublicKey struct {
	PublicKey []byte
}

func MakeKeypairBytes() (*Keypair, error) {
	const bitSize = 1024

	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return &Keypair{}, err
	}

	publicKey := &privateKey.PublicKey

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return &Keypair{}, err
	}

	return &Keypair{
		Public:  publicKeyBytes,
		Private: privateKey,
	}, nil
}

func (k *Keypair) Decrypt(bytearr []byte) error {

	return nil
}
