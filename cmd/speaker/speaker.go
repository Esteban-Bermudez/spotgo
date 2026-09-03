package speaker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/Esteban-Bermudez/spotgo/internal/nowplaying"
	"github.com/Esteban-Bermudez/spotgo/internal/playback"
	"github.com/Esteban-Bermudez/spotgo/internal/session"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zmb3/spotify/v2"
)

var spotgoClient *spotify.Client
var spotgoConfig *viper.Viper

var speakerCmd = &cobra.Command{
	Use:   "speaker",
	Short: "Run the embedded spotgo speaker (local playback)",
	Long: `Start the embedded Spotify Connect speaker that renders audio on this
machine, like the old spotifyd daemon but in-process. The player and other
commands control it over the Web API; run once via 'spotgo connect'.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		v, err := config.LoadConfig()
		if err != nil {
			log.Fatal("Error loading config file, Run `spotgo connect` to create a config file and connect to Spotify")
		}
		spotgoConfig = v
		spotgoClient, err = config.SpotifyClient(context.Background(), v)
		if err != nil {
			log.Fatal("Error creating Spotify client, Run `spotgo connect` to connect to Spotify")
		}
	},
	Run: runSpeaker,
}

func runSpeaker(cmd *cobra.Command, args []string) {
	background, _ := cmd.Flags().GetBool("background")
	if os.Getenv(session.ChildEnv) != "" {
		runSpeakerChild()
		return
	}
	if !session.HasStoredCredentials() {
		log.Fatal("speaker not set up yet — run `spotgo connect` first")
	}
	if background {
		pid, err := session.StartBackground()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("spotgo speaker running in the background (PID %d).\n", pid)
		return
	}
	runSpeakerForeground()
}

func runSpeakerForeground() {
	if pid, ok := session.BackgroundRunning(); ok {
		fmt.Printf("spotgo speaker already running in the background (PID %d).\n", pid)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sess, err := session.New(ctx, spotgoConfig)
	if err != nil {
		log.Fatalf("speaker: %v", err)
	}
	defer sess.Close()
	if err := session.WritePidFile(os.Getpid()); err != nil {
		log.Fatalf("speaker: %v", err)
	}
	defer session.RemovePidFile()
	go claimSpeaker()
	go watchShutdown(cancel)
	fmt.Println("spotgo speaker running — Ctrl+C to stop")
	runWithMediaKeys(sess, ctx)
}

func runSpeakerChild() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sess, err := session.New(ctx, spotgoConfig)
	if err != nil {
		log.Fatalf("speaker: %v", err)
	}
	defer sess.Close()
	defer session.RemovePidFile()
	go claimSpeaker()
	go watchShutdown(cancel)
	runWithMediaKeys(sess, ctx)
}

// watchShutdown cancels on SIGINT/SIGTERM and then force-exits: the
// embedded daemon can block past context cancellation, so without the
// exit the speaker would swallow `speaker stop` and live on.
func watchShutdown(cancel context.CancelFunc) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	cancel()
	time.Sleep(2 * time.Second)
	session.RemovePidFile()
	os.Exit(0)
}

// runWithMediaKeys runs the speaker session with the OS media controls
// attached: Control Center / media keys work while the speaker runs, with
// no TUI open. nowplaying.Run owns the main thread on darwin and runs the
// worker on a goroutine; elsewhere it just runs the worker directly.
func runWithMediaKeys(sess *session.Session, ctx context.Context) {
	nowplaying.Run(spotgoClient, func() {
		go publishLoop(ctx)
		if err := sess.Run(ctx); err != nil {
			log.Printf("speaker ended: %v", err)
		}
	})
}

// publishLoop mirrors the player TUI poll: every 2s the current Web API
// state is pushed to Control Center. The speaker is the single publisher
// while alive; the player yields to it (see player stateMsg).
func publishLoop(ctx context.Context) {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			state, err := spotgoClient.PlayerState(ctx)
			if err != nil || state == nil {
				continue
			}
			nowplaying.Update(state, int(state.Progress))
		}
	}
}

func claimSpeaker() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := playback.ClaimDevice(ctx, spotgoClient, session.DeviceName, 30*time.Second); err != nil {
		log.Printf("speaker: %v", err)
		return
	}
	log.Println("playback switched to spotgo")
}

func init() {
	root.RootCmd.AddCommand(speakerCmd)
	speakerCmd.Flags().Bool("background", false, "Run the speaker detached in the background (writes pid file, frees the terminal)")
}
