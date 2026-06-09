// Package playback provides shared Spotify playback helpers used by both the
// player commands and the macOS now-playing integration.
package playback

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/zmb3/spotify/v2"
)

// DaemonDeviceName is the Spotify Connect name the spotgo daemon registers
// under (see the --device-name argument in `spotgo daemon start`).
const DaemonDeviceName = "spotgo"

// Resume starts or resumes playback. When Spotify reports no active device —
// routine once a session idles out, even with spotifyd still registered — it
// transfers playback to the spotgo daemon device (or the only available
// device) and resumes there, like the official client does.
func Resume(ctx context.Context, client *spotify.Client) error {
	err := client.Play(ctx)
	if !isNoActiveDevice(err) {
		return err
	}

	devices, derr := client.PlayerDevices(ctx)
	if derr != nil {
		return errors.Join(err, derr)
	}
	id, ok := pickResumeDevice(devices)
	if !ok {
		return err
	}
	// play=true resumes playback on the target as part of the transfer.
	return client.TransferPlayback(ctx, id, true)
}

// isNoActiveDevice reports whether err is Spotify's "no active device" reply
// (a 404 from the player endpoints).
func isNoActiveDevice(err error) bool {
	var spotifyErr spotify.Error
	return errors.As(err, &spotifyErr) && spotifyErr.Status == http.StatusNotFound
}

// pickResumeDevice chooses where to send playback when nothing is active: the
// spotgo daemon if it's registered, else the sole remaining device, else
// nowhere — the caller keeps the original error and the user picks explicitly
// with `spotgo devices switch`. Restricted devices can't receive Web API
// commands, so they are never chosen.
func pickResumeDevice(devices []spotify.PlayerDevice) (spotify.ID, bool) {
	for _, d := range devices {
		if strings.EqualFold(d.Name, DaemonDeviceName) && !d.Restricted {
			return d.ID, true
		}
	}
	if len(devices) == 1 && !devices[0].Restricted {
		return devices[0].ID, true
	}
	return "", false
}
