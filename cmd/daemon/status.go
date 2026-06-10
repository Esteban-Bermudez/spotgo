package daemon

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the spotifyd daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		pidFile := filepath.Join(config.StateDir(), "spotifyd.pid")

		if pid, running := runningPidFromFile(pidFile); running {
			fmt.Printf("spotifyd daemon is running (PID %d).\n", pid)
			return nil
		}

		if _, err := os.Stat(pidFile); err == nil {
			fmt.Println("spotifyd daemon is stopped (stale pid file).")
			os.Remove(pidFile)
			return nil
		}

		fmt.Println("spotifyd daemon is stopped.")
		return nil
	},
}

func init() {
	DaemonCmd.AddCommand(statusCmd)
}
