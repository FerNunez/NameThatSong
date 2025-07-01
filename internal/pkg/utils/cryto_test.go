package utils

import (
	"crypto/rand"
	"testing"
)

func TestTokenEncryptor_EncryptDecrypt(t *testing.T) {
	// Generate a 32-byte key for AES-256
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate key: %v", err)
	}

	// Create a new TokenEncryptor
	encryptor, err := NewTokenEncryptor(key)
	if err != nil {
		t.Fatalf("failed to create TokenEncryptor: %v", err)
	}

	// Define a plaintext string
	plaintext := "Hello, TokenEncryptor!"

	// Encrypt the plaintext
	ciphertext, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Decrypt the ciphertext
	decryptedText, err := encryptor.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	// Check that the decrypted text matches the original plaintext
	if decryptedText != plaintext {
		t.Errorf("decrypted text mismatch: got %q, want %q", decryptedText, plaintext)
	}
}

func TestTokenEncrypto_encryptAccessToken(t *testing.T) {
	key := []byte("12345678901234567890123456789012") // 32-byte key
	token := "BQBogOQiKLjXlLF6xdwYWtR31q8jA3oQSiE58OsaAWqtnC2BK0hwroEIRbVMAUq94TIu0yRg5tEQBWo01l_3NEl1GgKGuUHud3ii1s_yreeENn1jQ402xSKqHj656WgowsjnjCjmXs9acEQ9RKzEcK2XhYwvlzYoV7D2FhqpCnLw4Sjw4nX9DkBjdhQeSrwXiaD9eqOZCkYAgBkUQIjaBS-qF0drjVjI7sB6vx6Yzn5BZmlDl4s-jQcKGm0-MC--bViQcoE"

	encryptor, err := NewTokenEncryptor(key)
	if err != nil {
		t.Fatalf("encryptor init error: %v\n", err)
	}

	encryptedToken, err := encryptor.Encrypt(token)
	if err != nil {
		t.Fatalf("encrypt error: %v\n", err)
	}

	decryptedToken, err := encryptor.Decrypt(encryptedToken)
	if err != nil {
		t.Fatalf("decrypt error: %v\n", err)
	}

	if token != decryptedToken {
		t.Error("expected error for difference in origial and encrypted-decrypted")
	}
}
