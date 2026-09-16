# spotgo
[![Go Report Card](https://goreportcard.com/badge/github.com/Esteban-Bermudez/spotgo)](https://goreportcard.com/report/github.com/Esteban-Bermudez/spotgo)
![GitHub Release](https://img.shields.io/github/v/release/Esteban-Bermudez/spotgo)

![Example](./assets/player.gif)

A CLI tool for Spotify written in Go

This project is a work in progress. The goal is to create a CLI tool for Spotify that allows users to interact with their currently playing music from the command line.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/Esteban-Bermudez/spotgo/main/install.sh | sh
```

This downloads the latest release for your OS/arch into `$HOME/.local/bin` (override with `INSTALL_DIR`).
On Linux, install the ALSA runtime library (the audio codecs are statically linked):
```bash
sudo apt-get install libasound2
```
Add it to your `$PATH` if needed:
```bash
export PATH="$HOME/.local/bin:$PATH"
```

Alternatives:

- Go: `go install github.com/Esteban-Bermudez/spotgo/cmd/spotgo@latest`
- Or grab a binary from the [releases page](https://github.com/Esteban-Bermudez/spotgo/releases)

Set your Spotify `client_id`:
```bash
export SPOTIFY_CLIENT_ID=<YOUR_CLIENT_ID>
```

Then connect:
```bash
spotgo connect
```

## Usage
```bash
spotgo [command]
```

Full command reference: [docs/COMMANDS.md](docs/COMMANDS.md)

| Command | Description |
|---------|-------------|
| `connect` | Authenticate with Spotify (also sets up the speaker) |
| `player` | Now-playing TUI; sub-commands `play`, `pause`, `toggle`, `next`, `previous`, `volume`, `shuffle`, `repeat` |
| `play <uri>` | Play a track, album, or playlist from a Spotify URI |
| `queue [uri]` | List the queue, or append a track URI |
| `search [type] [query]` | Search tracks, albums, or playlists |
| `devices` | List Connect devices; `devices switch <id-or-name>` transfers playback |
| `speaker` | Run the embedded local-playback speaker (`--background`, `status`, `stop`) |
