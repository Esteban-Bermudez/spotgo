package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the spotifyd daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		pidFile := filepath.Join(config.StateDir(), "spotifyd.pid")

		data, err := os.ReadFile(pidFile)
		if err != nil {
			fmt.Println("spotifyd daemon is stopped.")
			return nil
		}

		pidStr := strings.TrimSpace(string(data))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return fmt.Errorf("invalid pid in file: %w", err)
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Println("spotifyd daemon is stopped (stale pid file).")
			os.Remove(pidFile)
			return nil
		}

		// On Unix, FindProcess always succeeds. We must send signal 0 to check if it's alive.
		err = process.Signal(syscall.Signal(0))
		if err != nil {
			fmt.Println("spotifyd daemon is stopped (stale pid file).")
			os.Remove(pidFile)
			return nil
		}

		fmt.Printf("spotifyd daemon is running (PID %d).\n", pid)
		return nil
	},
}

func init() {
	DaemonCmd.AddCommand(statusCmd)
}
