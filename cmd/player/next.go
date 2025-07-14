package player

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var playerNextCmd = &cobra.Command{
	Use:   "next",
	Short: "Skip to the next track",
	Long:  `Skip to the next track in the current playback session`,
	Run:   spotifyNext,
}

func init() {
	playerCmd.AddCommand(playerNextCmd)
}

func spotifyNext(cmd *cobra.Command, args []string) {
	err := spotgoClient.Next(context.Background())
	if err != nil {
		log.Fatalf("Error skipping to next track: %v", err)
	}

	fmt.Println("Next track")
}
