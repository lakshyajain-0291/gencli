package gemini

import (
	"context"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Session struct {
	ctx     context.Context
	client  *genai.Client
	session *genai.ChatSession
}

func NewsetupSession(ctx context.Context) (*Session, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &Session{
		ctx:     ctx,
		client:  client,
		session: client.GenerativeModel("gemini-flash").StartChat(),
	}, nil
}

// SendMessage sends a request to the model as part of a setup session.
func (c *Session) SendMessage(input string) (*genai.GenerateContentResponse, error) {
	return c.session.SendMessage(c.ctx, genai.Text(input))
}

// SendMessageStream is like SendMessage, but with a streaming request.
func (c *Session) SendMessageStream(input string) *genai.GenerateContentResponseIterator {
	return c.session.SendMessageStream(c.ctx, genai.Text(input))
}

// ClearHistory clears setup history
func (c *Session) ClearHistory() {
	c.session.History = make([]*genai.Content, 0)
}

// Close closes the genai.Client
func (c *Session) Close() error {
	return c.client.Close()
}
