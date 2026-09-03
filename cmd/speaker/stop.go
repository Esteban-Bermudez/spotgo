package speaker

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/Esteban-Bermudez/spotgo/internal/session"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the spotgo speaker",
	Run: func(cmd *cobra.Command, args []string) {
		pid, ok := session.BackgroundRunning()
		if !ok {
			session.RemovePidFile()
			fmt.Println("spotgo speaker is not running.")
			return
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			session.RemovePidFile()
			fmt.Println("spotgo speaker is not running (stale pid file removed).")
			return
		}
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			fmt.Printf("failed to stop speaker (PID %d): %v\n", pid, err)
			return
		}
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if err := proc.Signal(syscall.Signal(0)); err != nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if err := proc.Signal(syscall.Signal(0)); err == nil {
			_ = proc.Signal(syscall.SIGKILL)
		}
		session.RemovePidFile()
		fmt.Printf("Stopped spotgo speaker (PID %d).\n", pid)
	},
}

func init() {
	speakerCmd.AddCommand(stopCmd)
}
