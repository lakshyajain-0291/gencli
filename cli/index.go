package cli

import (
	"fmt"
	"gemini_cli_tool/fileinfo"
	"gemini_cli_tool/gemini"
	"os"
	"path/filepath"
	"strings"
)

// var writer = bufio.NewWriter(os.Stdout)
// var spinners = newSpinner(5, time.Second, writer)

func indexFilesCmd(hs *fileinfo.HashSet) error {
	err := indexFiles(hs)

	// spinners.stop()

	if err == nil {
		fmt.Println("Indexing completed successfully.")
	}

	return err
}

func indexFiles(hs *fileinfo.HashSet) error {

	// spinners.start()

	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config : %w", err)
	}

	apiKeys := config.APIKeys
	if apiKeys == nil {
		return fmt.Errorf("no apikeys provided")
	}
	defaultApiKey := apiKeys[0]

	indexedFiles, err := LoadIndex()
	if err != nil {
		return err
	}

	var toIndexFiles = []fileinfo.FileInfo{}
	var newFiles = []fileinfo.FileInfo{}
	var finalFiles = []fileinfo.FileInfo{}

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
				file := fileinfo.FileInfo{
					Id:           i,
					Name:         info.Name(),
					Directory:    filepath.Dir(path),
					Size:         info.Size(),
					ModifiedTime: info.ModTime(),
					FileUploaded: false,
				}

				toIndexFiles = append(toIndexFiles, file)
				i++
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk dircetory %s : %w", dir, err)
		}

	}

	//implemented sorting for better management of resorces during threading
	// sort.Slice(files, func(i, j int) bool {
	// 	return files[i].Size < files[j].Size
	// })

	// for _, file := range files {
	// 	fmt.Printf("\nFile Size : %d \n", file.Size)
	// 	fmt.Printf("\nFile Name : %s \n", file.Name)
	// }

	existingFiles := make(map[string]fileinfo.FileInfo)

	for _, file := range indexedFiles {
		fileHash := fileinfo.GenerateFileHash(file)
		existingFiles[fileHash] = file
	}

	print("existingFiles : ")

	for _, file := range toIndexFiles {
		fmt.Printf("\nFile Size : %d \n", file.Size)
		fmt.Printf("\nFile Name : %s \n", file.Name)
		fmt.Printf("\nDescription : %s \n", file.Description)
	}

	// Identify deleted files
	for _, file := range toIndexFiles {
		fileHash := fileinfo.GenerateFileHash(file)
		if _, exists := existingFiles[fileHash]; exists {
			finalFiles = append(finalFiles, existingFiles[fileHash])
		} else {
			hs.Remove(fileHash)
		}
	}

	// Identify new files
	for _, file := range toIndexFiles {
		fileHash := fileinfo.GenerateFileHash(file)
		if !hs.Exists(fileHash) {
			newFiles = append(newFiles, file)
			hs.Add(fileHash) // Add hash to the set
			fmt.Printf("\nNot Skipping file %s\\%s\n", file.Directory, file.Name)
		} else {
			fmt.Printf("\nSkipping file %s\\%s\n", file.Directory, file.Name)

		}
	}

	//Generate descriptions using Gemini
	newFiles = gemini.GenerateDescriptions(newFiles, apiKeys, hs)
	newFiles = gemini.GenerateEmbeddings(newFiles, defaultApiKey)

	finalFiles = append(finalFiles, newFiles...)

	print("finalFiles : ")

	for _, file := range finalFiles {
		fmt.Printf("\nFile Size : %d \n", file.Size)
		fmt.Printf("\nFile Name : %s \n", file.Name)
		fmt.Printf("\nDescription : %s \n", file.Description)
	}

	if err := StoreIndex(finalFiles); err != nil {
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
