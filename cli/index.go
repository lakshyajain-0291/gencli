package cli

import (
	"fmt"
	"gemini_cli_tool/fileinfo"
	"gemini_cli_tool/gemini"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/googleapis/gax-go/v2/apierror"
	"github.com/spf13/cobra"
)

const (
	maxRetries = 20
	baseDelay  = 100 * time.Millisecond
)

// var writer = bufio.NewWriter(os.Stdout)
// var spinners = newSpinner(5, time.Second, writer)

func indexFilesCmd(cmd *cobra.Command, args []string) error {
	err := indexFiles()

	// spinners.stop()

	if err == nil {
		fmt.Println("Indexing completed successfully.")
	}

	return err
}

func indexFiles() error {

	// spinners.start()

	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config : %w", err)
	}

	files, err := LoadIndex()

	i := 0
	for _, dir := range config.Directories {
		// fmt.Printf("Checking directory: %s\n", dir)

		// 	Check if the directory exists
		info, err := os.Stat(dir)
		if os.IsNotExist(err) {
			fmt.Printf("Directory %s does not exist. Please create it.\n", dir)
			continue
		} else if err != nil {
			fmt.Printf("Error checking directory %s: %v\n", dir, err)
			continue
		} else if !info.IsDir() {
			fmt.Printf("Path %s is not a directory.\n", dir)
			continue
		}

		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			// fmt.Printf(" in file walk in dir %s ", dir)

			// fmt.Printf("%t, %t, i: %d, name: %s, dir: %s, size: %d, time: %v", !info.IsDir(), !shouldSkip(info.Name(), config.SkipType, config.SkipFile), i, info.Name(), filepath.Dir(path), info.Size(), info.ModTime())

			if !info.IsDir() && !shouldSkip(info.Name(), config.SkipType, config.SkipFile) {
				files = append(files, fileinfo.FileInfo{
					Id:           i,
					Name:         info.Name(),
					Directory:    filepath.Dir(path),
					Size:         info.Size(),
					ModifiedTime: info.ModTime(),
				})

				i = i + 1
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk dircetory %s : %w", dir, err)
		}

	}

	//Generate descriptions using Gemini

	for i, file := range files {

		description, err := gemini.GenerateDescription(file)
		if err != nil {
			// fmt.Printf("%w", err.(*apierror.APIError))
			if apiErr, ok := err.(*apierror.APIError); ok {
				// fmt.Printf("\n||%d", apiErr.HTTPCode())
				if apiErr.HTTPCode() == http.StatusTooManyRequests {
					err = retryWithBackoff(func() error {
						var retryErr error
						description, retryErr = gemini.GenerateDescription(file)
						return retryErr
					})

					if err != nil {
						return fmt.Errorf("failed to generate description : %w", err)
					}
				}
			}
		}

		// fmt.Printf("\n\nFile Name : %s \n", file.Name)
		// fmt.Printf("\nDescription : %s \n", description)

		embedding, err := gemini.GenerateEmbeddings(description)
		if err != nil {
			// fmt.Printf("%w", err.(*apierror.APIError))
			if apiErr, ok := err.(*apierror.APIError); ok {
				// fmt.Printf("\n||%d", apiErr.HTTPCode())
				if apiErr.HTTPCode() == http.StatusTooManyRequests {

					err = retryWithBackoff(func() error {
						var retryErr error
						embedding, retryErr = gemini.GenerateEmbeddings(description)
						return retryErr
					})

					if err != nil {
						return fmt.Errorf("failed to generate description : %w", err)
					}
				}
			}
		}

		files[i].Description = description
		files[i].Embedding = embedding

	}

	if err := StoreIndex(files); err != nil {
		return fmt.Errorf("failed to store index : %w", err)
	}

	return nil
}

func retryWithBackoff(operation func() error) error {
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil //success
		}
		// fmt.Printf("\n%d\n", err.(*apierror.APIError).HTTPCode())
		if apiErr, ok := err.(*apierror.APIError); ok {
			// fmt.Printf("\n||Attempt no : %d||\n", attempt)
			if apiErr.HTTPCode() == http.StatusTooManyRequests {
				delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
				time.Sleep(delay)
				continue
			}
		}
		break
	}

	return fmt.Errorf("operation failed after %d attempts : %w", maxRetries+1, err)

}

func shouldSkip(fileName string, skipTypes []string, skipFiles []string) bool {
	for _, skipType := range skipTypes {
		if strings.HasSuffix(fileName, skipType) {
			return true
		}
	}

	for _, skipFile := range skipFiles {
		if strings.HasPrefix(fileName, skipFile) {
			return true
		}
	}
	return false
}
