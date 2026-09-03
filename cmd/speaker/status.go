package speaker

import (
	"fmt"

	"github.com/Esteban-Bermudez/spotgo/internal/session"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check whether the spotgo speaker is running",
	Run: func(cmd *cobra.Command, args []string) {
		if pid, ok := session.BackgroundRunning(); ok {
			fmt.Printf("spotgo speaker is running (PID %d).\n", pid)
			return
		}
		fmt.Println("spotgo speaker is stopped.")
	},
}

func init() {
	speakerCmd.AddCommand(statusCmd)
}
