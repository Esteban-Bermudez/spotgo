package connect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/spf13/cobra"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

var (
	clientId = "762ca48c26614d3389d0c1337bc65d48" // Spotify client ID
	Auth     = spotifyauth.New(
		spotifyauth.WithClientID(clientId),
		spotifyauth.WithRedirectURL("http://localhost:8679/callback"),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserModifyPlaybackState))
)

var ConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to Spotify",
	Long:  `Connect to Spotify to receive now playing information`,
	Run:   connectToSpotify,
}

func init() {
	root.RootCmd.AddCommand(ConnectCmd)
}

func connectToSpotify(cmd *cobra.Command, args []string) {
	v, err := config.InitConfig()
	if err != nil {
		log.Fatal("Error initializing config: ", err)
	}
	fmt.Println("Using config file:", v.ConfigFileUsed())

	ctx := context.Background()
	spcli, err := config.SpotifyClient(ctx, v)
	if err != nil {
		log.Fatal("Error getting Spotify client:", err)
	}

	user, err := spcli.CurrentUser(ctx)
	if err != nil {
		log.Fatal("Error getting current user:", err)
	}
	fmt.Println("Connected to Spotify as:", user.DisplayName)
}

func LoadOAuthToken() (*oauth2.Token, error) {
	v, err := config.InitConfig()
	if err != nil {
		return nil, fmt.Errorf("error initializing config: %w", err)
	}

	tok := v.GetStringMap("token")
	if len(tok) == 0 {
		return nil, fmt.Errorf("token not found")
	}

	exp, err := time.Parse(time.RFC3339, tok["expiry"].(string))
	if err != nil {
		return nil, fmt.Errorf("error parsing token expiry: %w", err)
	}
	token := &oauth2.Token{
		AccessToken:  tok["access_token"].(string),
		TokenType:    tok["token_type"].(string),
		RefreshToken: tok["refresh_token"].(string),
		Expiry:       exp,
	}

	if token.Expiry.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return token, nil
}
