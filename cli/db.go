package cli

import (
	"encoding/json"
	"gemini_cli_tool/fileinfo"
	"os"
	"path/filepath"
	"runtime"
)

func GetConfigDir() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		configDir := filepath.Join(homedir, "Appdata", "Local", "gencli")
		return configDir, nil
	case "darwin":
		configDir := filepath.Join(homedir, "Library", "Application Support", "gencli")
		return configDir, nil
	default:
		configDir := filepath.Join(homedir, ".config", "gencli")
		return configDir, nil

	}

}

func LoadIndex() ([]fileinfo.FileInfo, error) {

	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	var files []fileinfo.FileInfo

	indexPath := filepath.Join(configDir, ".gencli-index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, it's not an error; it just means no hashes were saved yet.
			return files, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &files); err != nil {
		return nil, err
	}

	return files, nil
}

func StoreIndex(files []fileinfo.FileInfo) error {

	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(configDir, ".gencli-index.json")
	jsonData, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, jsonData, 0644)
}
