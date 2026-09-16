//go:build windows

package speaker

import (
	"fmt"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the spotgo speaker",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("speaker stop is not supported on Windows; stop the foreground `spotgo speaker` with Ctrl+C.")
	},
}

func init() {
	speakerCmd.AddCommand(stopCmd)
}
