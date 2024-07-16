package gemini

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"mime"
	"net/http"
	"os"
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
	maxRetries            = 20
	baseDelay             = 100 * time.Millisecond
	maxConcurrentRequests = 15
	maxTokensPerRequest   = 1000000
	maxFilesPerBatch      = 10
)

func GenerateDescriptions(files []fileinfo.FileInfo) []fileinfo.FileInfo {

	ctx := context.Background()
	session, err := NewsetupSession(ctx)
	if err != nil {
		fmt.Println("Failed to start new setup session with Gemini:", err)
		return files
	}

	fileCh := make(chan fileinfo.FileInfo, len(files))
	resultCh := make(chan fileinfo.FileInfo, len(files))

	for _, file := range files {
		fileCh <- file
	}
	close(fileCh)

	var wg sync.WaitGroup
	fmt.Println("Starting concurrent processing with", maxConcurrentRequests, "goroutines")

	for i := 0; i < maxConcurrentRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Printf("Goroutine %d started\n", id)
			for file := range fileCh {
				fmt.Printf("Goroutine %d processing file: %s\n", id, file.Name)
				var err error
				file.Description, err = GenerateDescription(session, file)
				if err != nil {
					// fmt.Printf("%w", err.(*apierror.APIError))
					if apiErr, ok := err.(*apierror.APIError); ok {
						// fmt.Printf("\n||%d", apiErr.HTTPCode())
						if apiErr.HTTPCode() == http.StatusTooManyRequests {
							err = retryWithBackoff(func() error {
								var retryErr error
								file.Description, retryErr = GenerateDescription(session, file)
								return retryErr
							})

							if err != nil {
								file.Description = "nil"
							}
						} else {
							file.Description = "nil"
						}
					} else {
						file.Description = "nil"
					}
				}
				resultCh <- file
			}
			fmt.Printf("Goroutine %d finished\n", id)
		}(i)
	}

	wg.Wait()
	close(resultCh)
	fmt.Println("All goroutines have finished processing")

	var processedFiles []fileinfo.FileInfo
	for file := range resultCh {
		processedFiles = append(processedFiles, file)
	}
	return processedFiles
}

func GenerateDescription(session *Session, file fileinfo.FileInfo) (string, error) {

	filePath := filepath.Join(file.Directory, file.Name)

	mimeType := mime.TypeByExtension(filepath.Ext(filePath))

	// fmt.Printf("\nmimeType : %s \n", mimeType)

	var descriptionFunc func() (string, error)

	switch {
	case strings.HasPrefix(mimeType, "text/"):
		descriptionFunc = func() (string, error) {
			return handleTextFile(session, file)
		}
	case strings.HasSuffix(mimeType, "/pdf"):
		descriptionFunc = func() (string, error) {
			return handlePdfFile(session, file)
		}
	case strings.HasPrefix(mimeType, "image/"):
		descriptionFunc = func() (string, error) {
			return handleImageFile(session, file)
		}
	case strings.HasPrefix(mimeType, "video/"):
		descriptionFunc = func() (string, error) {
			return handleVideoFile(session, file)
		}
	default:
		return getDefaultDesc(session, file)
	}

	description, err := timeOut(30*time.Second, descriptionFunc)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	return description, nil
}

func timeOut(duration time.Duration, fn func() (string, error)) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	ch := make(chan struct {
		desc string
		err  error
	}, 1)

	go func() {
		desc, err := fn()
		ch <- struct {
			desc string
			err  error
		}{desc, err}
	}()

	select {
	case result := <-ch:
		return result.desc, result.err
	case <-ctx.Done():
		return "No description", ctx.Err()
	}
}

func getDefaultDesc(session *Session, file fileinfo.FileInfo) (string, error) {

	model := session.client.GenerativeModel("gemini-1.5-flash")
	prompt := []genai.Part{
		genai.Text(fmt.Sprintf("Generate a description in less than 100 words about what this file is and what is it about, based on the meta data given : \n\nFile Name : %s\nFile Directory : %s\nFile Size : %d\nFile Modified Time : %v", file.Name, file.Directory, file.Size, file.ModifiedTime)),
	}

	resp, err := model.GenerateContent(session.ctx, prompt...)
	if err != nil {
		return "No description", err
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		var builder strings.Builder

		for _, candidate := range resp.Candidates {
			for _, part := range candidate.Content.Parts {
				builder.WriteString(fmt.Sprintf("%s", part))
			}
		}
		return builder.String(), nil
	}

	return "No description", fmt.Errorf("no description generated")

}

func handleTextFile(session *Session, file fileinfo.FileInfo) (string, error) {

	filePath := filepath.Join(file.Directory, file.Name)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	model := session.client.GenerativeModel("gemini-1.5-flash")
	prompt := []genai.Part{
		genai.Text(fmt.Sprintf("Generate a description in less than 100 words about what this file is and what is it about, based on the given content : \n\n%s", string(content))),
	}

	resp, err := model.GenerateContent(session.ctx, prompt...)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		var builder strings.Builder

		for _, candidate := range resp.Candidates {
			for _, part := range candidate.Content.Parts {
				builder.WriteString(fmt.Sprintf("%s", part))
			}
		}
		return builder.String(), nil
	}

	return getDefaultDesc(session, file)
}

func handlePdfFile(session *Session, file fileinfo.FileInfo) (string, error) {
	filePath := filepath.Join(file.Directory, file.Name)

	r, err := pdf.Open(filePath)
	if err != nil {
		return getDefaultDesc(session, file)
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return getDefaultDesc(session, file)
	}
	buf.ReadFrom(b)
	content := buf.String()

	model := session.client.GenerativeModel("gemini-1.5-flash")
	prompt := []genai.Part{
		genai.Text(fmt.Sprintf("Generate a description in less than 100 words about what this file is and what is it about, based on the given content : \n\n%s", content)),
	}

	resp, err := model.GenerateContent(session.ctx, prompt...)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		var builder strings.Builder

		for _, candidate := range resp.Candidates {
			for _, part := range candidate.Content.Parts {
				builder.WriteString(fmt.Sprintf("%s", part))
			}
		}
		return builder.String(), nil
	}

	return getDefaultDesc(session, file)

}

func handleImageFile(session *Session, file fileinfo.FileInfo) (string, error) {

	filePath := filepath.Join(file.Directory, file.Name)

	f, err := os.Open(filePath)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	//setting a display name
	opts := genai.UploadFileOptions{DisplayName: filepath.Base(filePath)}

	//Uploaded file to gemini
	img, err := session.client.UploadFile(session.ctx, "", f, &opts)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	//Got metadata of the uploaded file
	uploadedFile, err := session.client.GetFile(session.ctx, img.Name)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	model := session.client.GenerativeModel("gemini-1.5-flash")
	prompt := []genai.Part{
		genai.FileData{URI: uploadedFile.URI},
		genai.Text(fmt.Sprintf("Generate a description in less than 100 words about what this file %s is and what is it about ", filepath.Base(filePath))),
	}

	resp, err := model.GenerateContent(session.ctx, prompt...)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		var builder strings.Builder

		for _, candidate := range resp.Candidates {
			for _, part := range candidate.Content.Parts {
				builder.WriteString(fmt.Sprintf("%s", part))
			}
		}
		return builder.String(), nil
	}

	return getDefaultDesc(session, file)
}

func handleVideoFile(session *Session, file fileinfo.FileInfo) (string, error) {
	filePath := filepath.Join(file.Directory, file.Name)

	f, err := os.Open(filePath)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	//setting a display name
	opts := genai.UploadFileOptions{DisplayName: filepath.Base(filePath)}

	fmt.Print("uplaoding file..", file.Name, "\n")
	//Uploaded file to gemini
	vid, err := session.client.UploadFile(session.ctx, "", f, &opts)
	if err != nil {
		return getDefaultDesc(session, file)
	}
	fmt.Print("uplaoded file..", file.Name, "\n")

	// const baseDelay = 10 * time.Millisecond

	// Get metadata of the uploaded file

	uploadedFile, err := session.client.GetFile(session.ctx, vid.Name)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	for uploadedFile.State == genai.FileStateProcessing {
		fmt.Print(".")
		// Sleep for 10 seconds
		time.Sleep(10 * time.Second)

		// Fetch the file from the API again.
		uploadedFile, err = session.client.GetFile(session.ctx, file.Name)
		if err != nil {
			return getDefaultDesc(session, file)
		}
	}

	// fmt.Printf("state : %v", uploadedFile.State)
	if uploadedFile.State == 2 {

		model := session.client.GenerativeModel("gemini-1.5-flash")
		prompt := []genai.Part{
			genai.FileData{URI: uploadedFile.URI},
			genai.Text(fmt.Sprintf("Generate a description in less than 100 words about what this file is and what is it about, based on the first 7 seconds of the video file %s ", filepath.Base(filePath))),
		}

		resp, err := model.GenerateContent(session.ctx, prompt...)
		if err != nil {
			return getDefaultDesc(session, file)
		}

		if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
			var builder strings.Builder

			for _, candidate := range resp.Candidates {
				for _, part := range candidate.Content.Parts {
					builder.WriteString(fmt.Sprintf("%s", part))
				}
			}
			return builder.String(), nil
		}

		return getDefaultDesc(session, file)
	}

	return getDefaultDesc(session, file)

}

func GenerateEmbeddings(files []fileinfo.FileInfo) []fileinfo.FileInfo {

	ctx := context.Background()

	session, err := NewsetupSession(ctx)
	if err != nil {
		fmt.Println("Failed to start new setup session with Gemini:", err)
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
	fmt.Println("All goroutines have finished processing")

	var processedFiles []fileinfo.FileInfo
	for file := range resultCh {
		processedFiles = append(processedFiles, file)
	}
	return processedFiles
}

func GenerateEmbedding(setupSession *Session, desc string) ([]float32, error) {

	em := setupSession.client.EmbeddingModel("text-embedding-004")
	embeddingResult, err := em.EmbedContent(setupSession.ctx, genai.Text(desc))
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
			fmt.Printf("\n||Attempt no : %d||\n", attempt)
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
