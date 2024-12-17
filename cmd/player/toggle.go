package player

import (
	"context"
	"fmt"
	"log"

	"github.com/Esteban-Bermudez/spotgo/cmd/connect"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var playerToggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle playback state",
	Long:  `Toggle the playback state of the current device between play and pause`,
	Run:   spotifyToggle,
}

func init() {
	playerCmd.AddCommand(playerToggleCmd)
}

func spotifyToggle(cmd *cobra.Command, args []string) {
	token, err := connect.LoadOAuthToken()
	if err != nil {
		log.Fatal("Error loading token, Run `spotgo connect` to connect to Spotify")
	}

	client := spotify.New(connect.Auth.Client(context.Background(), token))

	playerState, err := client.PlayerState(context.Background())
	if err != nil {
		log.Fatal("Error getting player state")
	}

	if playerState.Playing {
		client.Pause(context.Background())
		fmt.Println("Paused playback")
	} else {
		client.Play(context.Background())
		fmt.Println("Resumed playback")
	}
}
