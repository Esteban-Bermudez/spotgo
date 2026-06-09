package player

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Esteban-Bermudez/spotgo/internal/nowplaying"
	"github.com/zmb3/spotify/v2"
)

func oneLineOutput(client *spotify.Client, noProgress bool, scroll int) {
	output := ""
	index := 0
	icon := "󰝛 "
	var (
		playerState *spotify.PlayerState
		fetchedAt   time.Time
		lastPoll    time.Time
		polled      bool
	)
	for {
		now := time.Now()
		// Only call the Spotify API once per pollInterval; the loop ticks faster
		// so the scroll effect and progress bar stay smooth between polls. This
		// is what keeps request volume low when several players are running.
		// Also poll as soon as the current track is expected to have ended, so a
		// natural song change shows up promptly without raising the steady rate.
		dueForPoll := now.Sub(lastPoll) >= pollInterval ||
			(trackEnded(playerState, fetchedAt) && now.Sub(lastPoll) >= trackEndRepollGap)
		if !polled || dueForPoll {
			lastPoll = now
			s, err := client.PlayerState(context.Background())
			if err != nil {
				// Transient error (incl. rate limiting after the client's own
				// retries): keep the last known state and retry next interval
				// instead of dying.
				if !polled {
					time.Sleep(pollInterval)
					continue
				}
			} else {
				playerState = s
				fetchedAt = now
				polled = true
				nowplaying.Update(playerState, interpolatedProgressMS(playerState, fetchedAt))
			}
		}

		if playerState.Item == nil {
			fmt.Println("\r󰝛  No Song Playing")
			os.Exit(0)
		} else if playerState.Item != nil && playerState.Playing {
			icon = "  "
			output = fmt.Sprintf(" %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
		} else {
			icon = "  "
			output = fmt.Sprintf(" %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
		}

		if !noProgress {
			output = fmt.Sprintf(" %s | %s ", output, progressBar(interpolatedProgressMS(playerState, fetchedAt), int(playerState.Item.Duration)))
		}

		// Rotate the output string by one character to the left. This creates a
		// scrolling effect for output strings that are longer than scroll characters.
		if len(output) > scroll && scroll > 0 {
			output = icon + output[index:] + " " + output[:index]
		} else {
			output = icon + output
		}

		// This overwrites the previous line with the new song info. This is done by
		// using a carriage return character (\r) to return the cursor to the start
		// of the line and then printing the new song info.
		fmt.Printf("\r%s", output)

		if index >= len(output)-len(icon)-1 {
			index = 0
		} else {
			index++
			index = index % (len(output) - len(icon) - 1)
		}

		// Refresh the line frequently for smooth scrolling/progress; the actual
		// API poll is throttled separately above (pollInterval).
		time.Sleep(renderInterval)
	}
}
