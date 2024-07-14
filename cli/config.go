package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type ConfigData struct {
	Directories    []string `json:"directories"`
	SkipType       []string `json:"skip_types"`
	SkipFile       []string `json:"skip_files"`
	RelevanceIndex float32  `json:"relevance_index"`
}

func setConfig(directories []string, skipTypes []string, skipFiles []string, relIndex float32) error {

	config, err := LoadConfig()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// if no directories are provided adds current working directory (can also change this to root directory of system)
			if len(directories) == 0 {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory : %w", err)
				}
				directories = append(directories, cwd)
			}

			config = &ConfigData{}

		} else {
			return err
		}
	}

	if len(directories) > 0 {
		config.Directories = directories
	}

	config.SkipType = skipTypes

	config.SkipFile = skipFiles

	if relIndex > 0 {
		config.RelevanceIndex = relIndex
	}

	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config : %w", err)
	}

	fmt.Println("Configurations Saved Successfully")
	return nil
}

func LoadConfig() (*ConfigData, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(homeDir, ".gencli-config.json")
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config ConfigData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(config *ConfigData) error {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(homeDir, ".gencli-config.json")
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(config)
}
