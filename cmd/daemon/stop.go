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

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background spotifyd daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		pidFile := filepath.Join(config.StateDir(), "spotifyd.pid")

		data, err := os.ReadFile(pidFile)
		if err != nil {
			return fmt.Errorf("could not read pid file (is the daemon running?): %w", err)
		}

		pidStr := strings.TrimSpace(string(data))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return fmt.Errorf("invalid pid in file: %w", err)
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			os.Remove(pidFile)
			return fmt.Errorf("could not find process %d", pid)
		}

		if err := process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("failed to stop process: %w", err)
		}

		os.Remove(pidFile)
		fmt.Printf("Stopped spotifyd daemon (PID %d)\n", pid)
		return nil
	},
}

func init() {
	DaemonCmd.AddCommand(stopCmd)
}
