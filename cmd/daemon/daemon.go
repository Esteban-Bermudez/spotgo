package daemon

import (
	"github.com/Esteban-Bermudez/spotgo/cmd/root"
	"github.com/spf13/cobra"
)

var DaemonCmd = &cobra.Command{
	Use:     "daemon",
	Aliases: []string{"d"},
	Short:   "Manage background spotifyd instance",
}

func init() {
	root.RootCmd.AddCommand(DaemonCmd)
}
