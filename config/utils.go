package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/oauth2"
)

func generateRandomString(size int) (string, error) {
	possible := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	values := make([]byte, size)
	_, err := rand.Read(values)
	if err != nil {
		return "", err
	}
	for i, b := range values {
		values[i] = possible[int(b)%len(possible)]
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(values), nil
}

func MarshalToken(token *oauth2.Token) (map[string]any, error) {
	if token == nil {
		return nil, fmt.Errorf("token is nil")
	}
	return map[string]any{
		"access_token":  token.AccessToken,
		"token_type":    token.TokenType,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry.Format(time.RFC3339),
	}, nil
}
