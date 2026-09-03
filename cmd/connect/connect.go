package connect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/Esteban-Bermudez/spotgo/internal/session"
	"github.com/spf13/cobra"
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

	if session.HasStoredCredentials() {
		fmt.Println("Speaker already set up.")
		return
	}
	fmt.Println("Setting up the spotgo speaker (browser opens once)...")
	setupCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if err := session.Setup(setupCtx); err != nil {
		log.Fatal("Error setting up speaker:", err)
	}
	fmt.Println("Speaker ready.")
}
