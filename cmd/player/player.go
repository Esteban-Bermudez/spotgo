package player

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Esteban-Bermudez/spotgo/cmd/connect"
	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var playerCmd = &cobra.Command{
	Use:   "player",
	Short: "Connect to Spotify",
	Long:  `Connect to Spotify to receive now playing information`,
	Run:   spotifyPlayer,
}

func init() {
	root.RootCmd.AddCommand(playerCmd)

	playerCmd.Flags().BoolP("oneline", "o", false, "Output playback data on one line")
	playerCmd.Flags().BoolP("no-progress", "", false, "Do not include progress bar")
}

func spotifyPlayer(cmd *cobra.Command, args []string) {
	oneLine, _ := cmd.Flags().GetBool("one-line")
	noProgress, _ := cmd.Flags().GetBool("no-progress")

	token, err := connect.LoadOAuthToken()
	if err != nil {
		log.Fatal("Error loading token, Run `spotgo connect` to connect to Spotify")
	}
	client := spotify.New(connect.Auth.Client(context.Background(), token))

	if oneLine {
		oneLineOutput(client, noProgress)
	}

	p := bubbletea.NewProgram(model{
		client:        client,
		songTitle:     "No Song Playing",
		progress:      "00:00 / 00:00",
		playbackState: false,
	}, bubbletea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}
}

type model struct {
	client         *spotify.Client
	songTitle      string
	currentArtists string
	currentAlbum   string
	progress       string
	playbackState  bool
	width          int
	height         int
}

func (m model) Init() bubbletea.Cmd {
	return fetchSongInfo(m)
}

func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case bubbletea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, bubbletea.SetWindowTitle("spotgo")
	case songInfoMsg:
		m.songTitle = msg.title
		m.currentArtists = msg.artists
		m.currentAlbum = msg.album
		m.progress = msg.progress
		return m, fetchSongInfo(m)

	case bubbletea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, bubbletea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	// Define styles
	var style = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("2")). // Green
		Align(lipgloss.Center).AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center).
		Width(70).
		Height(20)

	content := fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s\n\nProgress: %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true).Render("spotgo"), // Bold High Intensity Green
		m.songTitle,
		m.currentArtists,
		m.currentAlbum,
		m.progress,
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, style.Render(content))
}

type songInfoMsg struct {
	title    string
	artists  string
	album    string
	progress string
}

func fetchSongInfo(m model) bubbletea.Cmd {
	return func() bubbletea.Msg {
		// Load OAuth token and create Spotify client
		playerState, err := m.client.PlayerState(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		if playerState.Item == nil {
			time.Sleep(5 * time.Second)
			return songInfoMsg{
				title:    "No Song Playing",
				artists:  "",
				album:    "",
				progress: "00:00 / 00:00",
			}
		}

		songTitle := playerState.Item.Name
		artists := ""
		for i, artist := range playerState.Item.Artists {
			if i == 0 {
				artists = artist.Name
			} else {
				artists = fmt.Sprintf("%s, %s", artists, artist.Name)
			}
		}
		album := playerState.Item.Album.Name
		progress := progressBar(playerState)

		return songInfoMsg{title: songTitle, artists: artists, album: album, progress: progress}
	}
}

func progressBar(playerState *spotify.PlayerState) string {
	return fmt.Sprintf("%02d:%02d / %02d:%02d",
		(playerState.Progress/1000)/60,
		(playerState.Progress/1000)%60,
		(playerState.Item.Duration/1000)/60,
		(playerState.Item.Duration/1000)%60)
}
