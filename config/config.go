package config

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pkg/browser"
	"github.com/spf13/viper"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

func InitConfig() (*viper.Viper, error) {
	v, err := LoadConfig()
	if err == nil {
		return v, nil
	}

	missingConfig := errors.As(err, &viper.ConfigFileNotFoundError{})

	if missingConfig && v.ConfigFileUsed() == "" {
		err = setDefaultConfig(v) // Set default values with a login prompt
		if err != nil {
			return nil, fmt.Errorf("failed to set default config: %w", err)
		}
		v.SetConfigFile(os.ExpandEnv("$HOME/.config/spotgo/spotgo.json"))
		return v, nil
	} else if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	return v, nil
}

func RefreshAndSaveToken(v *viper.Viper) (*oauth2.Token, error) {
	t, err := getOAuthToken(v)
	if err != nil {
		return nil, fmt.Errorf("error getting OAuth token: %w", err)
	}

	if t.Expiry.Before(time.Now()) {
		auth := getAuthenticator()
		token, err := auth.RefreshToken(context.Background(), t)
		if err != nil {
			return nil, fmt.Errorf("error refreshing token: %w", err)
		}
		tokenMap, err := MarshalToken(token)
		if err != nil {
			return nil, fmt.Errorf("error marshaling token: %w", err)
		}
		v.Set("token", tokenMap)
		err = v.WriteConfig()
		if err != nil {
			return nil, fmt.Errorf("error writing config file: %w", err)
		}
		return token, nil
	}
	return t, nil
}

func getAuthenticator() *spotifyauth.Authenticator {
	clientID := "762ca48c26614d3389d0c1337bc65d48"
	auth := spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithRedirectURL("http://localhost:8679/callback"),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserModifyPlaybackState))

	return auth
}

func LoadConfig() (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName("spotgo")
	v.SetConfigType("json")
	v.AddConfigPath(os.ExpandEnv("$HOME/.config/spotgo/"))    // Default config directory
	v.AddConfigPath(os.ExpandEnv("$XDG_CONFIG_HOME/spotgo/")) // Optional: XDG config directory
	v.AddConfigPath(os.ExpandEnv("$HOME/.spotgo/"))           // Optional: Home directory

	err := v.ReadInConfig()
	if err != nil {
		return v, fmt.Errorf("error reading config file: %w", err)
	}
	_, err = RefreshAndSaveToken(v)
	if err != nil {
		return v, fmt.Errorf("error refreshing and saving token: %w", err)
	}
	return v, nil
}

func setDefaultConfig(v *viper.Viper) error {
	// login the user and create a token
	codeVerifier, err := generateRandomString(32)
	if err != nil {
		return fmt.Errorf("failed to generate random code verifier: %w", err)
	}
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])

	state, err := generateRandomString(16)
	if err != nil {
		return fmt.Errorf("failed to generate random state: %w", err)
	}

	auth := getAuthenticator()

	url := auth.AuthURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("state", state),
	)

	ch := make(chan *spotify.Client)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		t, err := auth.Token(r.Context(), state, r,
			oauth2.SetAuthURLParam("code_verifier", codeVerifier))
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, state)
		}
		client := spotify.New(auth.Client(r.Context(), t))
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "Login completed! You can now close this tab.")
		ch <- client
	})

	go func() {
		err := http.ListenAndServe(":8679", nil)
		if err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	fmt.Printf(
		"Please log in to Spotify by visiting the following page in your browser: %s\n\n",
		url,
	)

	browser.OpenURL(url)

	client := <-ch
	if client == nil {
		return fmt.Errorf("failed to get client from channel")
	}

	token, err := client.Token()
	if err != nil {
		return fmt.Errorf("failed to get token from client: %w", err)
	}
	tokenMap := map[string]any{
		"access_token":  token.AccessToken,
		"token_type":    token.TokenType,
		"refresh_token": token.RefreshToken,
		"expiry":        token.Expiry.Format(time.RFC3339),
	}
	// Set the token in the viper configuration
	v.Set("token", tokenMap)
	return v.SafeWriteConfig()
}

func getOAuthToken(v *viper.Viper) (*oauth2.Token, error) {
	tokenMap := v.GetStringMap("token")
	if tokenMap == nil {
		return nil, fmt.Errorf("no token found in config")
	}
	exp, err := time.Parse(time.RFC3339, tokenMap["expiry"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to parse token expiry: %w", err)
	}
	token := &oauth2.Token{
		AccessToken:  tokenMap["access_token"].(string),
		TokenType:    tokenMap["token_type"].(string),
		RefreshToken: tokenMap["refresh_token"].(string),
		Expiry:       exp,
	}
	return token, nil
}

func SpotifyClient(ctx context.Context, v *viper.Viper) (*spotify.Client, error) {
	token, err := getOAuthToken(v)
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth token: %w", err)
	}

	auth := getAuthenticator()
	client := spotify.New(auth.Client(ctx, token))
	if client == nil {
		return nil, fmt.Errorf("failed to create Spotify client")
	}
	return client, nil
}
