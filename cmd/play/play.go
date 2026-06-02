package play

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

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "play a song, album or playlist from uri",
	Long:  `play a song, album or playlist from uri, e.g. spotify:track:6rqhFgbbKwnb9MLmUQDhG6`,
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
	Run: playFromURI,
}

func init() {
	root.RootCmd.AddCommand(playCmd)
}

func playFromURI(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		log.Fatal("Please provide a Spotify URI to play, e.g. spotify:track:6rqhFgbbKwnb9MLmUQDhG6")
	}

	uri := spotify.URI(args[0])

	if strings.HasPrefix(string(uri), "spotify:track:") {
		err := spotgoClient.PlayOpt(context.Background(), &spotify.PlayOptions{
			URIs: []spotify.URI{uri},
		})
		if err != nil {
			log.Fatalf("Error playing track URI: %v", err)
		}
		log.Printf("Playing Track : %s", uri)
		return
	}

	err := spotgoClient.PlayOpt(context.Background(), &spotify.PlayOptions{
		PlaybackContext: &uri,
	})
	if err != nil {
		log.Fatalf("Error playing URI: %v", err)
	}

	log.Printf("Playing URI: %s", uri)
}
