package config

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
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

// lockToken acquires an exclusive, cross-process advisory lock so that two
// spotgo instances (e.g. a foreground command and a long-running
// `player --oneline` in tmux) never refresh — and therefore rotate — the same
// Spotify refresh token concurrently. It returns an unlock function that the
// caller must invoke (typically via defer).
func lockToken() (func(), error) {
	if err := os.MkdirAll(ConfigDir(), 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	lockPath := filepath.Join(ConfigDir(), "token.lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open token lock file: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("failed to acquire token lock: %w", err)
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
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
