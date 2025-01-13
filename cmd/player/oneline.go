package player

import (
	"context"
	"fmt"
	"github.com/zmb3/spotify/v2"
	"log"
	"os"
	"time"
)

func oneLineOutput(client *spotify.Client, noProgress bool) {
	playerState, err := client.PlayerState(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	output := ""

	if playerState.Item == nil {
		fmt.Print("  No Song Playing")
		time.Sleep(5 * time.Second)
	} else if playerState.Item != nil && playerState.Playing {
		output = fmt.Sprintf("  %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
	} else {
		output = fmt.Sprintf("  %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
	}

	if playerState.Item != nil && !noProgress {
		output = fmt.Sprintf("%s | %s", output, progressBar(playerState))
	}

	fmt.Println(output)
	os.Exit(0)
}
