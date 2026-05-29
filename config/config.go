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

	if missingConfig {
		err = os.MkdirAll(ConfigDir(), 0700)
		if err != nil {
			return nil, fmt.Errorf("failed to create config directory: %w", err)
		}

		err = setDefaultConfig(v) // Set default values with a login prompt
		if err != nil {
			return nil, fmt.Errorf("failed to set default config: %w", err)
		}
		return v, nil
	} else if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	return v, nil
}

// saveToken persists token to the config file. The caller is responsible for
// holding the token lock (see lockToken) so concurrent writers don't clobber
// each other.
func saveToken(v *viper.Viper, token *oauth2.Token) error {
	tokenMap, err := MarshalToken(token)
	if err != nil {
		return fmt.Errorf("error marshaling token: %w", err)
	}
	v.Set("token", tokenMap)
	if err := v.WriteConfig(); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	return nil
}

// RefreshAndSaveToken refreshes the stored token if it has expired and persists
// the result. It is safe to call from multiple spotgo processes concurrently:
// it takes an exclusive lock and re-reads the token from disk before deciding
// to refresh, so a token another process just rotated is reused rather than
// refreshed again (which would invalidate it under Spotify's PKCE rotation).
func RefreshAndSaveToken(v *viper.Viper) (*oauth2.Token, error) {
	unlock, err := lockToken()
	if err != nil {
		return nil, err
	}
	defer unlock()

	// Re-read from disk: another instance may have refreshed (and rotated)
	// the token since this viper instance last loaded it.
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error re-reading config file: %w", err)
	}

	t, err := getOAuthToken(v)
	if err != nil {
		return nil, fmt.Errorf("error getting OAuth token: %w", err)
	}

	// Token.Valid() applies a small expiry buffer, so we don't refresh a token
	// that's about to expire mid-request either.
	if t.Valid() {
		return t, nil
	}

	auth := getAuthenticator()
	token, err := auth.RefreshToken(context.Background(), t)
	if err != nil {
		return nil, fmt.Errorf("error refreshing token: %w", err)
	}
	if err := saveToken(v, token); err != nil {
		return nil, err
	}
	return token, nil
}

// persistingTokenSource is an oauth2.TokenSource that refreshes expired tokens
// and writes the rotated token back to the config file. Spotify's
// Authorization Code + PKCE flow issues a new refresh token on every refresh
// and invalidates the previous one, so a refresh whose result is not persisted
// leaves a dead refresh token on disk — the cause of "invalid refresh token"
// after a long-running session. This source closes that gap and coordinates
// with other processes via the same lock used by RefreshAndSaveToken.
type persistingTokenSource struct {
	v    *viper.Viper
	auth *spotifyauth.Authenticator
	ctx  context.Context
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	unlock, err := lockToken()
	if err != nil {
		return nil, err
	}
	defer unlock()

	// Re-read from disk so we pick up a token another instance may have just
	// refreshed, instead of refreshing our own (now possibly stale) copy.
	if err := p.v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error re-reading config file: %w", err)
	}

	tok, err := getOAuthToken(p.v)
	if err != nil {
		return nil, err
	}
	if tok.Valid() {
		return tok, nil
	}

	newTok, err := p.auth.RefreshToken(p.ctx, tok)
	if err != nil {
		return nil, fmt.Errorf("error refreshing token: %w", err)
	}
	if err := saveToken(p.v, newTok); err != nil {
		return nil, err
	}
	return newTok, nil
}

func getAuthenticator() *spotifyauth.Authenticator {
	clientID := "762ca48c26614d3389d0c1337bc65d48"
	auth := spotifyauth.New(
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithRedirectURL("http://127.0.0.1:8679/callback"),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserModifyPlaybackState))

	return auth
}

func LoadConfig() (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName("spotgo")
	v.SetConfigType("json")

	v.AddConfigPath(ConfigDir())

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
	// Wrap our persisting source in a ReuseTokenSource so the in-memory access
	// token is reused for every API call and we only hit the lock + disk + a
	// network refresh when it actually expires. Crucially, when it does refresh,
	// persistingTokenSource writes the rotated refresh token back to disk —
	// unlike the default auth.Client, whose refreshes were lost.
	src := oauth2.ReuseTokenSource(token, &persistingTokenSource{v: v, auth: auth, ctx: ctx})
	client := spotify.New(oauth2.NewClient(ctx, src))
	if client == nil {
		return nil, fmt.Errorf("failed to create Spotify client")
	}
	return client, nil
}
