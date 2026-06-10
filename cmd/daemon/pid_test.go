package daemon

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func writePidFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spotifyd.pid")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunningPidFromFile(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		if _, ok := runningPidFromFile(filepath.Join(t.TempDir(), "nope.pid")); ok {
			t.Error("missing pid file reported as running")
		}
	})

	t.Run("malformed pid", func(t *testing.T) {
		if _, ok := runningPidFromFile(writePidFile(t, "not-a-pid\n")); ok {
			t.Error("malformed pid file reported as running")
		}
	})

	t.Run("live process", func(t *testing.T) {
		// Our own pid is guaranteed alive.
		path := writePidFile(t, itoa(os.Getpid()))
		pid, ok := runningPidFromFile(path)
		if !ok || pid != os.Getpid() {
			t.Errorf("got (%d, %v), want (%d, true)", pid, ok, os.Getpid())
		}
	})

	t.Run("stale pid of exited process", func(t *testing.T) {
		// Spawn a process that exits immediately; its pid is then dead — the
		// exact state a crashed spotifyd leaves behind, which used to block
		// `daemon start`.
		cmd := exec.Command("true")
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
		path := writePidFile(t, itoa(cmd.Process.Pid))
		if _, ok := runningPidFromFile(path); ok {
			t.Error("dead process reported as running")
		}
	})
}
