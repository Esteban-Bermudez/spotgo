package player

import (
	"context"
	"log"

	"github.com/Esteban-Bermudez/spotgo/internal/playback"
	"github.com/spf13/cobra"
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
	err := playback.Resume(context.Background(), spotgoClient)
	if err != nil {
		log.Fatalf("Error starting or resuming playback: %v", err)
	}
}
