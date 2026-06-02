package devices

import (
	"context"
	"log"

	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var spotgoClient *spotify.Client

var DevicesCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"device"},
	Short:   "Manage Spotify Connect devices",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		v, err := config.LoadConfig()
		if err != nil {
			log.Fatal("Error loading config file, Run `spotgo connect` to create a config file and connect to Spotify")
		}
		spotgoClient, err = config.SpotifyClient(context.Background(), v)
		if err != nil {
			log.Fatal("Error creating Spotify client, Run `spotgo connect` to connect to Spotify")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		lsCmd.Run(cmd, args)
	},
}

func init() {
	root.RootCmd.AddCommand(DevicesCmd)
}
