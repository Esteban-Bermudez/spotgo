
package player

import (
	"context"
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var playerVolumeCmd = &cobra.Command{
	Use:   "volume",
	Short: "Set the volume of the current device",
	Long:  `Set the volume of the current device to a specified level (0-100)`,
	Run:   spotifyVolume,
}

func init() {
	playerCmd.AddCommand(playerVolumeCmd)
}

func spotifyVolume(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		state, err := spotgoClient.PlayerState(context.Background())
		if err != nil {
			log.Fatalf("Error fetching player state: %v", err)
		}
		currentVolume := state.Device.Volume
		fmt.Printf("Current volume: %d%%\n", currentVolume)
		return
	}

	var volume int
	_, err := fmt.Sscanf(args[0], "%d", &volume)
	if err != nil || volume < 0 || volume > 100 {
		log.Fatal("Invalid volume level. Please provide a number between 0 and 100")
	}

	err = spotgoClient.Volume(context.Background(), volume)
	if err != nil {
		log.Fatalf("Error setting volume: %v", err)
	}
}
