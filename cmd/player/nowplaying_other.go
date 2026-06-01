//go:build !darwin

package player

import "github.com/zmb3/spotify/v2"

// On non-macOS platforms there is no OS now-playing surface to integrate with
// (spotifyd exposes MPRIS on Linux), so these are no-ops: runNowPlaying just
// runs the player directly and updateNowPlaying does nothing.

func runNowPlaying(_ *spotify.Client, worker func()) { worker() }

func updateNowPlaying(_ *spotify.PlayerState, _ int) {}
