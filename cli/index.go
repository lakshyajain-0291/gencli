package cli

import (
	"fmt"
	"gemini_cli_tool/fileinfo"
	"gemini_cli_tool/gemini"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
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

	// files, err := LoadIndex()

	var files = []fileinfo.FileInfo{}

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
					Id:              i,
					Name:            info.Name(),
					Directory:       filepath.Dir(path),
					Size:            info.Size(),
					ModifiedTime:    info.ModTime(),
					FileUploaded:    false,
					UploadedFileUrl: nil,
				})

				i = i + 1
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk dircetory %s : %w", dir, err)
		}

	}

	//implemented sorting for better management of resorces during threading
	sort.Slice(files, func(i, j int) bool {
		return files[i].Size < files[j].Size
	})

	for _, file := range files {
		fmt.Printf("\nFile Size : %d \n", file.Size)
		fmt.Printf("\nFile Name : %s \n", file.Name)
	}

	//Generate descriptions using Gemini
	files = gemini.GenerateDescriptions(files)

	for _, file := range files {
		fmt.Printf("\nFile Size : %d \n", file.Size)
		fmt.Printf("\nFile Name : %s \n", file.Name)
		fmt.Printf("\nDescription : %s \n", file.Description)
	}

	files = gemini.GenerateEmbeddings(files)

	if err := StoreIndex(files); err != nil {
		return fmt.Errorf("failed to store index : %w", err)
	}

	return nil
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
