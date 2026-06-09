//go:build darwin

// Package nowplaying integrates spotgo's player with the macOS "now playing"
// surface — the Control Center widget and the keyboard media keys — via a small
// Objective-C shim (see nowplaying_darwin.m). On non-macOS platforms the same
// functions are no-ops (nowplaying_other.go), so callers stay platform-agnostic.
package nowplaying

/*
#cgo CFLAGS: -fobjc-arc
#cgo LDFLAGS: -framework Cocoa -framework MediaPlayer
#include <stdlib.h>
#include "nowplaying_darwin.h"
*/
import "C"

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/Esteban-Bermudez/spotgo/internal/playback"
	"github.com/zmb3/spotify/v2"
)

var (
	spotifyClient *spotify.Client
	workerFunc    func()
	playing       atomic.Bool // last published playback state, read by the toggle key

	artClient = &http.Client{Timeout: 10 * time.Second}
	artMu     sync.Mutex
	artURL    string // album-art URL applied for the current track; guards stale fetches
)

// Run registers the keyboard media keys and runs the macOS run loop so the track
// shows in Control Center and the media keys drive playback. It blocks on the
// run loop (main thread); worker (the player) runs on a goroutine started from
// goAppReady once the app finishes launching.
func Run(client *spotify.Client, worker func()) {
	spotifyClient = client
	workerFunc = worker
	C.npRun()
}

//export goAppReady
func goAppReady() {
	go func() {
		workerFunc()
		C.npStop()
	}()
}

//export goMediaCommand
func goMediaCommand(cmd C.int) {
	client := spotifyClient
	if client == nil {
		return
	}
	// Run the Web API call off the run loop so a slow request never blocks it.
	// Resume (rather than Play) falls back to the spotgo daemon device when the
	// session has gone idle and Spotify reports no active device.
	go func() {
		var err error
		switch int(cmd) {
		case int(C.NP_CMD_PLAY):
			err = playback.Resume(context.Background(), client)
		case int(C.NP_CMD_PAUSE):
			err = client.Pause(context.Background())
		case int(C.NP_CMD_TOGGLE):
			if playing.Load() {
				err = client.Pause(context.Background())
			} else {
				err = playback.Resume(context.Background(), client)
			}
		case int(C.NP_CMD_NEXT):
			err = client.Next(context.Background())
		case int(C.NP_CMD_PREV):
			err = client.Previous(context.Background())
		}
		if err != nil {
			log.Printf("nowplaying: media command failed: %v", err)
		}
	}()
}

// Update pushes the current track and playback position to Control Center while
// Spotify is playing, and clears it otherwise so spotgo never holds the Now
// Playing slot when it isn't the thing actually making sound. The OS
// interpolates elapsed time from the playback rate, so publishing once per poll
// keeps the timeline in sync without extra API traffic.
func Update(state *spotify.PlayerState, elapsedMS int) {
	hasTrack := state != nil && state.Item != nil
	isPlaying := hasTrack && state.Playing
	playing.Store(isPlaying)

	// Only claim the OS Now Playing slot while Spotify is actually playing.
	// When it's paused, stopped, or idle we clear our entry so we don't overwrite
	// whatever else owns the media controls (a browser tab, Jellyfin, etc.).
	if !isPlaying {
		setArtwork("")
		C.npClearNowPlaying()
		return
	}

	setArtwork(bestArtworkURL(state.Item.Album.Images))

	title := C.CString(state.Item.Name)
	artist := C.CString(joinArtists(state.Item.Artists))
	album := C.CString(state.Item.Album.Name)
	defer C.free(unsafe.Pointer(title))
	defer C.free(unsafe.Pointer(artist))
	defer C.free(unsafe.Pointer(album))

	rate := 0.0
	pbState := C.int(C.NP_STATE_PAUSED)
	if isPlaying {
		rate = 1.0
		pbState = C.int(C.NP_STATE_PLAYING)
	}

	C.npSetNowPlaying(
		title, artist, album,
		C.double(float64(state.Item.Duration)/1000.0),
		C.double(float64(elapsedMS)/1000.0),
		C.double(rate),
		pbState,
	)
}

// setArtwork drives album-art updates. When the art URL changes it clears the
// previous cover immediately and downloads the new one in the background; an
// unchanged URL is a no-op so we don't re-download every poll.
func setArtwork(url string) {
	artMu.Lock()
	if url == artURL {
		artMu.Unlock()
		return
	}
	artURL = url
	artMu.Unlock()

	C.npClearArtwork()
	if url == "" {
		return
	}
	go fetchArtwork(url)
}

func fetchArtwork(url string) {
	data, err := downloadArtwork(url)
	if err != nil {
		log.Printf("nowplaying: artwork fetch failed: %v", err)
		return
	}
	if len(data) == 0 {
		return
	}

	// Hold the lock across the apply so a track that changed mid-download can't
	// have its (now stale) cover published.
	artMu.Lock()
	defer artMu.Unlock()
	if artURL != url {
		return
	}
	C.npSetArtwork(unsafe.Pointer(&data[0]), C.int(len(data)))
}

func downloadArtwork(url string) ([]byte, error) {
	resp, err := artClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Cap the read so a misbehaving server can't balloon memory; album art is
	// well under this.
	return io.ReadAll(io.LimitReader(resp.Body, 8<<20))
}

// bestArtworkURL returns the album-art URL to display. Spotify orders images
// largest-first; the largest looks crisp in Control Center and is fetched only
// once per album.
func bestArtworkURL(images []spotify.Image) string {
	if len(images) == 0 {
		return ""
	}
	return images[0].URL
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
