# spotgo commands

Full command reference for `spotgo`. Run `spotgo [command] --help` for flags on any command.

First time setup: [`spotgo connect`](#connect), then optionally [`spotgo speaker`](#speaker) for local playback.

- [`connect`](#connect)
- [`player`](#player) — TUI plus `play`, `pause`, `toggle`, `next`, `previous`, `volume`, `shuffle`, `repeat`
- [`play <uri>`](#play-uri) — play a Spotify URI
- [`queue [uri]` / `queue ls`](#queue-uri--queue-ls)
- [`search [type] [query]`](#search-type-query)
- [`devices` / `devices ls` / `devices switch`](#devices--devices-ls--devices-switch)
- [`speaker` / `speaker status` / `speaker stop`](#speaker--speaker-status--speaker-stop)

## `connect`

```bash
spotgo connect
```

Opens a browser window to authenticate with Spotify. On success the speaker
credentials are set up as well, so local playback is ready.

Requires the `SPOTIFY_CLIENT_ID` environment variable:

```bash
export SPOTIFY_CLIENT_ID=<YOUR_CLIENT_ID>
```

## `player`

```bash
spotgo player [--one-line] [--no-progress] [-s|--scroll n]
spotgo player [sub-command]
```

With no sub-command, opens the full-screen now-playing TUI. If the
[embedded speaker](#speaker--speaker-status--speaker-stop) isn't running, the
player offers to start it in the background, then acts as a remote.

TUI keybinds:

| Key      | Action                        |
|----------|-------------------------------|
| `space`  | Play / pause                  |
| `n`      | Next track                    |
| `p`      | Previous track                |
| `-`/`+`  | Volume down / up (10%)        |
| `s`      | Toggle shuffle                |
| `r`      | Cycle repeat off → context → track |
| `q`, `ctrl+c` | Quit                     |

Flags (TUI and one-line mode):

- `-o, --one-line` — single-line output for status bars (never prompts to start the speaker)
- `--no-progress` — hide the progress bar in one-line mode
- `-s, --scroll int` — scroll the one-line output when longer than n characters

### `player play` / `player pause` / `player toggle`

```bash
spotgo player play
spotgo player pause
spotgo player toggle
```

Start/resume, pause, or toggle the current playback session. `toggle` resumes
via the spotgo speaker when the session has gone idle with no active device.

### `player next` / `player previous`

```bash
spotgo player next
spotgo player previous
```

Skip to the next or previous track in the current playback session.

### `player volume [0-100]`

```bash
spotgo player volume
spotgo player volume 75
```

With no argument, prints the current device volume. With an argument, sets it
to a level from 0 to 100.

### `player shuffle`

```bash
spotgo player shuffle
```

Toggles shuffle mode for the current playback session.

### `player repeat`

```bash
spotgo player repeat
```

Cycles repeat mode: `off` → `context` (repeat playlist) → `track` → `off`.

## `play <uri>`

```bash
spotgo play spotify:track:6rqhFgbbKwnb9MLmUQDhG6
spotgo play spotify:album:...
spotgo play spotify:playlist:...
```

Plays a song, album, or playlist from a Spotify URI. A track URI is played
inside its album context (starting at that track) so every Connect device —
including the embedded speaker — resolves it, with the rest of the album
queued behind it.

## `queue [uri]` / `queue ls`

```bash
spotgo queue
spotgo queue ls
spotgo queue spotify:track:6rqhFgbbKwnb9MLmUQDhG6
```

With no argument, lists the queue (currently playing track first). With a
track URI, appends that track to the queue. Only `spotify:track:` URIs can be
queued. `queue ls` always lists.

## `search [type] [query]`

```bash
spotgo search "midnight city"
spotgo search track "midnight city"
spotgo search album "random access memories"
spotgo search playlist "road trip"
```

Searches Spotify and prints a table with a URI column you can pass to
[`play`](#play-uri) or [`queue`](#queue-uri--queue-ls). Defaults to `track`
when no type is given. Supported types: `album`, `playlist`, `track`
(`artist`, `show`, and `episode` are not currently supported).

## `devices` / `devices ls` / `devices switch`

```bash
spotgo devices
spotgo devices ls
spotgo devices switch <device-id-or-name>
```

`devices` (alias: `device`, defaults to `ls`) lists available Spotify Connect
devices with name, type, active/restricted flags, and ID. `switch` transfers
playback to a device matched by ID or case-insensitive name substring, e.g.:

```bash
spotgo devices switch spotgo
```

## `speaker` / `speaker status` / `speaker stop`

```bash
spotgo speaker
spotgo speaker --background
spotgo speaker status
spotgo speaker stop
```

Runs the embedded Spotify Connect speaker — local audio playback in-process,
no separate daemon. Other commands control it over the Web API. Requires a
prior `spotgo connect` (speaker credentials are provisioned there).

- `spotgo speaker` — run in the foreground (Ctrl+C to stop). Owns OS media
  keys / Control Center while alive; the player TUI yields to it and acts as
  a pure remote.
- `spotgo speaker --background` — run detached, freeing the terminal.
- `spotgo speaker status` — report whether the background speaker is running.
- `spotgo speaker stop` — stop the background speaker.
