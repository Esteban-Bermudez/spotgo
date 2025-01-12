package player

import (
  "context"
  "fmt"
  "log"
  "os"
  "time"
	"golang.org/x/oauth2"
	"github.com/Esteban-Bermudez/spotgo/cmd/connect"
  "github.com/zmb3/spotify/v2"
)

func oneLineOutput(token *oauth2.Token, noProgress bool) {
	client := spotify.New(connect.Auth.Client(context.Background(), token))
	playerState, err := client.PlayerState(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	output := ""

	if playerState.Item == nil {
		fmt.Print("󰓇")
		time.Sleep(5 * time.Second)
	} else if playerState.Item != nil && playerState.Playing {
		output = fmt.Sprintf("󰓇  %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
	} else {
		output = fmt.Sprintf("󰓇  %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
	}

	if playerState.Item != nil && !noProgress {
		output = fmt.Sprintf("%s | %s", output, progressBar(playerState))
	}

	fmt.Println(output)
	os.Exit(0)
}
