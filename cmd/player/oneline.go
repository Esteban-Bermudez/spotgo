package player

import (
	"context"
	"fmt"
	"github.com/zmb3/spotify/v2"
	"log"
	"time"
)

func oneLineOutput(client *spotify.Client, noProgress bool) {

	output := ""
	icon := " "
	for {
		playerState, err := client.PlayerState(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		if playerState.Item == nil {
			fmt.Print("  No Song Playing")
			time.Sleep(5 * time.Second)
		} else if playerState.Item != nil && playerState.Playing {
			icon = " "
			output = fmt.Sprintf(" %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
		} else {
			icon = " "
			output = fmt.Sprintf(" %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
		}

		if playerState.Item != nil && !noProgress {
			output = fmt.Sprintf("%s | %s ", output, progressBar(playerState))
		}



		// This overwrites the previous line with the new song info. This is done by
		// using a carriage return character (\r) to return the cursor to the start
		// of the line and then printing the new song info.
		// The ANSI escape code \033[K clears the line from the current cursor
		// position to the end of the line.
    fmt.Printf("\r%s\033[K", icon + output)

		// Sleep for a second before fetching the next song info. This helps to
		// reduce the number of requests made to the Spotify API.
		time.Sleep(500 * time.Millisecond)
	}
}
