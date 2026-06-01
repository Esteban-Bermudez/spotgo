//go:build darwin

package player

import (
	"context"
	"log"
	"sync"
	"sync/atomic"

	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/macos"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/macos/mediaplayer"
	"github.com/progrium/darwinkit/objc"
	"github.com/zmb3/spotify/v2"
)

// MPNowPlayingInfo dictionary keys. DarwinKit doesn't export the
// MPMediaItemProperty*/MPNowPlayingInfoProperty* NSString constants, and
// SetNowPlayingInfo takes Go-string keys, so we use their verified runtime
// values (probed from the MediaPlayer framework). Title/Artist/Album are their
// plain names; elapsed/rate are the full symbol names.
const (
	npKeyTitle    = "title"
	npKeyArtist   = "artist"
	npKeyAlbum    = "albumTitle"
	npKeyDuration = "playbackDuration"
	npKeyElapsed  = "MPNowPlayingInfoPropertyElapsedPlaybackTime"
	npKeyRate     = "MPNowPlayingInfoPropertyPlaybackRate"
)

var (
	npCenter  mediaplayer.NowPlayingInfoCenter
	npPlaying atomic.Bool // last published playback state, read by the toggle key
	npStop    sync.Once
)

// runNowPlaying starts the NSApplication run loop so macOS registers spotgo as
// the Now Playing app (showing the track in Control Center) and delivers
// media-key commands, then runs worker (the player) on a goroutine. macos.RunApp
// locks the OS thread and must run on the main goroutine; the player code stays
// unaware it's no longer on main.
func runNowPlaying(client *spotify.Client, worker func()) {
	macos.RunApp(func(app appkit.Application, _ *appkit.ApplicationDelegate) {
		// Accessory: no Dock icon and no focus stealing, but still a real app so
		// Control Center surfaces our now-playing info.
		app.SetActivationPolicy(appkit.ApplicationActivationPolicyAccessory)
		npCenter = mediaplayer.NowPlayingInfoCenter_DefaultCenter()
		registerMediaKeys(client)
		go func() {
			worker()
			stopNowPlaying()
		}()
	})
}

func registerMediaKeys(client *spotify.Client) {
	cc := mediaplayer.RemoteCommandCenter_SharedCommandCenter()
	bindMediaKey(cc.PlayCommand(), func(ctx context.Context) error { return client.Play(ctx) })
	bindMediaKey(cc.PauseCommand(), func(ctx context.Context) error { return client.Pause(ctx) })
	bindMediaKey(cc.NextTrackCommand(), func(ctx context.Context) error { return client.Next(ctx) })
	bindMediaKey(cc.PreviousTrackCommand(), func(ctx context.Context) error { return client.Previous(ctx) })
	bindMediaKey(cc.TogglePlayPauseCommand(), func(ctx context.Context) error {
		if npPlaying.Load() {
			return client.Pause(ctx)
		}
		return client.Play(ctx)
	})
}

// bindMediaKey wires a remote command to a playback action. The action runs on
// its own goroutine so a slow Web API call never blocks the run loop; the
// handler reports success immediately.
func bindMediaKey(cmd mediaplayer.RemoteCommand, action func(context.Context) error) {
	cmd.SetEnabled(true)
	cmd.AddTargetWithHandler(func(_ mediaplayer.RemoteCommandEvent) mediaplayer.RemoteCommandHandlerStatus {
		go func() {
			if err := action(context.Background()); err != nil {
				log.Printf("nowplaying: media command failed: %v", err)
			}
		}()
		return mediaplayer.RemoteCommandHandlerStatusSuccess
	})
}

// updateNowPlaying pushes the current track and playback position to Control
// Center. The OS interpolates elapsed time from the playback rate, so publishing
// once per poll keeps the timeline in sync without extra API traffic.
func updateNowPlaying(state *spotify.PlayerState, elapsedMS int) {
	hasTrack := state != nil && state.Item != nil
	playing := hasTrack && state.Playing
	npPlaying.Store(playing)

	if !hasTrack {
		dispatch.MainQueue().DispatchAsync(func() {
			npCenter.SetNowPlayingInfo(map[string]objc.IObject{})
			npCenter.SetPlaybackState(mediaplayer.NowPlayingPlaybackStateStopped)
		})
		return
	}

	// Capture plain values so the closure doesn't read the shared state struct
	// from the main-queue goroutine.
	title := state.Item.Name
	artist := joinArtists(state.Item.Artists)
	album := state.Item.Album.Name
	durationSec := float64(state.Item.Duration) / 1000.0
	elapsedSec := float64(elapsedMS) / 1000.0
	rate := 0.0
	playbackState := mediaplayer.NowPlayingPlaybackStatePaused
	if playing {
		rate = 1.0
		playbackState = mediaplayer.NowPlayingPlaybackStatePlaying
	}

	dispatch.MainQueue().DispatchAsync(func() {
		npCenter.SetNowPlayingInfo(map[string]objc.IObject{
			npKeyTitle:    foundation.MutableString_StringWithString(title),
			npKeyArtist:   foundation.MutableString_StringWithString(artist),
			npKeyAlbum:    foundation.MutableString_StringWithString(album),
			npKeyDuration: foundation.Number_NumberWithDouble(durationSec),
			npKeyElapsed:  foundation.Number_NumberWithDouble(elapsedSec),
			npKeyRate:     foundation.Number_NumberWithDouble(rate),
		})
		npCenter.SetPlaybackState(playbackState)
	})
}

func stopNowPlaying() {
	npStop.Do(func() {
		dispatch.MainQueue().DispatchAsync(func() {
			appkit.Application_SharedApplication().Terminate(nil)
		})
	})
}
