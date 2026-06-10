package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/Esteban-Bermudez/spotgo/config"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the background spotifyd daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		binPath, err := spotifydBinaryPath(runtime.GOOS, config.DataDir(), exec.LookPath, fileExists)
		if err != nil {
			return err
		}

		stateDir := config.StateDir()
		os.MkdirAll(stateDir, 0755)
		pidFile := filepath.Join(stateDir, "spotifyd.pid")

		if pid, running := runningPidFromFile(pidFile); running {
			return fmt.Errorf("spotifyd is already running (PID %d). Run `spotgo daemon stop` first", pid)
		}
		// A leftover pid file whose process is gone (e.g. spotifyd crashed)
		// must not block a fresh start.
		os.Remove(pidFile)

		spotifydArgs := spotifydStartArgs(runtime.GOOS, pidFile)
		c := exec.Command(binPath, spotifydArgs...)

		if runtime.GOOS == "darwin" {
			c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

			// Capture spotifyd's output; in foreground mode it otherwise goes to
			// /dev/null, leaving nothing to inspect when the session drops.
			logPath := filepath.Join(stateDir, "spotifyd.log")
			logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err == nil {
				defer logFile.Close()
				c.Stdout = logFile
				c.Stderr = logFile
			}

			if err := c.Start(); err != nil {
				return fmt.Errorf("failed to start spotifyd: %w", err)
			}

			pid := c.Process.Pid
			if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
				return fmt.Errorf("failed to write pid file: %w", err)
			}
			if err := c.Process.Release(); err != nil {
				return fmt.Errorf("failed to release spotifyd process: %w", err)
			}

			fmt.Print(spotifydStartedMessage(pid))
			return nil
		}

		if err := c.Run(); err != nil {
			return fmt.Errorf("failed to start spotifyd: %w", err)
		}

		fmt.Printf("Started spotifyd daemon in the background.\n")
		return nil
	},
}

func spotifydBinaryPath(goos string, dataDir string, lookPath func(string) (string, error), exists func(string) bool) (string, error) {
	localPath := filepath.Join(dataDir, "bin", "spotifyd")

	if goos == "darwin" {
		if path, err := lookPath("spotifyd"); err == nil {
			return path, nil
		}
		if exists(localPath) {
			return localPath, nil
		}
		return "", fmt.Errorf("spotifyd not found. Run `spotgo daemon install` first")
	}

	if exists(localPath) {
		return localPath, nil
	}
	if path, err := lookPath("spotifyd"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("spotifyd not found at %s. Run `spotgo daemon install` first", localPath)
}

func spotifydStartArgs(goos string, pidFile string) []string {
	if goos == "darwin" {
		return []string{"--no-daemon", "--backend", "portaudio", "--device-name", "spotgo"}
	}

	return []string{"--pid", pidFile, "--device-name", "spotgo"}
}

func spotifydStartedMessage(pid int) string {
	return fmt.Sprintf("Started spotifyd daemon in the background (PID %d).\n", pid)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func init() {
	DaemonCmd.AddCommand(startCmd)
}
