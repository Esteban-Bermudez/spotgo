# spotgo
[![Go Report Card](https://goreportcard.com/badge/github.com/Esteban-Bermudez/spotgo)](https://goreportcard.com/report/github.com/Esteban-Bermudez/spotgo)
![GitHub Release](https://img.shields.io/github/v/release/Esteban-Bermudez/spotgo)

a CLI tool for Spotify written in Go

This project is a work in progress. The goal is to create a CLI tool for Spotify that allows users to interact with their currently playing music from the command line.

## Installation
```bash
go install github.com/Esteban-Bermudez/spotgo@latest
```
Or install the binary from the [releases page](https://github.com/Esteban-Bermudez/spotgo/releases)

Once the binary is installed, you can add it to your `$PATH` for easy access.
Make sure to also include your Spotify `client_id` as an environment variable.
```bash
export SPOTIFY_CLIENT_ID=<YOUR_CLIENT_ID>
```

## Usage
```bash
spotgo [command]
```
### Commands
- `connect` - Connect to a Spotify account
- `player` - View current Spotify playing data
    - `--inline` - show playing data in one line (made for tmux)
