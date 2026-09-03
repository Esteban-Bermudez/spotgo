package session

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

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

func newLogger() librespot.Logger {
	stateDir := config.StateDir()
	_ = os.MkdirAll(stateDir, 0755)
	logPath := filepath.Join(stateDir, "librespot.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	var w io.Writer = io.Discard
	var ew io.Writer = io.Discard
	if err == nil {
		w = f
		ew = f
	} else {
		w = io.Discard
		ew = os.Stderr
	}
	return &logger{
		fields: map[string]interface{}{},
		out:    log.New(w, "", log.LstdFlags),
		err:    log.New(ew, "", log.LstdFlags),
	}
}

func (l *logger) Tracef(format string, args ...interface{}) { l.out.Printf("[TRACE] "+format, args...) }
func (l *logger) Debugf(format string, args ...interface{}) { l.out.Printf("[DEBUG] "+format, args...) }
func (l *logger) Infof(format string, args ...interface{}) {
	l.out.Printf("[INFO] "+format, args...)
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
func (l *logger) Warnf(format string, args ...interface{})  { l.err.Printf("[WARN] "+format, args...) }
func (l *logger) Errorf(format string, args ...interface{}) { l.err.Printf("[ERROR] "+format, args...) }

func (l *logger) Trace(args ...interface{}) {
	l.out.Println(append([]interface{}{"[TRACE]"}, args...)...)
}
func (l *logger) Debug(args ...interface{}) {
	l.out.Println(append([]interface{}{"[DEBUG]"}, args...)...)
}
func (l *logger) Info(args ...interface{}) {
	l.out.Println(append([]interface{}{"[INFO]"}, args...)...)
}
func (l *logger) Warn(args ...interface{}) {
	l.err.Println(append([]interface{}{"[WARN]"}, args...)...)
}
func (l *logger) Error(args ...interface{}) {
	l.err.Println(append([]interface{}{"[ERROR]"}, args...)...)
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
