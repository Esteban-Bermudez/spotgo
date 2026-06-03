package player

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var playerShuffleCmd = &cobra.Command{
	Use:   "shuffle",
	Short: "Toggle shuffle mode",
	Long:  `Toggle shuffle mode for the current playback session on Spotify`,
	Run:   spotifyShuffle,
}

func init() {
	playerCmd.AddCommand(playerShuffleCmd)
}

func spotifyShuffle(cmd *cobra.Command, args []string) {
	shuffleState, err := spotgoClient.PlayerState(context.Background())
	if err != nil {
		log.Fatalf("Error getting player state: %v", err)
	}

	err = spotgoClient.Shuffle(context.Background(), !shuffleState.ShuffleState)
	if err != nil {
		log.Fatalf("Error toggling shuffle: %v", err)
	}

	log.Printf("Shuffle is now %v", !shuffleState.ShuffleState)

}
