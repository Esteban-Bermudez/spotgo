package devices

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var switchCmd = &cobra.Command{
	Use:   "switch <device-id-or-name>",
	Short: "Transfer playback to another device",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		devices, err := spotgoClient.PlayerDevices(context.Background())
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

		err = spotgoClient.TransferPlayback(context.Background(), targetID, true)
		if err != nil {
			log.Fatalf("Failed to transfer playback: %v", err)
		}
		fmt.Println("Playback transferred successfully.")
	},
}

func init() {
	DevicesCmd.AddCommand(switchCmd)
}