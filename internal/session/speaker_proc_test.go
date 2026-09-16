//go:build !windows

package session

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func writePidFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "speaker.pid")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunningPid(t *testing.T) {
	t.Run("missing file", func(t *testing.T) {
		if _, ok := runningPid(filepath.Join(t.TempDir(), "nope.pid")); ok {
			t.Error("missing pid file reported as running")
		}
	})

	t.Run("malformed pid", func(t *testing.T) {
		if _, ok := runningPid(writePidFile(t, "not-a-pid\n")); ok {
			t.Error("malformed pid file reported as running")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		if _, ok := runningPid(writePidFile(t, "")); ok {
			t.Error("empty pid file reported as running")
		}
	})

	t.Run("live process", func(t *testing.T) {
		// Our own pid is guaranteed alive; trailing newline matches
		// WritePidFile-adjacent formatting.
		path := writePidFile(t, strconv.Itoa(os.Getpid())+"\n")
		pid, ok := runningPid(path)
		if !ok || pid != os.Getpid() {
			t.Errorf("got (%d, %v), want (%d, true)", pid, ok, os.Getpid())
		}
	})

	t.Run("stale pid of exited process", func(t *testing.T) {
		// Spawn a process that exits immediately; its pid is then dead — the
		// exact state a crashed speaker leaves behind, which must not block
		// a fresh `speaker --background`.
		cmd := exec.Command("true")
		if err := cmd.Run(); err != nil {
			t.Fatal(err)
		}
		if _, ok := runningPid(writePidFile(t, strconv.Itoa(cmd.Process.Pid))); ok {
			t.Error("dead process reported as running")
		}
	})
}

func TestStartBackgroundDuplicate(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	if err := WritePidFile(os.Getpid()); err != nil {
		t.Fatal(err)
	}
	defer RemovePidFile()
	pid, err := StartBackground()
	if err == nil {
		t.Fatalf("duplicate start succeeded with pid %d, want already-running error", pid)
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("error %q does not mention already running", err)
	}
	if pid != os.Getpid() {
		t.Errorf("got pid %d, want existing pid %d", pid, os.Getpid())
	}
}
