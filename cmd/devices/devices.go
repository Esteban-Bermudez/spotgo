package devices

import (
	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	"github.com/spf13/cobra"
)

var DevicesCmd = &cobra.Command{
	Use:     "devices",
	Aliases: []string{"device"},
	Short:   "Manage Spotify Connect devices",
}

func init() {
	root.RootCmd.AddCommand(DevicesCmd)
}