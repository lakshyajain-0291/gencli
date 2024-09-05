package fileinfo

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/google/generative-ai-go/genai"
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

type FileInfo struct {
	Id              int         `json:"id"`
	Name            string      `json:"name"`
	Directory       string      `json:"directory"`
	Description     string      `json:"description"`
	Size            int64       `json:"size"`
	ModifiedTime    time.Time   `json:"modifiedTime"`
	Embedding       []float32   `json:"embedding"`
	FileUploaded    bool        `json:"fileUploaded"`
	UploadedFileUrl *genai.File `json:"uploadedFIleUrl"`
}
