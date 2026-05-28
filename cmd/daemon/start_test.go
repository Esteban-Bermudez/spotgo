package daemon

import (
	"errors"
	"reflect"
	"testing"
)

func TestSpotifydBinaryPathPrefersPathOnMacOS(t *testing.T) {
	path, err := spotifydBinaryPath(
		"darwin",
		"/Users/me/.local/share/spotgo",
		func(name string) (string, error) { return "/opt/homebrew/bin/spotifyd", nil },
		func(path string) bool { return path == "/Users/me/.local/share/spotgo/bin/spotifyd" },
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/opt/homebrew/bin/spotifyd" {
		t.Fatalf("expected Homebrew spotifyd path, got %q", path)
	}
}

func TestSpotifydStartArgsUsePortAudioForegroundModeOnMacOS(t *testing.T) {
	args := spotifydStartArgs("darwin", "/Users/me/.local/state/spotgo/spotifyd.pid")
	want := []string{"--no-daemon", "--backend", "portaudio", "--device-name", "spotgo", "--device-type", "computer"}

	if !reflect.DeepEqual(args, want) {
		t.Fatalf("expected args %#v, got %#v", want, args)
	}
}

func TestSpotifydStartArgsUseNativePidFileOnLinux(t *testing.T) {
	args := spotifydStartArgs("linux", "/home/me/.local/state/spotgo/spotifyd.pid")
	want := []string{"--pid", "/home/me/.local/state/spotgo/spotifyd.pid", "--device-name", "spotgo", "--device-type", "computer"}

	if !reflect.DeepEqual(args, want) {
		t.Fatalf("expected args %#v, got %#v", want, args)
	}
}

func TestSpotifydBinaryPathFallsBackToLocalInstallOnLinux(t *testing.T) {
	path, err := spotifydBinaryPath(
		"linux",
		"/home/me/.local/share/spotgo",
		func(name string) (string, error) { return "", errors.New("not found") },
		func(path string) bool { return path == "/home/me/.local/share/spotgo/bin/spotifyd" },
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "/home/me/.local/share/spotgo/bin/spotifyd" {
		t.Fatalf("expected local spotifyd path, got %q", path)
	}
}

func TestSpotifydStartedMessageIncludesCapturedPid(t *testing.T) {
	got := spotifydStartedMessage(70672)
	want := "Started spotifyd daemon in the background (PID 70672).\n"

	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
