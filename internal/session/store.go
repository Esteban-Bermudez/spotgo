package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	librespot "github.com/devgianlu/go-librespot"
)

type fileStateStore struct {
	path string
	log  librespot.Logger
}

func newFileStateStore(path string, log librespot.Logger) librespot.StateStore {
	return &fileStateStore{path: path, log: log}
}

func (s *fileStateStore) Load() (*librespot.AppState, error) {
	state := &librespot.AppState{}
	data, err := os.ReadFile(s.path)
	switch {
	case err == nil:
		if err := json.Unmarshal(data, state); err != nil {
			return nil, fmt.Errorf("unmarshal state: %w", err)
		}
		s.log.Debugf("state loaded from %s", s.path)
	case errors.Is(err, os.ErrNotExist):
		s.log.Debugf("no state file at %s", s.path)
	default:
		return nil, fmt.Errorf("read state file: %w", err)
	}
	return state, nil
}

func (s *fileStateStore) Save(state *librespot.AppState) error {
	state.Lock()
	defer state.Unlock()

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	tmp, err := os.CreateTemp(dir, "state-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp state file: %w", err)
	}
	tmpPath := tmp.Name()
	if err := json.NewEncoder(tmp).Encode(state); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("encode state: %w", err)
	}
	tmp.Close()
	if err := os.Rename(tmpPath, s.path); err != nil {
		// Windows can't rename over existing file.
		if err2 := os.Remove(s.path); err2 != nil && !errors.Is(err2, os.ErrNotExist) {
			os.Remove(tmpPath)
			return err
		}
		if err := os.Rename(tmpPath, s.path); err != nil {
			os.Remove(tmpPath)
			return err
		}
	}
	return nil
}
