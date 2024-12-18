package player

import (
	"context"
	"fmt"
	"log"

	"github.com/Esteban-Bermudez/spotgo/cmd/connect"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var playerPreviousCmd = &cobra.Command{
	Use:   "previous",
	Short: "Skip to the previous track",
	Long:  `Skip to the previous track in the current playback session`,
	Run:   spotifyPrevious,
}

func init() {
	playerCmd.AddCommand(playerPreviousCmd)
}

func spotifyPrevious(cmd *cobra.Command, args []string) {
	token, err := connect.LoadOAuthToken()
	if err != nil {
		log.Fatal("Error loading token, Run `spotgo connect` to connect to Spotify")
	}

	client := spotify.New(connect.Auth.Client(context.Background(), token))

	client.Previous(context.Background())

	fmt.Println("Previous track")
}
