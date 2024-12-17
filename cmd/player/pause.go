package player

import (
	"context"
	"fmt"
	"log"

	"github.com/Esteban-Bermudez/spotgo/cmd/connect"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var playerPauseCmd = &cobra.Command{
	Use:   "pause",
	Short: "Pause playback session",
	Long:  `Pause the current playback session on Spotify`,
	Run:   spotifyPause,
}

func init() {
	playerCmd.AddCommand(playerPauseCmd)
}

func spotifyPause(cmd *cobra.Command, args []string) {
	token, err := connect.LoadOAuthToken()
	if err != nil {
		log.Fatal("Error loading token, Run `spotgo connect` to connect to Spotify")
	}

	client := spotify.New(connect.Auth.Client(context.Background(), token))

	client.Pause(context.Background())

	fmt.Println("Playback paused")
}
