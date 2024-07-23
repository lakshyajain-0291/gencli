package fileinfo

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type HashSet struct {
	mu    sync.Mutex
	store map[string]struct{}
}

func NewHashSet() *HashSet {
	return &HashSet{
		store: make(map[string]struct{}),
	}
}

func (hs *HashSet) Add(hashString string) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	hs.store[hashString] = struct{}{}
}

func (hs *HashSet) Exists(hashString string) bool {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	_, exists := hs.store[hashString]
	return exists
}

func (hs *HashSet) Remove(hashString string) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	delete(hs.store, hashString)
}

func (hs *HashSet) SaveToFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	hashesPath := filepath.Join(homeDir, "gencli\\.gencli-hashes.json")
	file, err := os.Create(hashesPath)
	if err != nil {
		return err
	}
	defer file.Close()

	hs.mu.Lock()
	defer hs.mu.Unlock()

	encoder := json.NewEncoder(file)
	return encoder.Encode(hs.store)

}

func (hs *HashSet) LoadFromFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	hashesPath := filepath.Join(homeDir, "gencli\\.gencli-hashes.json")

	file, err := os.Open(hashesPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, it's not an error; it just means no hashes were saved yet.
			return nil
		}
		return err
	}
	defer file.Close()

	hs.mu.Lock()
	defer hs.mu.Unlock()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&hs.store)
}

// Generate a unique hash string for a file based on its properties
func GenerateFileHash(file FileInfo) string {
	// Combine file properties to create a unique string
	fileData := fmt.Sprintf("%s|%s|%d|%s", file.Name, file.Directory, file.Size, file.ModifiedTime)
	// fmt.Printf("generating hash for %s|%s|%d|%s", file.Name, file.Directory, file.Size, file.ModifiedTime)
	hash := sha256.Sum256([]byte(fileData))
	return fmt.Sprintf("%x", hash)
}
