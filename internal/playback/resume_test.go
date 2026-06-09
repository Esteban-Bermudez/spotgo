package playback

import (
	"errors"
	"fmt"
	"testing"

	"github.com/zmb3/spotify/v2"
)

func TestIsNoActiveDevice(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		// What the player endpoints return when no device is active.
		{"spotify 404", spotify.Error{Message: "Player command failed: No active device found", Status: 404}, true},
		{"spotify 403", spotify.Error{Message: "Player command failed: Restriction violated", Status: 403}, false},
		{"wrapped spotify 404", fmt.Errorf("resuming: %w", spotify.Error{Status: 404}), true},
		{"plain error", errors.New("connection refused"), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isNoActiveDevice(tc.err); got != tc.want {
				t.Errorf("isNoActiveDevice(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func device(name string, restricted bool) spotify.PlayerDevice {
	return spotify.PlayerDevice{ID: spotify.ID("id-" + name), Name: name, Restricted: restricted}
}

func TestPickResumeDevice(t *testing.T) {
	t.Run("prefers the spotgo daemon device", func(t *testing.T) {
		id, ok := pickResumeDevice([]spotify.PlayerDevice{
			device("Living Room TV", false),
			device("spotgo", false),
		})
		if !ok || id != "id-spotgo" {
			t.Errorf("got (%q, %v), want (id-spotgo, true)", id, ok)
		}
	})

	t.Run("matches daemon name case-insensitively", func(t *testing.T) {
		id, ok := pickResumeDevice([]spotify.PlayerDevice{device("SpotGo", false)})
		if !ok || id != "id-SpotGo" {
			t.Errorf("got (%q, %v), want (id-SpotGo, true)", id, ok)
		}
	})

	t.Run("falls back to a sole device", func(t *testing.T) {
		id, ok := pickResumeDevice([]spotify.PlayerDevice{device("Living Room TV", false)})
		if !ok || id != "id-Living Room TV" {
			t.Errorf("got (%q, %v), want sole device", id, ok)
		}
	})

	t.Run("refuses to guess between multiple devices", func(t *testing.T) {
		if id, ok := pickResumeDevice([]spotify.PlayerDevice{
			device("Living Room TV", false),
			device("Kitchen", false),
		}); ok {
			t.Errorf("got (%q, true), want no pick", id)
		}
	})

	t.Run("never picks restricted devices", func(t *testing.T) {
		if id, ok := pickResumeDevice([]spotify.PlayerDevice{device("spotgo", true)}); ok {
			t.Errorf("got (%q, true), want no pick", id)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		if _, ok := pickResumeDevice(nil); ok {
			t.Error("got a pick from an empty list")
		}
	})
}
