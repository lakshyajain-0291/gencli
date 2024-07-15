package gemini

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gemini_cli_tool/fileinfo"

	"github.com/dslipak/pdf"

	"github.com/google/generative-ai-go/genai"
)

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

func GenerateDescription(file fileinfo.FileInfo) (string, error) {

	ctx := context.Background()

	session, err := NewsetupSession(ctx)
	if err != nil {
		return "No description", fmt.Errorf("failed to start new setup session with gemini : %w", err)
	}
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

	description, err := timeOut(20*time.Second, descriptionFunc)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	return description, nil
}

func getDefaultDesc(session *Session, file fileinfo.FileInfo) (string, error) {

	model := session.client.GenerativeModel("gemini-1.5-pro")
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

	model := session.client.GenerativeModel("gemini-1.5-pro")
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

	model := session.client.GenerativeModel("gemini-1.5-pro")
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

	model := session.client.GenerativeModel("gemini-1.5-pro")
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

	//Uploaded file to gemini
	vid, err := session.client.UploadFile(session.ctx, "", f, &opts)
	if err != nil {
		return getDefaultDesc(session, file)
	}

	const maxRetries = 20
	const baseDelay = 10 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		// Get metadata of the uploaded file

		uploadedFile, err := session.client.GetFile(session.ctx, vid.Name)
		if err != nil {
			return getDefaultDesc(session, file)
		}

		// fmt.Printf("state : %v", uploadedFile.State)
		if uploadedFile.State == 2 {

			model := session.client.GenerativeModel("gemini-1.5-pro")
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

		delay := time.Duration(math.Pow(2, float64(i))) * baseDelay
		time.Sleep(delay)
	}

	return getDefaultDesc(session, file)

}

func GenerateEmbeddings(desc string) ([]float32, error) {
	ctx := context.Background()

	setupSession, err := NewsetupSession(ctx)
	if err != nil {
		return nil, err
	}

	em := setupSession.client.EmbeddingModel("text-embedding-004")
	embeddingResult, err := em.EmbedContent(ctx, genai.Text(desc))
	if err != nil {
		return nil, err
	}

	// fmt.Println(embeddingResult.Embedding.Values)
	return embeddingResult.Embedding.Values, nil

}
