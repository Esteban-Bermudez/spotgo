package queue

import (
	"context"
	"fmt"
	"log"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List items in the queue",
	Run: func(cmd *cobra.Command, args []string) {
		queue, err := spotgoClient.GetQueue(context.Background())
		if err != nil {
			log.Fatalf("Error fetching queue: %v", err)
		}

		fmt.Printf("%-40s | %-35s | %-20s | %s\n", "TITLE", "ALBUM", "ARTIST", "URI")
		fmt.Println("--------------------------------------------------------------------------------------------------------------------------")

		// Add the currently playing track at the top of the list

		var tracks []spotify.FullTrack
		tracks = append([]spotify.FullTrack{queue.CurrentlyPlaying}, queue.Items...)

		for _, item := range tracks {
			artistNames := ""
			for i, artist := range item.SimpleTrack.Artists {
				if i > 0 {
					artistNames += ", "
				}
				artistNames += artist.Name
			}

			// Truncate values for display to keep the table aligned
			title := truncateForDisplay(item.SimpleTrack.Name, 40)
			album := truncateForDisplay(item.Album.Name, 35)
			artists := truncateForDisplay(artistNames, 20)

			fmt.Printf("%-40s | %-35s | %-20s | %s\n", title, album, artists, item.SimpleTrack.URI)
		}
	},
}

func init() {
	queueCmd.AddCommand(lsCmd)
}

// truncateForDisplay returns a rune-aware truncated string for display.
// If the input length (in runes) is greater than width, it returns the first
// width-3 runes followed by "...". If width is <= 3, it will hard-cut to
// width runes with no ellipsis.
func truncateForDisplay(s string, width int) string {
	if width <= 0 {
		return ""
	}
	rcount := utf8.RuneCountInString(s)
	if rcount <= width {
		return s
	}
	if width <= 3 {
		// Hard cut to width runes, no room for dots
		rs := []rune(s)
		if len(rs) <= width {
			return s
		}
		return string(rs[:width])
	}
	// Reserve 3 chars for dots
	rs := []rune(s)
	cut := width - 3
	if cut < 0 {
		cut = 0
	}
	if cut > len(rs) {
		cut = len(rs)
	}
	return string(rs[:cut]) + "..."
}
