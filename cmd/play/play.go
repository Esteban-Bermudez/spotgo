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
		playTrackInAlbumContext(uri)
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

// playTrackInAlbumContext plays a track inside its album, starting at the
// track. A bare uris play ships the device an empty-context command that the
// embedded speaker (go-librespot) cannot resolve into a track list, so
// Spotify answers 403 Restriction violated. An album context plus offset
// carries the same intent in a shape every Connect device understands, and
// queues the rest of the album behind the track like tapping it in the
// official apps.
func playTrackInAlbumContext(uri spotify.URI) {
	ctx := context.Background()
	id := spotify.ID(strings.TrimPrefix(string(uri), "spotify:track:"))
	track, err := spotgoClient.GetTrack(ctx, id)
	if err != nil {
		log.Fatalf("Error looking up track %s: %v", uri, err)
	}
	albumURI := track.Album.URI
	err = spotgoClient.PlayOpt(ctx, &spotify.PlayOptions{
		PlaybackContext: &albumURI,
		PlaybackOffset:  &spotify.PlaybackOffset{URI: uri},
	})
	if err != nil {
		log.Fatalf("Error playing track URI: %v", err)
	}
	log.Printf("Playing Track : %s", uri)
}
