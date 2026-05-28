package devices

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var switchCmd = &cobra.Command{
	Use:   "switch <device-id-or-name>",
	Short: "Transfer playback to another device",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		v, err := config.LoadConfig()
		if err != nil {
			log.Fatal("Error loading config.")
		}
		client, err := config.SpotifyClient(context.Background(), v)
		if err != nil {
			log.Fatal("Error creating Spotify client.")
		}

		target := args[0]
		devices, err := client.PlayerDevices(context.Background())
		if err != nil {
			log.Fatalf("Error fetching devices: %v", err)
		}

		var targetID spotify.ID
		for _, d := range devices {
			if string(d.ID) == target || strings.Contains(strings.ToLower(d.Name), strings.ToLower(target)) {
				targetID = d.ID
				fmt.Printf("Switching to device: %s\n", d.Name)
				break
			}
		}

		if targetID == "" {
			log.Fatalf("Device '%s' not found.", target)
		}

		err = client.TransferPlayback(context.Background(), targetID, true)
		if err != nil {
			log.Fatalf("Failed to transfer playback: %v", err)
		}
		fmt.Println("Playback transferred successfully.")
	},
}

func init() {
	DevicesCmd.AddCommand(switchCmd)
}