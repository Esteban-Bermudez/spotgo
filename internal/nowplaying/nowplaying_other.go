//go:build !darwin

package nowplaying

import "github.com/zmb3/spotify/v2"

// On non-macOS platforms there is no OS now-playing surface to integrate with
// (spotifyd exposes MPRIS on Linux), so these are no-ops: Run just runs the
// player directly and Update does nothing.

func Run(_ *spotify.Client, worker func()) { worker() }

func Update(_ *spotify.PlayerState, _ int) {}
