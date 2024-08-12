package gemini

import (
	"context"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gemini_cli_tool/fileinfo"

	"github.com/dslipak/pdf"
	"github.com/googleapis/gax-go/v2/apierror"

	"github.com/google/generative-ai-go/genai"
)

const (
	maxRetries            = 10
	baseDelay             = 100 * time.Millisecond
	maxConcurrentRequests = 10
	maxTokensPerRequest   = 900000
	timeOutDuration       = 20 * time.Second
)

func GenerateDescriptions(files []fileinfo.FileInfo, apiKeys []string, hs *fileinfo.HashSet) []fileinfo.FileInfo {
	// Create a buffered writer for the spinner output
	// writer := bufio.NewWriterSize(os.Stdout, 0)
	// spinner := fileinfo.NewSpinner(20, 100*time.Millisecond, writer)

	// // Start the spinner in a separate goroutine
	// spinner.Start()

	// // Channel to signal the progress indicator to stop
	// done := make(chan struct{})

	ctx := context.Background()
	var sessions []*Session

	for _, apikey := range apiKeys {
		session, err := NewchatSession(ctx, apikey)
		if err != nil {
			fmt.Println("Failed to start new chat session with Gemini:", err)
			return files
		}

		sessions = append(sessions, session)
	}

	var processedFiles []fileinfo.FileInfo

	// go func() {
	maxFilesPerBatch := len(apiKeys)

	fileCh := make(chan []fileinfo.FileInfo, len(files)/maxFilesPerBatch+1)
	resultCh := make(chan fileinfo.FileInfo, len(files))

	var batch []fileinfo.FileInfo
	// var currentTokens int

	for _, file := range files {

		if len(batch) >= maxFilesPerBatch {
			fileCh <- batch
			batch = []fileinfo.FileInfo{}
			// currentTokens = 0
		}
		batch = append(batch, file)
		// currentTokens += int(tokens)

	}
	if len(batch) > 0 {
		fileCh <- batch
	}
	close(fileCh)

	var wg sync.WaitGroup
	// fmt.Println("Starting concurrent processing with", maxConcurrentRequests, "goroutines")

	for i := 0; i < maxConcurrentRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// fmt.Printf("Goroutine %d started\n", id)
			for batch := range fileCh {
				// fmt.Printf("Goroutine %d processing batch", id)
				var err error
				resultBatch, err := GenerateBatchDescription(sessions, batch)
				if err != nil {
					return
				}

				for i, file := range batch {
					file.Description = resultBatch[i].Description
					resultCh <- file
				}
			}
			// fmt.Printf("Goroutine %d finished\n", id)
		}(i)
	}

	wg.Wait()
	close(resultCh)
	// fmt.Println("All goroutines have finished processing")

	for file := range resultCh {
		processedFiles = append(processedFiles, file)
	}

	// Signal the spinner to stop
	// done <- struct{}{}
	// }()

	// // Wait for the process to complete
	// <-done
	// spinner.Stop()

	return processedFiles
}

func GenerateBatchDescription(sessions []*Session, batch []fileinfo.FileInfo) ([]fileinfo.FileInfo, error) {

	// fmt.Print("length of each batch :", len(batch))
	var resultBatch []fileinfo.FileInfo

	for i, file := range batch {
		session := sessions[i]
		model := session.client.GenerativeModel("gemini-1.5-flash")

		prompt, err := GeneratePrompt(session, &file)
		if err != nil {
			fmt.Printf("Error generating prompt for file %s: %v\n", file.Name, err)
			file.Description = "nil"
			continue
		}

		resp, err := model.GenerateContent(session.ctx, prompt...)
		if err != nil {
			// fmt.Printf("%w", err.(*apierror.APIError))
			// fmt.Printf("Error generating content from Gemini: %v\n", err)
			if apiErr, ok := err.(*apierror.APIError); ok {
				// fmt.Printf("\n||%d", apiErr.HTTPCode())
				if apiErr.HTTPCode() == http.StatusTooManyRequests {
					err = retryWithBackoff(func() error {
						var retryErr error
						resp, retryErr = model.GenerateContent(session.ctx, prompt...)
						return retryErr
					})
					if err != nil {
						fmt.Printf("Error after retry: %v\n", err)
						file.Description = "nil"
						continue
					}
				} else {
					file.Description = "nil"
					continue
				}
			} else {
				file.Description = "nil"
				continue
			}
		}

		// fmt.Print("respose of file ", file.Name, " is ", resp.Candidates[0].Content.Parts[0])
		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			var builder strings.Builder

			for _, candidate := range resp.Candidates {
				for _, part := range candidate.Content.Parts {
					builder.WriteString(fmt.Sprintf("%s", part))
				}
			}
			file.Description = builder.String()
		} else {
			file.Description = "nil"
		}

		resultBatch = append(resultBatch, file)
	}

	return resultBatch, nil
}

func GeneratePrompt(session *Session, file *fileinfo.FileInfo) ([]genai.Part, error) {

	filePath := filepath.Join(file.Directory, file.Name)

	mimeType := mime.TypeByExtension(filepath.Ext(filePath))

	// fmt.Printf("\nmimeType : %s \n", mimeType)

	var descriptionFunc func() ([]genai.Part, error)

	switch {
	case strings.HasPrefix(mimeType, "text/"):
		descriptionFunc = func() ([]genai.Part, error) {
			return handleTextFile(*file)
		}
	case strings.HasSuffix(mimeType, "/pdf"):
		descriptionFunc = func() ([]genai.Part, error) {
			return handlePdfFile(*file)
		}
	case strings.HasPrefix(mimeType, "image/"):
		descriptionFunc = func() ([]genai.Part, error) {
			return handleImageFile(session, file)
		}
	case strings.HasPrefix(mimeType, "video/"):
		descriptionFunc = func() ([]genai.Part, error) {
			return handleVideoFile(session, file)
		}
	default:
		return getDefaultPrompt(*file)
	}

	prompt, err := timeOut(timeOutDuration, descriptionFunc)
	if err != nil {
		return getDefaultPrompt(*file)
	}

	return prompt, nil
}

func timeOut(duration time.Duration, fn func() ([]genai.Part, error)) ([]genai.Part, error) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	ch := make(chan struct {
		prompt []genai.Part
		err    error
	}, 1)

	go func() {
		prompt, err := fn()
		ch <- struct {
			prompt []genai.Part
			err    error
		}{prompt, err}
	}()

	select {
	case result := <-ch:
		return result.prompt, result.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func getDefaultPrompt(file fileinfo.FileInfo) ([]genai.Part, error) {
	filePath := filepath.Join(file.Directory, file.Name)

	// model := session.client.GenerativeModel("gemini-1.5-flash")
	prompt := []genai.Part{
		genai.Text(fmt.Sprintf("Using your comprehensive knowledge, generate a detailed and informative description in less than 200 words that accurately summarizes the content and purpose of this file. Consider all available metadata and context to provide insights into what this file is, its potential use, and its significance. Use the following metadata to guide your description:\n\n- File Id: %d\n- File Path: %s\n- File Size: %d bytes\n- Last Modified: %v\n\nPlease ensure the description is concise yet thorough.", file.Id, filePath, file.Size, file.ModifiedTime)),
	}

	return prompt, nil

}

func handleTextFile(file fileinfo.FileInfo) ([]genai.Part, error) {
	filePath := filepath.Join(file.Directory, file.Name)

	// Open the file for reading
	fileHandle, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()

	// buffer size (4KB) to limit memory usage
	bufferSize := 4096
	buffer := make([]byte, bufferSize)

	// Reading the first chunk of the file content
	n, err := fileHandle.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Convert the read bytes into a string
	contentSnippet := string(buffer[:n])

	// Create the prompt using the snippet
	prompt := []genai.Part{
		genai.Text(fmt.Sprintf(
			"Using the provided text snippet, generate a detailed and insightful description in less than 200 words that captures the essence, purpose, and key topics of this file.\n\nFile Id: %d\n- File Path: %s\n- File Size: %d bytes\n- Last Modified: %v\n\nPlease ensure the description is concise yet thorough.\n\nContent Snippet: \n\n%s\n\nIf relevant, infer the file's broader context or potential uses.",
			file.Id, filePath, file.Size, file.ModifiedTime, contentSnippet)),
	}

	return prompt, nil
}

func handlePdfFile(file fileinfo.FileInfo) ([]genai.Part, error) {
	filePath := filepath.Join(file.Directory, file.Name)

	// Opening the PDF file
	r, err := pdf.Open(filePath)
	if err != nil {
		return nil, err
	}

	// Getting plain text from PDF
	b, err := r.GetPlainText()
	if err != nil {
		return nil, err
	}
	// defer r.Close()

	// Read only the first portion of the text to avoid large memory usage
	bufferSize := 4096
	buffer := make([]byte, bufferSize)

	n, err := b.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Convert the read bytes into a string
	contentSnippet := string(buffer[:n])

	// Create the prompt using the snippet
	prompt := []genai.Part{
		genai.Text(fmt.Sprintf("From the provided PDF content snippet, generate an in-depth description in less than 200 words that highlights the main themes, purpose, and possible applications of this document. File Id: %dd\n- File Path: %s\n- File Size: %d bytes\n- Last Modified: %v\n\n. Content Snippet: \n\n%s\n\nAdditionally, consider the document's structure or any inferred context.", file.Id, filePath, file.Size, file.ModifiedTime, contentSnippet)),
	}
	// fmt.Println(prompt[0])
	return prompt, nil
}

func handleImageFile(session *Session, file *fileinfo.FileInfo) ([]genai.Part, error) {
	if file.FileUploaded {
		return processUploadedImage(*file, file.UploadedFileUrl)
	}

	uploadedFile, err := uploadImage(session, *file)
	if err != nil {
		return nil, err
	}

	file.FileUploaded = true
	file.UploadedFileUrl = uploadedFile

	return processUploadedImage(*file, uploadedFile)

}

func uploadImage(session *Session, file fileinfo.FileInfo) (*genai.File, error) {
	filePath := filepath.Join(file.Directory, file.Name)

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	//setting a display name
	opts := genai.UploadFileOptions{DisplayName: filepath.Base(filePath)}

	// fmt.Print("uploading file..", file.Name, "\n")

	//Uploaded file to gemini
	img, err := session.client.UploadFile(session.ctx, "", f, &opts)
	if err != nil {
		return nil, err
	}
	// fmt.Print("uploaded file..", file.Name, "\n")

	//Got metadata of the uploaded file
	uploadedFile, err := session.client.GetFile(session.ctx, img.Name)
	if err != nil {
		return nil, err
	}

	return uploadedFile, nil
}

func processUploadedImage(file fileinfo.FileInfo, uploadedFile *genai.File) ([]genai.Part, error) {
	// filePath := filepath.Join(file.Directory, file.Name)

	prompt := []genai.Part{
		genai.FileData{URI: uploadedFile.URI},
		genai.Text(fmt.Sprintf("Generate a rich and detailed description in less than 200 words about the subject, context, and potential significance of this image file. Consider its visual elements, style, and possible context. File Id: %d\n- File Name: %s\n- File Size: %d bytes\n- Last Modified: %v\n\n", file.Id, file.Name, file.Size, file.ModifiedTime)),
	}

	return prompt, nil
}

func handleVideoFile(session *Session, file *fileinfo.FileInfo) ([]genai.Part, error) {
	if file.Size > 50*1024*1024 {
		// fmt.Println("File too big .. calling default func")
		return getDefaultPrompt(*file)
	}

	if file.FileUploaded {
		// fmt.Println("Already uploaded .. moving to processing..")
		return processUploadedVideo(*file, file.UploadedFileUrl)
	}

	filePath := filepath.Join(file.Directory, file.Name)
	trimmedFilePath, err := trimVideo(filePath, 10)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer os.Remove(trimmedFilePath)

	uploadedFile, err := uploadVideo(session, trimmedFilePath, *file)
	if err != nil {
		return nil, err
	}

	file.FileUploaded = true
	file.UploadedFileUrl = uploadedFile

	return processUploadedVideo(*file, uploadedFile)

}

func trimVideo(filePath string, duration int) (string, error) {
	ext := filepath.Ext(filePath)
	base := filePath[:len(filePath)-len(ext)]

	// Create the trimmed file path by appending "_trimmed" before the file extension
	trimmedFilePath := base + "_trimmed" + ext
	// fmt.Printf("\nInside trimmed video func from %s : %s\n", filePath, trimmedFilePath)

	cmd := exec.Command("ffmpeg", "-i", filePath, "-t", fmt.Sprintf("%d", duration), "-c", "copy", trimmedFilePath)
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return trimmedFilePath, nil
}

func uploadVideo(session *Session, trimmedFilePath string, file fileinfo.FileInfo) (*genai.File, error) {

	f, err := os.Open(trimmedFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	//setting a display name
	opts := genai.UploadFileOptions{DisplayName: filepath.Base(trimmedFilePath)}

	// fmt.Print("uplaoding file..", file.Name, "\n")
	//Uploaded file to gemini
	vid, err := session.client.UploadFile(session.ctx, "", f, &opts)
	if err != nil {
		return nil, err
	}
	// fmt.Print("uplaoded file..", file.Name, "\n")

	// const baseDelay = 10 * time.Millisecond

	// Get metadata of the uploaded file

	uploadedFile, err := session.client.GetFile(session.ctx, vid.Name)
	if err != nil {
		return nil, err
	}

	for uploadedFile.State == genai.FileStateProcessing {
		fmt.Print(".")
		// Sleep for 10 seconds
		time.Sleep(10 * time.Second)

		var err error

		// Fetch the file from the API again.
		uploadedFile, err = session.client.GetFile(session.ctx, file.Name)
		if err != nil {
			return nil, err
		}
	}

	return uploadedFile, nil
}

func processUploadedVideo(file fileinfo.FileInfo, uploadedFile *genai.File) ([]genai.Part, error) {
	// filePath := filepath.Join(file.Directory, file.Name)

	// model := session.client.GenerativeModel("gemini-1.5-flash")
	prompt := []genai.Part{
		genai.FileData{URI: uploadedFile.URI},
		genai.Text(fmt.Sprintf("Generate a well-rounded description in less than 200 words about what this video file likely depicts and its possible purpose. Use your knowledge to interpret the content and any associated metadata. The file Id is %d\n- File Name: %s\n- File Size: %d bytes\n- Last Modified: %v\n\n", file.Id, file.Name, file.Size, file.ModifiedTime)),
	}

	return prompt, nil

}

func GenerateEmbeddings(files []fileinfo.FileInfo, defaultApiKey string) []fileinfo.FileInfo {

	ctx := context.Background()

	session, err := NewchatSession(ctx, defaultApiKey)
	if err != nil {
		fmt.Println("Failed to start new chat session with Gemini:", err)
		return files
	}

	fileCh := make(chan fileinfo.FileInfo, len(files))
	resultCh := make(chan fileinfo.FileInfo, len(files))

	for _, file := range files {
		fileCh <- file
	}
	close(fileCh)

	var wg sync.WaitGroup
	const maxConcurrentRequests = 10
	// fmt.Println("Starting concurrent processing with", maxConcurrentRequests, "goroutines")

	for i := 0; i < maxConcurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// fmt.Printf("Goroutine %d started\n", id)

			for file := range fileCh {
				// fmt.Printf("Goroutine %d processing file: %s\n", id, file.Name)

				var err error
				file.Embedding, err = GenerateEmbedding(session, file.Description)
				if err != nil {
					// fmt.Printf("%w", err.(*apierror.APIError))
					if apiErr, ok := err.(*apierror.APIError); ok {
						// fmt.Printf("\n||%d", apiErr.HTTPCode())
						if apiErr.HTTPCode() == http.StatusTooManyRequests {
							err = retryWithBackoff(func() error {
								var retryErr error
								file.Embedding, retryErr = GenerateEmbedding(session, file.Description)
								return retryErr
							})

							if err != nil {
								file.Embedding = nil
							}
						} else {
							file.Embedding = nil
						}
					} else {
						file.Embedding = nil
					}
				}
				resultCh <- file
			}
			// fmt.Printf("Goroutine %d finished\n", id)
		}()
	}

	wg.Wait()
	close(resultCh)
	// fmt.Println("All goroutines have finished processing")

	var processedFiles []fileinfo.FileInfo
	for file := range resultCh {
		processedFiles = append(processedFiles, file)
	}
	return processedFiles
}

func GenerateEmbedding(chatSession *Session, desc string) ([]float32, error) {

	em := chatSession.client.EmbeddingModel("text-embedding-004")
	embeddingResult, err := em.EmbedContent(chatSession.ctx, genai.Text(desc))
	if err != nil {
		return nil, err
	}

	// fmt.Println(embeddingResult.Embedding.Values)
	return embeddingResult.Embedding.Values, nil

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
