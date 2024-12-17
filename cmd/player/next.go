package player

import (
	"context"
	"fmt"
	"log"

	"github.com/Esteban-Bermudez/spotgo/cmd/connect"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var playerNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Skip to the next track",
	Long:  `Skip to the next track in the current playback session`,
	Run:   spotifyNext,
}

func init() {
	playerCmd.AddCommand(playerPauseCmd)
}

func spotifyNext(cmd *cobra.Command, args []string) {
	token, err := connect.LoadOAuthToken()
	if err != nil {
		log.Fatal("Error loading token, Run `spotgo connect` to connect to Spotify")
	}

	client := spotify.New(connect.Auth.Client(context.Background(), token))

	client.Next(context.Background())

	fmt.Println("Next track")
}
