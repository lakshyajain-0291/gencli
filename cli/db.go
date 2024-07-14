package cli

import (
	"encoding/json"
	"gemini_cli_tool/fileinfo"
	"os"
	"path/filepath"
)

func LoadIndex() ([]fileinfo.FileInfo, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	indexPath := filepath.Join(homeDir, ".gencli-index.json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, err
	}

	var files []fileinfo.FileInfo
	if err := json.Unmarshal(data, &files); err != nil {
		return nil, err
	}

	return files, nil
}

func StoreIndex(files []fileinfo.FileInfo) error {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	indexPath := filepath.Join(homeDir, ".gencli-index.json")
	jsonData, err := json.MarshalIndent(files, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, jsonData, 0644)
}
