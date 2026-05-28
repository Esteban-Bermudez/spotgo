package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigDirUsesXDGConfigHome(t *testing.T) {
	xdgConfigHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdgConfigHome)
	t.Setenv("HOME", t.TempDir())

	got := ConfigDir()
	want := filepath.Join(xdgConfigHome, "spotgo")

	if got != want {
		t.Fatalf("expected config dir %q, got %q", want, got)
	}
}

func TestConfigDirFallsBackToDotConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", home)

	got := ConfigDir()
	want := filepath.Join(home, ".config", "spotgo")

	if got != want {
		t.Fatalf("expected config dir %q, got %q", want, got)
	}
}

func TestLoadConfigUsesConfigDirFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", home)

	configDir := filepath.Join(home, ".config", "spotgo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	configFile := filepath.Join(configDir, "spotgo.json")
	expiry := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	configJSON := []byte(`{
  "token": {
    "access_token": "access",
    "token_type": "Bearer",
    "refresh_token": "refresh",
    "expiry": "` + expiry + `"
  }
}`)
	if err := os.WriteFile(configFile, configJSON, 0600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	v, err := LoadConfig()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if v.ConfigFileUsed() != configFile {
		t.Fatalf("expected config file %q, got %q", configFile, v.ConfigFileUsed())
	}
}
