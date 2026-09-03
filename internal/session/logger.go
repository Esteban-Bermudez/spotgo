package session

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Esteban-Bermudez/spotgo/config"
	librespot "github.com/devgianlu/go-librespot"
	"github.com/pkg/browser"
)

type logger struct {
	fields map[string]interface{}
	out    *log.Logger
	err    *log.Logger
}

// LogPath is the file all player-process logging goes to, so the bubbletea
// TUI never gets stdout/stderr spam.
func LogPath() string {
	return filepath.Join(config.StateDir(), "librespot.log")
}

// logWriters opens librespot.log once per process and hands out the shared
// handle. newLogger used to open the file on every call, and
// HasStoredCredentials calls it on every Setup ticker tick, leaking a file
// descriptor per call.
var (
	logFileOnce  sync.Once
	logOutWriter io.Writer
	logErrWriter io.Writer
)

func logWriters() (io.Writer, io.Writer) {
	logFileOnce.Do(func() {
		stateDir := config.StateDir()
		_ = os.MkdirAll(stateDir, 0755)
		logPath := filepath.Join(stateDir, "librespot.log")
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			logOutWriter = io.Discard
			logErrWriter = os.Stderr
			return
		}
		logOutWriter = f
		logErrWriter = f
	})
	return logOutWriter, logErrWriter
}

func newLogger() librespot.Logger {
	w, ew := logWriters()
	return &logger{
		fields: map[string]interface{}{},
		out:    log.New(w, "", log.LstdFlags),
		err:    log.New(ew, "", log.LstdFlags),
	}
}

// discardLogger logs nowhere. Credential/state probes must not touch the log
// file: Setup calls HasStoredCredentials every second until login completes.
func discardLogger() librespot.Logger {
	return &logger{
		fields: map[string]interface{}{},
		out:    log.New(io.Discard, "", 0),
		err:    log.New(io.Discard, "", 0),
	}
}

// suffix renders WithField/WithError entries as ` key=value` pairs so they
// are not silently dropped: go-librespot reports failure reasons (e.g. why
// a dealer play command was refused) exclusively through fields.
func (l *logger) suffix() string {
	if len(l.fields) == 0 {
		return ""
	}
	pairs := make([]string, 0, len(l.fields))
	for k, v := range l.fields {
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}
	sort.Strings(pairs)
	return " " + strings.Join(pairs, " ")
}

func (l *logger) Tracef(format string, args ...interface{}) {
	l.out.Printf("[TRACE] "+format+l.suffix(), args...)
}
func (l *logger) Debugf(format string, args ...interface{}) {
	l.out.Printf("[DEBUG] "+format+l.suffix(), args...)
}
func (l *logger) Infof(format string, args ...interface{}) {
	l.out.Printf("[INFO] "+format+l.suffix(), args...)
	if !strings.Contains(format, "complete authentication") && !strings.Contains(format, "visit the following link") {
		return
	}
	for _, a := range args {
		if s, ok := a.(string); ok && strings.HasPrefix(s, "http") {
			log.Printf("To complete Spotify Connect authentication, visit: %s", s)
			_ = browser.OpenURL(s)
			return
		}
	}
	log.Printf("[INFO] "+format, args...)
}
func (l *logger) Warnf(format string, args ...interface{}) {
	l.err.Printf("[WARN] "+format+l.suffix(), args...)
}
func (l *logger) Errorf(format string, args ...interface{}) {
	l.err.Printf("[ERROR] "+format+l.suffix(), args...)
}

func (l *logger) withSuffix(args []interface{}) []interface{} {
	out := append([]interface{}{}, args...)
	if s := l.suffix(); s != "" {
		out = append(out, strings.TrimPrefix(s, " "))
	}
	return out
}

func (l *logger) Trace(args ...interface{}) {
	l.out.Println(append([]interface{}{"[TRACE]"}, l.withSuffix(args)...)...)
}
func (l *logger) Debug(args ...interface{}) {
	l.out.Println(append([]interface{}{"[DEBUG]"}, l.withSuffix(args)...)...)
}
func (l *logger) Info(args ...interface{}) {
	l.out.Println(append([]interface{}{"[INFO]"}, l.withSuffix(args)...)...)
}
func (l *logger) Warn(args ...interface{}) {
	l.err.Println(append([]interface{}{"[WARN]"}, l.withSuffix(args)...)...)
}
func (l *logger) Error(args ...interface{}) {
	l.err.Println(append([]interface{}{"[ERROR]"}, l.withSuffix(args)...)...)
}

func (l *logger) WithField(key string, value interface{}) librespot.Logger {
	n := &logger{fields: map[string]interface{}{key: value}, out: l.out, err: l.err}
	for k, v := range l.fields {
		n.fields[k] = v
	}
	return n
}

func (l *logger) WithError(err error) librespot.Logger {
	return l.WithField("error", err)
}
