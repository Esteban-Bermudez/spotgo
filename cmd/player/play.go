package player

import (
	"context"
	"fmt"
	"log"

	"github.com/Esteban-Bermudez/spotgo/cmd/connect"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var playerPlayCmd = &cobra.Command{
	Use:   "play",
	Short: "Start or resume playback session",
	Long:  `Start or resume the last playback session on Spotify`,
	Run:   spotifyPlay,
}

func init() {
	playerCmd.AddCommand(playerPlayCmd)
}

func spotifyPlay(cmd *cobra.Command, args []string) {
	token, err := connect.LoadOAuthToken()
	if err != nil {
		log.Fatal("Error loading token, Run `spotgo connect` to connect to Spotify")
	}

	client := spotify.New(connect.Auth.Client(context.Background(), token))

	client.Play(context.Background())

	fmt.Println("Playback resumed")
}
