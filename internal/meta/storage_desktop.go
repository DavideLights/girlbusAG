//go:build !js || !wasm

package meta

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func saveFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "girlbusAG", "save.json"), nil
}

// Load reads SaveData from the user's config directory, returning defaults if missing or corrupt.
func Load() SaveData {
	data := DefaultSaveData()
	path, err := saveFilePath()
	if err != nil {
		return data
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return data
	}

	if err := json.Unmarshal(raw, &data); err != nil {
		return DefaultSaveData()
	}
	data.Normalize()
	return data
}

// Save normalizes and writes SaveData to the user's config directory.
func Save(data SaveData) error {
	data.Normalize()

	path, err := saveFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
