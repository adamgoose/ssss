package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
)

func deriveKey(passphrase string) []byte {
	sum := sha256.New()
	sum.Write([]byte(passphrase))
	return sum.Sum(nil)
}

func encrypt(plaintext []byte, passphrase string) ([]byte, error) {
	// Derive the key from the passphrase
	key := deriveKey(passphrase)

	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	// https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the data using aesGCM.Seal
	// Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(enctext []byte, passphrase string) ([]byte, error) {
	key := deriveKey(passphrase)

	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce size
	nonceSize := aesGCM.NonceSize()

	// Extract the nonce from the encrypted data
	nonce, ciphertext := enctext[:nonceSize], enctext[nonceSize:]

	// Decrypt the data
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}
