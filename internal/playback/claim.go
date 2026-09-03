package playback

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zmb3/spotify/v2"
)

// ClaimDevice waits up to timeout for a non-restricted device called name to
// appear in PlayerDevices, then transfers playback to it (play=true). It is
// how the speaker makes itself the active device after (re)starting.
func ClaimDevice(ctx context.Context, client *spotify.Client, name string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		devices, err := client.PlayerDevices(ctx)
		if err == nil {
			for _, d := range devices {
				if strings.EqualFold(d.Name, name) && !d.Restricted {
					if err := client.TransferPlayback(ctx, d.ID, true); err != nil {
						return fmt.Errorf("transfer to %s: %w", name, err)
					}
					return nil
				}
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s device not seen yet", name)
		case <-ticker.C:
		}
	}
}
