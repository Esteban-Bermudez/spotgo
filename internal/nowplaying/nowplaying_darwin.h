#ifndef SPOTGO_NOWPLAYING_DARWIN_H
#define SPOTGO_NOWPLAYING_DARWIN_H

// Media-key command identifiers passed to the Go callback goMediaCommand.
enum np_command {
  NP_CMD_PLAY = 0,
  NP_CMD_PAUSE = 1,
  NP_CMD_TOGGLE = 2,
  NP_CMD_NEXT = 3,
  NP_CMD_PREV = 4,
};

// Playback states, matching MPNowPlayingPlaybackState.
enum np_state {
  NP_STATE_STOPPED = 0,
  NP_STATE_PLAYING = 1,
  NP_STATE_PAUSED = 2,
};

// npRun registers the media-key handlers and runs the NSApplication run loop. It
// blocks and MUST be called on the main thread. Once the app finishes launching
// it calls goAppReady(); each media-key press calls goMediaCommand(id).
void npRun(void);

// npStop terminates the app, ending the run loop.
void npStop(void);

// npSetNowPlaying publishes the current track to Control Center. Strings are
// UTF-8 and copied; state is one of enum np_state.
void npSetNowPlaying(const char *title, const char *artist, const char *album,
                     double durationSec, double elapsedSec, double rate,
                     int state);

// npClearNowPlaying clears the now-playing info (nothing playing).
void npClearNowPlaying(void);

#endif
