package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

// TokenManager generates and verifies badge tokens.
type TokenManager struct {
	secret []byte
}

const tokenSizeBytes = 32

// NewTokenManager creates a TokenManager with the provided secret key.
func NewTokenManager(secret string) (*TokenManager, error) {
	if secret == "" {
		return nil, errors.New("secret key is required")
	}
	return &TokenManager{secret: []byte(secret)}, nil
}

// GenerateToken returns a new token and its HMAC-SHA256 hash.
func (m *TokenManager) GenerateToken() (string, string, error) {
	if m == nil {
		return "", "", errors.New("token manager is not configured")
	}
	raw := make([]byte, tokenSizeBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	hash, err := m.HashToken(token)
	if err != nil {
		return "", "", err
	}
	return token, hash, nil
}

// HashToken returns the HMAC-SHA256 hash of a token.
func (m *TokenManager) HashToken(token string) (string, error) {
	if m == nil {
		return "", errors.New("token manager is not configured")
	}
	if token == "" {
		return "", errors.New("token is required")
	}
	mac := hmac.New(sha256.New, m.secret)
	_, _ = mac.Write([]byte(token))
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// CompareHash verifies that the token matches the stored hash.
func (m *TokenManager) CompareHash(hash, token string) bool {
	if hash == "" || token == "" || m == nil {
		return false
	}
	computed, err := m.HashToken(token)
	if err != nil {
		return false
	}
	return hmac.Equal([]byte(hash), []byte(computed))
}
