package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
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

func DataDir() string {
	if xdgDataHome := os.Getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "spotgo")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "spotgo")
}

func ConfigDir() string {
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		return filepath.Join(xdgConfigHome, "spotgo")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "spotgo")
}

func StateDir() string {
	if xdgStateHome := os.Getenv("XDG_STATE_HOME"); xdgStateHome != "" {
		return filepath.Join(xdgStateHome, "spotgo")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "state", "spotgo")
}
