package player

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/Esteban-Bermudez/spotgo/internal/nowplaying"
	"github.com/Esteban-Bermudez/spotgo/internal/playback"
	"github.com/Esteban-Bermudez/spotgo/internal/session"
	bubbletea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
)

var spotgoClient *spotify.Client

var playerCmd = &cobra.Command{
	Use:   "player",
	Short: "Show now playing information",
	Long:  `Show the current spotify playback session in a full screen terminal interface`,
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
	Run: spotifyPlayer,
}

func init() {
	root.RootCmd.AddCommand(playerCmd)

	playerCmd.Flags().BoolP("oneline", "o", false, "Output playback data on one line")
	playerCmd.Flags().BoolP("no-progress", "", false, "Do not include progress bar")
	playerCmd.Flags().
		IntP("scroll", "s", 0, "Scroll the output string if greater than n characters")
}

// maybeStartSpeaker prompts before the TUI takes over when no speaker is
// running. A yes starts it detached (the speaker claims the device itself);
// the player then acts as a remote. One-line mode never prompts.
func maybeStartSpeaker() {
	if !session.HasStoredCredentials() {
		return
	}
	if _, ok := session.BackgroundRunning(); ok {
		return
	}
	fmt.Print("The spotgo speaker is not running. Start it in the background? [y/N]: ")
	var answer string
	if _, err := fmt.Scanln(&answer); err != nil {
		return
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer != "y" && answer != "yes" {
		return
	}
	pid, err := session.StartBackground()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("spotgo speaker running in the background (PID %d).\n", pid)
}

func spotifyPlayer(cmd *cobra.Command, args []string) {
	oneLine, _ := cmd.Flags().GetBool("one-line")
	noProgress, _ := cmd.Flags().GetBool("no-progress")
	scroll, _ := cmd.Flags().GetInt("scroll")

	if !oneLine {
		maybeStartSpeaker()
	}

	// On macOS, nowplaying.Run owns the NSApplication run loop so the track shows
	// in Control Center and the media keys drive playback; the player runs inside
	// it on a goroutine. On other platforms it just calls the worker directly.
	if oneLine {
		nowplaying.Run(spotgoClient, func() { oneLineOutput(spotgoClient, noProgress, scroll) })
		return
	}

	nowplaying.Run(spotgoClient, func() {
		p := bubbletea.NewProgram(model{
			client: spotgoClient,
		}, bubbletea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	})
}

type model struct {
	client    *spotify.Client
	state     *spotify.PlayerState
	fetchedAt time.Time
	lastPoll  time.Time
	width     int
	height    int
}

func (m model) Init() bubbletea.Cmd {
	return bubbletea.Batch(pollState(m.client), tick())
}

// dueForPoll decides whether this render tick should also trigger an API poll:
// either the steady-state interval has elapsed, or the current track is
// expected to have ended (rate-limited by trackEndRepollGap).
func (m model) dueForPoll() bool {
	if time.Since(m.lastPoll) >= pollInterval {
		return true
	}
	return trackEnded(m.state, m.fetchedAt) && time.Since(m.lastPoll) >= trackEndRepollGap
}

func (m model) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
	switch msg := msg.(type) {
	case bubbletea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, bubbletea.SetWindowTitle("spotgo")

	case tickMsg:
		// Re-render on every tick (the progress bar is interpolated locally),
		// but only hit the Spotify API once per pollInterval — or as soon as the
		// current track is expected to have ended, so a natural song change
		// shows up promptly without raising the steady-state request rate.
		if m.dueForPoll() {
			m.lastPoll = time.Time(msg)
			return m, bubbletea.Batch(pollState(m.client), tick())
		}
		return m, tick()

	case stateMsg:
		// Transient errors (network blips, rate limiting) keep the last known
		// state on screen and retry on the next poll instead of crashing.
		if msg.err == nil {
			m.state = msg.state
			m.fetchedAt = msg.fetchedAt
			elapsed := 0
			if m.state.Item != nil {
				elapsed = interpolatedProgressMS(m.state, m.fetchedAt)
			}
			nowplaying.Update(m.state, elapsed)
		}
		return m, nil

	case bubbletea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, bubbletea.Quit

		case " ":
			// Control errors are non-fatal: re-poll and let the UI reflect the
			// real state rather than killing the player on a transient failure.
			if m.state != nil && m.state.Playing {
				_ = m.client.Pause(context.Background())
			} else {
				// Resume falls back to the spotgo daemon device when the
				// session has gone idle and nothing is active anymore.
				_ = playback.Resume(context.Background(), m.client)
			}
			return m, pollState(m.client)

		case "n":
			_ = m.client.Next(context.Background())
			return m, pollState(m.client)

		case "p":
			_ = m.client.Previous(context.Background())
			return m, pollState(m.client)

		case "-":
			if m.state != nil {
				newVolume := m.state.Device.Volume - 10
				newVolume = max(0, newVolume)
				_ = m.client.Volume(context.Background(), int(newVolume))
			}
			return m, pollState(m.client)

		case "+", "=":
			if m.state != nil {
				newVolume := m.state.Device.Volume + 10
				newVolume = min(100, newVolume)
				_ = m.client.Volume(context.Background(), int(newVolume))
			}
			return m, pollState(m.client)

		case "s":
			_ = m.client.Shuffle(context.Background(), !m.state.ShuffleState)
			return m, pollState(m.client)

		case "r":
			var newRepeat string
			switch m.state.RepeatState {
			case "off":
				newRepeat = "context"
			case "context":
				newRepeat = "track"
			case "track":
				newRepeat = "off"
			default:
				newRepeat = "off"
			}
			_ = m.client.Repeat(context.Background(), newRepeat)
			return m, pollState(m.client)
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
		Width(50).
		Height(10)

	songTitle := "No Song Playing"
	artists := ""
	album := ""
	progress := "00:00 // 00:00"
	playing := false
	var shuffle, repeat string

	if m.state != nil && m.state.Item != nil {
		songTitle = m.state.Item.Name
		artists = joinArtists(m.state.Item.Artists)
		album = m.state.Item.Album.Name
		progress = progressBar(interpolatedProgressMS(m.state, m.fetchedAt), int(m.state.Item.Duration))
		playing = m.state.Playing
		if m.state.ShuffleState {
			shuffle = ""
		} else {
			shuffle = " "
		}

		switch m.state.RepeatState {
		case "off":
			repeat = "󰑗"
		case "context":
			repeat = "󰑖"
		case "track":
			repeat = "󰑘"
		default:
			repeat = " "
		}
	}

	var icon string
	if playing {
		icon = ""
	} else {
		icon = ""
	}

	content := fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s |<| %s |>| %s\n%s",
		songTitle,
		artists,
		album,
		shuffle,
		icon,
		repeat,
		progress,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true).
			Align(lipgloss.Left).
			Width(50).
			Render("spotgo")+"\n"+
			style.Render(
				content),
	)
}

// pollInterval is how often the players actually call the Spotify API.
// renderInterval is how often the UI refreshes locally. Keeping the API poll
// well above the render rate is what keeps request volume low enough to avoid
// rate limiting, even with several players running at once.
const (
	pollInterval   = 2 * time.Second
	renderInterval = 1 * time.Second

	// trackEndRepollGap bounds how often the track-end predictor may fire, so we
	// don't poll repeatedly in the brief window between requesting a poll at a
	// track boundary and the fresh state arriving.
	trackEndRepollGap = 1 * time.Second
)

// trackEnded reports whether a playing track has reached (interpolated) its end,
// which is our cue that a new song has likely started.
func trackEnded(state *spotify.PlayerState, fetchedAt time.Time) bool {
	if state == nil || state.Item == nil || !state.Playing {
		return false
	}
	return interpolatedProgressMS(state, fetchedAt) >= int(state.Item.Duration)
}

// tickMsg drives the render loop; stateMsg carries a freshly polled player
// state (or the error from trying to fetch it).
type (
	tickMsg  time.Time
	stateMsg struct {
		state     *spotify.PlayerState
		fetchedAt time.Time
		err       error
	}
)

func tick() bubbletea.Cmd {
	return bubbletea.Tick(renderInterval, func(t time.Time) bubbletea.Msg {
		return tickMsg(t)
	})
}

func pollState(client *spotify.Client) bubbletea.Cmd {
	return func() bubbletea.Msg {
		state, err := client.PlayerState(context.Background())
		return stateMsg{state: state, fetchedAt: time.Now(), err: err}
	}
}

// interpolatedProgressMS advances the track progress by the wall-clock time
// elapsed since the state was fetched, so the progress bar stays live between
// API polls. Progress only advances while playing and never exceeds the track
// duration.
func interpolatedProgressMS(state *spotify.PlayerState, fetchedAt time.Time) int {
	if state == nil || state.Item == nil {
		return 0
	}
	progress := int(state.Progress)
	if state.Playing {
		progress += int(time.Since(fetchedAt).Milliseconds())
	}
	if duration := int(state.Item.Duration); progress > duration {
		progress = duration
	}
	return progress
}

func joinArtists(artists []spotify.SimpleArtist) string {
	names := ""
	for i, artist := range artists {
		if i == 0 {
			names = artist.Name
		} else {
			names = fmt.Sprintf("%s, %s", names, artist.Name)
		}
	}
	return names
}

func progressBar(progressMS, durationMS int) string {
	return fmt.Sprintf("%02d:%02d // %02d:%02d",
		(progressMS/1000)/60,
		(progressMS/1000)%60,
		(durationMS/1000)/60,
		(durationMS/1000)%60)
}
