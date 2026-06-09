package player

import (
	"context"
	"fmt"
	"log"

	"github.com/Esteban-Bermudez/spotgo/internal/playback"
	"github.com/spf13/cobra"
)

var playerToggleCmd = &cobra.Command{
	Use:   "toggle",
	Short: "Toggle playback state",
	Long:  `Toggle the playback state of the current device between play and pause`,
	Run:   spotifyToggle,
}

func init() {
	playerCmd.AddCommand(playerToggleCmd)
}

func spotifyToggle(cmd *cobra.Command, args []string) {
	playerState, err := spotgoClient.PlayerState(context.Background())
	if err != nil {
		log.Fatalf("Error getting player state: %v", err)
	}

	if playerState.Playing {
		spotgoClient.Pause(context.Background())
		fmt.Println("Paused playback")
	} else {
		if err := playback.Resume(context.Background(), spotgoClient); err != nil {
			log.Fatalf("Error resuming playback: %v", err)
		}
		fmt.Println("Resumed playback")
	}
}
