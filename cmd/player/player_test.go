package player

import (
	"testing"
	"time"

	"github.com/zmb3/spotify/v2"
)

func trackState(progress, duration int, playing bool) *spotify.PlayerState {
	state := &spotify.PlayerState{}
	state.Progress = spotify.Numeric(progress)
	state.Playing = playing
	state.Item = &spotify.FullTrack{
		SimpleTrack: spotify.SimpleTrack{Duration: spotify.Numeric(duration)},
	}
	return state
}

func TestInterpolatedProgressMS(t *testing.T) {
	now := time.Now()

	// Regression: when nothing is playing the Spotify API returns a non-nil
	// state with a nil Item (HTTP 204). Dereferencing Item used to segfault.
	t.Run("nil state returns 0", func(t *testing.T) {
		if got := interpolatedProgressMS(nil, now); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("nil item returns 0", func(t *testing.T) {
		if got := interpolatedProgressMS(&spotify.PlayerState{}, now); got != 0 {
			t.Errorf("got %d, want 0", got)
		}
	})

	t.Run("paused returns progress without interpolation", func(t *testing.T) {
		// fetchedAt long ago, but paused, so no wall-clock time is added.
		got := interpolatedProgressMS(trackState(5000, 240000, false), now.Add(-time.Hour))
		if got != 5000 {
			t.Errorf("got %d, want 5000", got)
		}
	})

	t.Run("clamps to duration", func(t *testing.T) {
		got := interpolatedProgressMS(trackState(300000, 240000, false), now)
		if got != 240000 {
			t.Errorf("got %d, want 240000", got)
		}
	})
}
