package player

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zmb3/spotify/v2"
)

func oneLineOutput(client *spotify.Client, noProgress bool, scroll int) {
	output := ""
	index := 0
	icon := "󰝛 "
	for {
		playerState, err := client.PlayerState(context.Background())
		if err != nil {
			log.Fatal(err)
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

		if playerState.Item != nil && !noProgress {
			output = fmt.Sprintf(" %s | %s ", output, progressBar(playerState))
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

		// Sleep for a second before fetching the next song info. This helps to
		// reduce the number of requests made to the Spotify API.
		time.Sleep(500 * time.Millisecond)
	}
}
