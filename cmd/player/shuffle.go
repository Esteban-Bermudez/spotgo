package player

import (
	"context"
	"fmt"
	"log"

	"github.com/Esteban-Bermudez/spotgo/cmd/connect"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var playerShuffleCmd = &cobra.Command{
	Use:   "shuffle",
	Short: "Toggle shuffle on Spotify",
	Long:  `Toggle between shuffle states on a Spotify playback session`,
	Run:   spotifyShuffle,
}

func init() {
	playerCmd.AddCommand(playerShuffleCmd)
}

func spotifyShuffle(cmd *cobra.Command, args []string) {
	token, err := connect.LoadOAuthToken()
	if err != nil {
		log.Fatal("Error loading token, Run `spotgo connect` to connect to Spotify")
	}

	client := spotify.New(connect.Auth.Client(context.Background(), token))

	playerState, err := client.PlayerState(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	shuffleState := !playerState.ShuffleState

	client.Shuffle(context.Background(), shuffleState)

	if shuffleState {
		fmt.Println("Enabled shuffle")
	} else {
		fmt.Println("Disabled shuffle")
	}
}
