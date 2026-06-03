package queue

import (
	"context"
	"fmt"
	"log"

	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var spotgoClient *spotify.Client

// SearchTypeAlbum    SearchType = 1 << iota
// SearchTypeArtist              = 1 << iota
// SearchTypePlaylist            = 1 << iota
// SearchTypeTrack               = 1 << iota
// SearchTypeShow                = 1 << iota
// SearchTypeEpisode             = 1 << iota

var searchCmd = &cobra.Command{
	Use:   "search [type] [query]",
	Short: "Search for tracks, artists, albums, playlists, shows, or episodes",
	Long:  `Search for tracks, artists, albums, playlists, shows, or episodes.  The search type can be one of "album", "artist", "playlist", "track", "show", or "episode".  Multiple types can be searched simultaneously by separating them with commas (e.g. "track,artist").  If no type is specified, it defaults to "track".`,
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
	Run: spotifySearch,
}

func init() {
	root.RootCmd.AddCommand(searchCmd)
}

func searchTypeFromString(s string) spotify.SearchType {
	switch s {
	case "album":
		return spotify.SearchTypeAlbum
	case "artist":
		log.Fatal("Error: Searching for artists is not currently supported in spotgo")
		return 0
	case "playlist":
		return spotify.SearchTypePlaylist
	case "track":
		return spotify.SearchTypeTrack
	case "show":
		log.Fatal("Error: Searching for shows is not currently supported in spotgo")
		return 0
	case "episode":
		log.Fatal("Error: Searching for episodes is not currently supported in spotgo")
		return 0
	default:
		log.Fatalf("Error: Invalid search type '%s'. Valid types are: album, artist, playlist, track, show, episode", s)
		return 0
	}
}

func spotifySearch(cmd *cobra.Command, args []string) {
	searchType := "track"
	searchQuery := ""
	if len(args) > 1 {
		searchType = args[0]
		searchQuery = args[1]
	} else if len(args) == 1 {
		searchQuery = args[0]
	} else {
		log.Fatal("Error: No search query provided. Usage: spotgo search [type] [query]")
	}


	search, err := spotgoClient.Search(context.Background(), searchQuery, searchTypeFromString(searchType))
	if err != nil {
		log.Fatalf("Error fetching queue: %v", err)
	}

	if search.Tracks != nil {
		fmt.Printf("%-40s | %-35s | %-20s | %s\n", "TITLE", "ALBUM", "ARTIST", "URI")
		fmt.Println("--------------------------------------------------------------------------------------------------------------------------")
		for _, item := range search.Tracks.Tracks {
			artistNames := ""
			for i, artist := range item.SimpleTrack.Artists {
				if i > 0 {
					artistNames += ", "
				}
				artistNames += artist.Name
			}

			artists := artistNames
			album := item.Album.Name
			title := item.SimpleTrack.Name

			// Truncate values for display to keep the table aligned
			// title = truncateForDisplay(title, 40)
			// album = truncateForDisplay(album, 35)
			// artists = truncateForDisplay(artists, 20)

			fmt.Printf("%-40s | %-35s | %-20s | %s\n", title, album, artists, item.SimpleTrack.URI)
		}
	}

	if search.Albums != nil {
		fmt.Printf("%-40s | %-35s | %s\n", "ALBUM", "ARTIST", "URI")
		fmt.Println("--------------------------------------------------------------------------------------------------------------------------")
		for _, item := range search.Albums.Albums {
			artistNames := ""
			for i, artist := range item.Artists {
				if i > 0 {
					artistNames += ", "
				}
				artistNames += artist.Name
			}

			artists := artistNames
			album := item.Name

			// Truncate values for display to keep the table aligned
			// album = truncateForDisplay(album, 40)
			// artists = truncateForDisplay(artists, 35)

			fmt.Printf("%-40s | %-35s | %s\n", album, artists, item.URI)
		}
	}	

	if search.Playlists != nil {
		fmt.Printf("%-40s | %-35s | %s\n", "PLAYLIST", "OWNER", "URI")
		fmt.Println("--------------------------------------------------------------------------------------------------------------------------")
		for _, item := range search.Playlists.Playlists {
			owner := item.Owner.DisplayName
			playlist := item.Name

			// Truncate values for display to keep the table aligned
			// playlist = truncateForDisplay(playlist, 40)
			// owner = truncateForDisplay(owner, 35)

			fmt.Printf("%-40s | %-35s | %s\n", playlist, owner, item.URI)
		}
	}
}
