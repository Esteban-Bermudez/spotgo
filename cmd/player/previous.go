package player

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
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
	err := spotgoClient.Previous(context.Background())
	if err != nil {
		log.Fatalf("Error skipping to previous track: %v", err)
	}

	fmt.Println("Previous track")
}
