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

func GenerateDescription(file fileinfo.FileInfo) (string, error) {

	ctx := context.Background()

	session, err := NewsetupSession(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start new setup session with gemini : %w", err)
	}

	filePath := filepath.Join(file.Directory, file.Name)

	mimeType := mime.TypeByExtension(filepath.Ext(filePath))

	fmt.Printf("\nmimeType : %s \n", mimeType)

	switch {
	case strings.HasPrefix(mimeType, "text/"):
		return handleTextFile(session, filePath)
	case strings.HasSuffix(mimeType, "/pdf"):
		return handlePdfFile(session, filePath)
	case strings.HasPrefix(mimeType, "image/"):
		return handleImageFile(session, filePath)
	case strings.HasPrefix(mimeType, "video/"):
		return handleVideoFile(session, filePath)
	default:
		return getDefaultDesc(session, file)
	}
}

func getDefaultDesc(session *Session, file fileinfo.FileInfo) (string, error) {

	model := session.client.GenerativeModel("gemini-1.5-pro")
	prompt := []genai.Part{
		genai.Text(fmt.Sprintf("Generate a description in less than 100 words about what this file is and what is it about, based on the meta data given : \n\nFile Name : %s\nFile Directory : %s\nFile Size : %d\nFile Modified Time : %v", file.Name, file.Directory, file.Size, file.ModifiedTime)),
	}

	resp, err := model.GenerateContent(session.ctx, prompt...)
	if err != nil {
		return "", err
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

	return "", fmt.Errorf("no description generated")

}

func handleTextFile(session *Session, filePath string) (string, error) {

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	model := session.client.GenerativeModel("gemini-1.5-pro")
	prompt := []genai.Part{
		genai.Text(fmt.Sprintf("Generate a description in less than 100 words about what this file is and what is it about, based on the given content : \n\n%s", string(content))),
	}

	resp, err := model.GenerateContent(session.ctx, prompt...)
	if err != nil {
		return "", err
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

	return "", fmt.Errorf("no description generated")
}

func handlePdfFile(session *Session, filePath string) (string, error) {
	r, err := pdf.Open(filePath)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)
	content := buf.String()

	model := session.client.GenerativeModel("gemini-1.5-pro")
	prompt := []genai.Part{
		genai.Text(fmt.Sprintf("Generate a description in less than 100 words about what this file is and what is it about, based on the given content : \n\n%s", content)),
	}

	resp, err := model.GenerateContent(session.ctx, prompt...)
	if err != nil {
		return "", err
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

	return "", fmt.Errorf("no description generated")

}

func handleImageFile(session *Session, filePath string) (string, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open the file : %w", err)
	}

	//setting a display name
	opts := genai.UploadFileOptions{DisplayName: filepath.Base(filePath)}

	//Uploaded file to gemini
	img, err := session.client.UploadFile(session.ctx, "", f, &opts)
	if err != nil {
		return "", err
	}

	//Got metadata of the uploaded file
	uploadedFile, err := session.client.GetFile(session.ctx, img.Name)
	if err != nil {
		return "", err
	}

	model := session.client.GenerativeModel("gemini-1.5-pro")
	prompt := []genai.Part{
		genai.FileData{URI: uploadedFile.URI},
		genai.Text(fmt.Sprintf("Generate a description in less than 100 words about what this file %s is and what is it about ", filepath.Base(filePath))),
	}

	resp, err := model.GenerateContent(session.ctx, prompt...)
	if err != nil {
		return "", err
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

	return "", fmt.Errorf("no description generated")
}

func handleVideoFile(session *Session, filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open the file : %w", err)
	}

	//setting a display name
	opts := genai.UploadFileOptions{DisplayName: filepath.Base(filePath)}

	//Uploaded file to gemini
	vid, err := session.client.UploadFile(session.ctx, "", f, &opts)
	if err != nil {
		return "", err
	}

	const maxRetries = 20
	const baseDelay = 10 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		// Get metadata of the uploaded file

		uploadedFile, err := session.client.GetFile(session.ctx, vid.Name)
		if err != nil {
			return "", err
		}

		fmt.Printf("state : %v", uploadedFile.State)
		if uploadedFile.State == 2 {

			model := session.client.GenerativeModel("gemini-1.5-pro")
			prompt := []genai.Part{
				genai.FileData{URI: uploadedFile.URI},
				genai.Text(fmt.Sprintf("Generate a description in less than 100 words about what this file is and what is it about, based on the first 7 seconds of the video file %s ", filepath.Base(filePath))),
			}

			resp, err := model.GenerateContent(session.ctx, prompt...)
			if err != nil {
				return "", err
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

			return "", fmt.Errorf("no description generated")
		}

		delay := time.Duration(math.Pow(2, float64(i))) * baseDelay
		time.Sleep(delay)
	}

	return "", fmt.Errorf("file %s is not in an ACTIVE state after %d retries", vid.Name, maxRetries)

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
