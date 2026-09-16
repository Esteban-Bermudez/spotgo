//go:build windows

package session

import "errors"

// Background detaching relies on Unix process semantics (setsid, signal
// probing), so the speaker only runs in the foreground on Windows.
func BackgroundRunning() (int, bool) { return 0, false }

func StartBackground() (int, error) {
	return 0, errors.New("speaker --background is not supported on Windows; run `spotgo speaker` in the foreground")
}
