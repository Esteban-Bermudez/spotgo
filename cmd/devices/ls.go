package devices

import (
	"context"
	"fmt"
	"log"

	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List available Spotify Connect devices",
	Run: func(cmd *cobra.Command, args []string) {
		v, err := config.LoadConfig()
		if err != nil {
			log.Fatal("Error loading config. Run `spotgo connect`.")
		}
		client, err := config.SpotifyClient(context.Background(), v)
		if err != nil {
			log.Fatal("Error creating Spotify client.")
		}

		devices, err := client.PlayerDevices(context.Background())
		if err != nil {
			log.Fatalf("Error fetching devices: %v", err)
		}

		fmt.Printf("%-30s | %-15s | %-10s | %s\n", "NAME", "TYPE", "ACTIVE", "ID")
		fmt.Println("--------------------------------------------------------------------------------")
		for _, d := range devices {
			activeStr := ""
			if d.Active {
				activeStr = "yes"
			}
			fmt.Printf("%-30s | %-15s | %-10s | %s\n", d.Name, d.Type, activeStr, d.ID)
		}
	},
}

func init() {
	DevicesCmd.AddCommand(lsCmd)
}