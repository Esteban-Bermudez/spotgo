package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Esteban-Bermudez/spotgo/config"
)

// PidPath is where the detached background speaker records its pid,
// so a second `player` invocation can attach instead of starting a
// duplicate speaker on the same device ID.
func PidPath() string {
	return filepath.Join(config.StateDir(), "speaker.pid")
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
