package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

type TokenEncryptor struct {
	gcm cipher.AEAD
}

func NewTokenEncryptor(key []byte) (*TokenEncryptor, error) {
	// AES Cipher: core encryption algorithm
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// Galois/Counter mode:  encryption mode (Authentication & Nonce handling)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &TokenEncryptor{gcm}, nil
}

func (te *TokenEncryptor) Encrypt(plaintext string) (string, error) {
	// Random data for nonce
	nonce := make([]byte, te.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// use nonce to encrypt and append nonce
	ciphertext := te.gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// encode to string
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (te *TokenEncryptor) Decrypt(ciphertext string) (string, error) {
	// decode into bytes
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	// check nonce size
	nonceSize := te.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// retirve nonce
	nonce, cipher := data[:nonceSize], data[:nonceSize]

	// decrypt
	plaintext, err := te.gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
