// Package settings stores the user preferences of the pak.
//
// The file lives in NextUI's user data folder (not in the pak folder) so that
// it survives updates from the Pak Store.
package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Settings are the user preferences.
type Settings struct {
	Language string `json:"language"`
}

// Path returns the settings file path inside the NextUI user data folder.
func Path(userdata string) string {
	return filepath.Join(userdata, "Cheat Manager", "settings.json")
}

// Load reads the settings. A missing file returns empty settings and no
// error: this is the first launch.
func Load(path string) (Settings, error) {
	var s Settings
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	err = json.Unmarshal(data, &s)
	return s, err
}

// Save writes the settings atomically.
func Save(path string, s Settings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
