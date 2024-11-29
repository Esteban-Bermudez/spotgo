package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

var playerCmd = &cobra.Command{
	Use:   "player",
	Short: "Connect to Spotify",
	Long:  `Connect to Spotify to receive now playing information`,
	Run:   spotifyPlayer,
}

func init() {
	rootCmd.AddCommand(playerCmd)

	playerCmd.Flags().BoolP("oneline", "o", false, "Output playback data on one line")
	playerCmd.Flags().BoolP("no-progress", "", false, "Do not include progress bar")
}

var songTitle = "No Song Playing"
var currentArtists = ""
var currentAlbum = ""

func spotifyPlayer(cmd *cobra.Command, args []string) {
	oneLine, _ := cmd.Flags().GetBool("one-line")
	noProgress, _ := cmd.Flags().GetBool("no-progress")

	token, err := loadOAuthToken()
	if err != nil {
		log.Fatal("Error loading token, Run `spotgo connect` to connect to Spotify")
	}

	if oneLine {
		inlineSongLoop(token, noProgress)
	}

	p := bubbletea.NewProgram(model{
		songTitle:      songTitle,
		currentArtists: currentArtists,
		currentAlbum:   currentAlbum,
		progress:       "00:00 / 00:00",
		playbackState:  false,
	})

	_, err = p.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func inlineSongLoop(token *oauth2.Token, noProgress bool) {
	client := spotify.New(auth.Client(context.Background(), token))
	playerState, err := client.PlayerState(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	output := ""

	if playerState.Item == nil {
		fmt.Print("󰝛")
		time.Sleep(5 * time.Second)
	} else if playerState.Item != nil && playerState.Playing {
		output = fmt.Sprintf("󰝚  %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
	} else {
		output = fmt.Sprintf("󰝚  %s - %s", playerState.Item.Name, playerState.Item.Artists[0].Name)
	}

	if playerState.Item != nil && !noProgress {
		output = fmt.Sprintf("%s | %s", output, progressBar(playerState))
	}

	fmt.Println(output)
	os.Exit(0)
}

type model struct {
	songTitle      string
	currentArtists string
	currentAlbum   string
	progress       string
	playbackState  bool
}

func (m model) Init() bubbletea.Cmd {
	return fetchSongInfo
}

func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case songInfoMsg:
		m.songTitle = msg.title
		m.currentArtists = msg.artists
		m.currentAlbum = msg.album
		m.progress = msg.progress
		return m, fetchSongInfo

	case bubbletea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, bubbletea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	// TODO add a viewport so it has a border and do not include that with tmux
	content := fmt.Sprintf("# spotgo\n\n%s\n\n%s\n\n%s\n\nProgress: %s", m.songTitle, m.currentArtists, m.currentAlbum, m.progress)
	rendered, _ := glamour.Render(content, "dark")
	return rendered
}

type songInfoMsg struct {
	title    string
	artists  string
	album    string
	progress string
}

func fetchSongInfo() bubbletea.Msg {
	// Load OAuth token and create Spotify client
	token, err := loadOAuthToken()
	if err != nil {
		log.Fatal("Error loading token, Run `spotgo connect` to connect to Spotify")
	}

	client := spotify.New(auth.Client(context.Background(), token))
	playerState, err := client.PlayerState(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	if playerState.Item == nil {
		time.Sleep(5 * time.Second)
		return songInfoMsg{title: "No Song Playing", artists: "", album: "", progress: "00:00 / 00:00"}
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

func progressBar(playerState *spotify.PlayerState) string {
	return fmt.Sprintf("%02d:%02d / %02d:%02d",
		(playerState.Progress/1000)/60,
		(playerState.Progress/1000)%60,
		(playerState.Item.Duration/1000)/60,
		(playerState.Item.Duration/1000)%60)
}
