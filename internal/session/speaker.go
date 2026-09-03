package session

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/Esteban-Bermudez/spotgo/config"
)

// PidPath is where the detached background speaker records its pid,
// so a second `player` invocation can attach instead of starting a
// duplicate speaker on the same device ID.
func PidPath() string {
	return filepath.Join(config.StateDir(), "speaker.pid")
}

// runningPid returns the pid in path when that process is still alive.
// Missing, malformed, or stale files yield false.
func runningPid(path string) (int, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}
	return pid, proc.Signal(syscall.Signal(0)) == nil
}

// BackgroundRunning reports whether a detached speaker is currently alive.
func BackgroundRunning() (int, bool) {
	return runningPid(PidPath())
}

// WritePidFile records pid for later attach/stop.
func WritePidFile(pid int) error {
	if err := os.MkdirAll(config.StateDir(), 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	return os.WriteFile(PidPath(), []byte(strconv.Itoa(pid)), 0644)
}

// RemovePidFile clears the pid file. Stale files are removed silently.
func RemovePidFile() {
	os.Remove(PidPath())
}

// ChildEnv is set on the detached speaker child so it runs headless
// instead of spawning a grandchild.
const ChildEnv = "SPOTGO_SPEAKER_CHILD"

// StartBackground detaches `spotgo speaker --background` as a daemon: stdio
// goes to the librespot log, the pid is recorded, and the caller returns
// immediately with the terminal free. A duplicate start reports the
// existing pid as an error.
func StartBackground() (int, error) {
	if pid, ok := BackgroundRunning(); ok {
		return pid, fmt.Errorf("speaker already running (PID %d)", pid)
	}
	RemovePidFile()
	exe, err := os.Executable()
	if err != nil {
		return 0, fmt.Errorf("find executable: %w", err)
	}
	logFile, err := os.OpenFile(LogPath(), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer logFile.Close()
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return 0, err
	}
	defer devNull.Close()
	child := exec.Command(exe, "speaker", "--background")
	child.Env = append(os.Environ(), ChildEnv+"=1")
	child.Stdin = devNull
	child.Stdout = logFile
	child.Stderr = logFile
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := child.Start(); err != nil {
		return 0, fmt.Errorf("start speaker: %w", err)
	}
	pid := child.Process.Pid
	if err := WritePidFile(pid); err != nil {
		return 0, err
	}
	_ = child.Process.Release()
	return pid, nil
}
