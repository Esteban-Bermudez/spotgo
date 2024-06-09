package cmd

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

  "github.com/adrg/xdg"
	"github.com/mozillazg/request"
	"github.com/spf13/cobra"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

// redirectURI is the OAuth redirect URI for the application.
// You must register an application at Spotify's developer portal
// and enter this value.
const redirectURI = "http://localhost:8080/callback"


// clientId is the client ID for the application. You must register
// an application at Spotify's developer portal and enter this value.
//
// clientID might not be confidential and this application will be easier to run
// if it does not require the user to get their own clientID.
var (
	clientId = os.Getenv("SPOTIFY_CLIENT_ID")
	auth     = spotifyauth.New(
		spotifyauth.WithClientID(clientId),
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadPlaybackState,
			spotifyauth.ScopeUserReadCurrentlyPlaying,
			spotifyauth.ScopeUserModifyPlaybackState))
	tokenCh       = make(chan *oauth2.Token)
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to Spotify",
	Long:  `Connect to Spotify to receive now playing information`,
	Run:   connectToSpotify,
}

func init() {
	rootCmd.AddCommand(connectCmd)
}

func connectToSpotify(cmd *cobra.Command, args []string) {
	token, err := loadOAuthToken()

	if err != nil {
		if err.Error() == "Token not found" {
			login()
			token = <-tokenCh
		} else if !token.Valid() {
			fmt.Println("Refreshing token...")
			token = updateToken(token)
		} else {
			log.Fatal(err)
		}
	}
	saveOAuthToken(token)
	client := spotify.New(auth.Client(context.Background(), token))
	user, err := client.CurrentUser(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("You are logged in as:", user.DisplayName)

}

func generateRandomString(size int) (string, error) {
	possible := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	values := make([]byte, size)
	_, err := rand.Read(values)
	if err != nil {
		return "", err
	}
	for i, b := range values {
		values[i] = possible[int(b)%len(possible)]
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(values), nil
}

func hashSha256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return string(bs)
}

func login() {
	codeVerifier, err := generateRandomString(32)
	if err != nil {
		log.Fatal(err)
	}
	codeChallenge := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(hashSha256(codeVerifier)))
	state, err := generateRandomString(16)
	http.HandleFunc("/callback", completeAuth(state, codeVerifier))
	go http.ListenAndServe(":8080", nil)

	url := auth.AuthURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
	)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
}

func completeAuth(state string, codeVerifier string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.Token(r.Context(), state, r,
			oauth2.SetAuthURLParam("code_verifier", codeVerifier))
		if err != nil {
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			log.Fatal(err)
		}
		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			log.Fatalf("State mismatch: %s != %s\n", st, state)
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "Login completed! You can now close this tab.")
		tokenCh <- tok
	}
}
func loadOAuthToken() (*oauth2.Token, error) {
  tokenFile, _ := xdg.SearchConfigFile("spotgo/token.json")
	jsonToken, err := os.ReadFile(tokenFile) 
	if err != nil {
		return nil, fmt.Errorf("Token not found")
	}

	var token oauth2.Token
	if err := json.Unmarshal(jsonToken, &token); err != nil {
		return nil, err
	}

	if token.Expiry.Before(time.Now()) {
		return &token, fmt.Errorf("Token expired")
	}

	return &token, nil

}

func saveOAuthToken(token *oauth2.Token) error {
  tokenFilePath, err := xdg.ConfigFile("spotgo/token.json")
	if err != nil {
		log.Fatal(err)
	}

	jsonToken, err := json.Marshal(token)
	if err != nil {
		return err
	}
	fmt.Println("Saving token to", tokenFilePath)
	return os.WriteFile(tokenFilePath , jsonToken, 0600)
}

func updateToken(token *oauth2.Token) *oauth2.Token {
	refreshToken := token.RefreshToken
	if refreshToken == "" {
		log.Fatal("No refresh token found")
	}

	c := &http.Client{}
	req := request.NewArgs(c)
	req.Headers = map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	req.Data = map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     clientId,
	}

	resp, err := request.Post("https://accounts.spotify.com/api/token", req)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Status)

	content, err := resp.Content()
	if err != nil {
		log.Fatal(err)
	}

	var newToken oauth2.Token
	newToken.Expiry = time.Now().Add(time.Duration(3600) * time.Second)
	if err := json.Unmarshal(content, &newToken); err != nil {
		log.Fatal(err)
	}
	return &newToken

}
