//go:build darwin

// Package nowplaying integrates spotgo's player with the macOS "now playing"
// surface — the Control Center widget and the keyboard media keys — via a small
// Objective-C shim (see nowplaying_darwin.m). On non-macOS platforms the same
// functions are no-ops (nowplaying_other.go), so callers stay platform-agnostic.
package nowplaying

/*
#cgo LDFLAGS: -framework Cocoa -framework MediaPlayer
#include <stdlib.h>
#include "nowplaying_darwin.h"
*/
import "C"

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"unsafe"

	"github.com/zmb3/spotify/v2"
)

var (
	spotifyClient *spotify.Client
	workerFunc    func()
	playing       atomic.Bool // last published playback state, read by the toggle key
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
	go func() {
		var err error
		switch int(cmd) {
		case int(C.NP_CMD_PLAY):
			err = client.Play(context.Background())
		case int(C.NP_CMD_PAUSE):
			err = client.Pause(context.Background())
		case int(C.NP_CMD_TOGGLE):
			if playing.Load() {
				err = client.Pause(context.Background())
			} else {
				err = client.Play(context.Background())
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

// Update pushes the current track and playback position to Control Center. The
// OS interpolates elapsed time from the playback rate, so publishing once per
// poll keeps the timeline in sync without extra API traffic.
func Update(state *spotify.PlayerState, elapsedMS int) {
	hasTrack := state != nil && state.Item != nil
	isPlaying := hasTrack && state.Playing
	playing.Store(isPlaying)

	if !hasTrack {
		C.npClearNowPlaying()
		return
	}

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
