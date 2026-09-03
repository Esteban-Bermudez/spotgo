package session

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/Esteban-Bermudez/spotgo/config"
	librespot "github.com/devgianlu/go-librespot"
	"github.com/devgianlu/go-librespot/daemon"
	"github.com/spf13/viper"
)

const (
	DeviceName = "spotgo"
	DeviceType = "computer"
)

// Session wraps a go-librespot daemon.App that registers spotgo as a Spotify
// Connect device. Web API (spotgo connect, 762c… with streaming) and the
// embedded librespot daemon are separate Spotify apps: the daemon uses
// go-librespot's whitelisted app 65b7… via Interactive OAuth (browser on
// 127.0.0.1) or Zeroconf, and its credentials are persisted in
// librespot-state.json. Audio renders via native backends (audio-toolbox on
// darwin, pulseaudio on linux, wasapi on windows) — no oto, no external
// spotifyd binary.
type Session struct {
	app *daemon.App
	log librespot.Logger
}

func audioBackend() string {
	switch runtime.GOOS {
	case "darwin":
		return "audio-toolbox"
	case "windows":
		return "wasapi"
	default:
		return "pulseaudio"
	}
}

func statePath() string {
	return filepath.Join(config.DataDir(), "librespot-state.json")
}

func newDaemonConfig() *daemon.Config {
	cfg := &daemon.Config{
		DeviceName:   DeviceName,
		DeviceType:   DeviceType,
		AudioBackend: audioBackend(),
		Bitrate:      320,
		VolumeSteps:  100,
		ImageSize:    "default",
		// Zeroconf off for now: the builtin mDNS Register panics at
		// zeroconf.go:64/app.go:393 on this macOS (spotifyd's own responder
		// works, ours doesn't via the embedded daemon) and crashes the TUI.
		// Interactive login doesn't need mDNS; re-enable once builtin is
		// fixed or avahi is available.
		ZeroconfEnabled: false,
		ZeroconfBackend: "builtin",
		Cache: daemon.CacheConfig{
			Enabled: false,
		},
	}
	cfg.Credentials.Type = "interactive"
	cfg.Credentials.Interactive.CallbackPort = 0
	return cfg
}

// New creates a Session that authenticates via go-librespot's own OAuth app
// (65b7… whitelisted for client token + AP). On first run it opens a
// browser; credentials are then persisted to librespot-state.json and reused.
// The Web API token (spotgo connect, 762c… streaming) is separate.
func New(_ context.Context, _ *viper.Viper) (*Session, error) {
	l := newLogger()
	path := statePath()
	store := newFileStateStore(path, l)
	cfg := newDaemonConfig()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	app, err := daemon.New(&daemon.Options{
		Logger:     l,
		Config:     cfg,
		StateStore: store,
	})
	if err != nil {
		return nil, fmt.Errorf("create daemon: %w", err)
	}
	return &Session{app: app, log: l}, nil
}

// HasStoredCredentials reports whether a previous speaker login persisted
// credentials to disk. When false, the player must not start an interactive
// login (no browser from the TUI) and should point at `spotgo connect`.
func HasStoredCredentials() bool {
	l := newLogger()
	store := newFileStateStore(statePath(), l)
	state, err := store.Load()
	if err != nil || state == nil {
		return false
	}
	return state.Credentials.Username != "" && len(state.Credentials.Data) > 0
}

// Setup runs the one-time interactive speaker login (browser on 127.0.0.1)
// and returns once credentials are persisted, or ctx expires. Call only
// from `spotgo connect`, never from the player TUI.
func Setup(ctx context.Context) error {
	if HasStoredCredentials() {
		return nil
	}
	sess, err := New(ctx, nil)
	if err != nil {
		return err
	}
	defer sess.Close()

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	runErr := make(chan error, 1)
	go func() { runErr <- sess.Run(runCtx) }()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-runErr:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if HasStoredCredentials() {
				return nil
			}
		}
	}
}

// Run blocks until ctx is cancelled or the daemon exits with error.
// It registers mDNS (Zeroconf) and connects to the Spotify AP.
func (s *Session) Run(ctx context.Context) error {
	return s.app.Run(ctx)
}

func (s *Session) Close() error {
	return s.app.Close()
}
