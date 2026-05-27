package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Service provides AES-256-GCM encryption/decryption using a system-level key.
// All commodity data (title, description, metadata) is encrypted at rest;
// the API layer transparently decrypts on read so other users can view
// products without needing the seller's RSA private key.
type Service struct {
	key []byte // 32-byte AES-256 key
}

// NewService initialises the crypto service. The key is read from the
// SYSTEM_CRYPTO_KEY environment variable (hex-encoded, 64 chars = 32 bytes).
// If the variable is empty a fatal error is logged — the system key must be
// configured explicitly in production.
func NewService() (*Service, error) {
	keyStr := os.Getenv("SYSTEM_CRYPTO_KEY")
	if keyStr == "" {
		return nil, errors.New("SYSTEM_CRYPTO_KEY environment variable is not set")
	}
	keyStr = strings.TrimSpace(keyStr)
	key, err := hex.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("SYSTEM_CRYPTO_KEY is not valid hex: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("SYSTEM_CRYPTO_KEY must be 32 bytes (64 hex chars), got %d", len(key))
	}
	return &Service{key: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns the ciphertext
// (nonce prefix + ciphertext + tag). Returns nil on error.
func (s *Service) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("read nonce: %w", err)
	}
	return aesgcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts a ciphertext produced by Encrypt. Returns the plaintext.
// The function returns an error if the ciphertext is malformed or the tag
// verification fails.
func (s *Service) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, errors.New("ciphertext is empty")
	}
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, fmt.Errorf("create cipher: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}
	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short: %d < %d", len(ciphertext), nonceSize)
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesgcm.Open(nil, nonce, ciphertext, nil)
}