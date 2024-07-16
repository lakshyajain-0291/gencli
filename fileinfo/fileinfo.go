package fileinfo

import (
	"time"

	"github.com/google/generative-ai-go/genai"
)

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
