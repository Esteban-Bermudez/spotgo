package player

import (
	"context"
	"log"

	"github.com/spf13/cobra"
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
	err := spotgoClient.Pause(context.Background())
	if err != nil {
		log.Fatalf("Error pausing playback: %v", err)
	}
}
