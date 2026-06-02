package queue

import (
	"context"
	"log"
	"strings"

	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var spotgoClient *spotify.Client

var queueCmd = &cobra.Command{
	Use:   "queue",
	Short: "View Queue, Queue a song, album or playlist from a Spotify URI",
	Long:  `View Queue, Queue a song, album or playlist from a Spotify URI. Example: spotgo queue spotify:track:6rqhFgbbKwnb9MLmUQDhG6`,
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
		if len(args) == 0 {
			lsCmd.Run(cmd, args)
			return
		}
		queueFromURI(cmd, args)
	},
}

func init() {
	root.RootCmd.AddCommand(queueCmd)
}

func queueFromURI(cmd *cobra.Command, args []string) {
	uri := args[0]

	if !strings.HasPrefix(uri, "spotify:track:") {
		log.Fatalf("Invalid URI: %s. Only Spotify track URIs are supported (e.g. spotify:track:6rqhFgbbKwnb9MLmUQDhG6)", uri)
	}

	err := spotgoClient.QueueSong(context.Background(), spotify.ID(strings.TrimPrefix(uri, "spotify:track:")))
	if err != nil {
		log.Fatalf("Error playing URI: %v", err)
	}

	log.Printf("Queued URI: %s", uri)
}
