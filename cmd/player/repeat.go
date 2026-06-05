package player

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var playerRepeatCmd = &cobra.Command{
	Use:   "repeat",
	Short: "Toggle repeat mode for the current playback",
	Long:  `Toggle repeat mode for the current playback. This will switch between repeating the current track, repeating the entire playlist (context), or turning off repeat mode.`,
	Run:   spotifyRepeat,
}

func init() {
	playerCmd.AddCommand(playerRepeatCmd)
}

func spotifyRepeat(cmd *cobra.Command, args []string) {
	repeatState, err := spotgoClient.PlayerState(context.Background())
	if err != nil {
		log.Fatalf("Error getting player state: %v", err)
	}

	if repeatState == nil {
		log.Fatal("No active playback found. Please start playing something on Spotify and try again.")
	}

	var newRepeatMode string
	switch repeatState.RepeatState {
	case "off":
		newRepeatMode = "context"
	case "context":
		newRepeatMode = "track"
	case "track":
		newRepeatMode = "off"
	default:
		newRepeatMode = "off"
	}

	err = spotgoClient.Repeat(context.Background(), newRepeatMode)
	if err != nil {
		log.Fatalf("Error toggling repeat: %v", err)
	}
	log.Printf("Repeat mode is now set to %v", newRepeatMode)
}
