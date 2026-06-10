package daemon

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

// runningPidFromFile returns the pid recorded in pidFile when that process is
// still alive. A missing, malformed, or stale file (the process has exited —
// e.g. spotifyd crashed without cleaning up) yields false.
func runningPidFromFile(pidFile string) (int, bool) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}
	// On Unix, FindProcess always succeeds; probe with signal 0 to learn
	// whether the process actually exists.
	return pid, process.Signal(syscall.Signal(0)) == nil
}

func itoa(pid int) string { return strconv.Itoa(pid) }
