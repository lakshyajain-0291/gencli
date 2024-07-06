package gemini

import (
	"context"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ChatSession struct {
	ctx     context.Context
	client  *genai.Client
	session *genai.ChatSession
}

func NewChatSession(ctx context.Context) (*ChatSession, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return &ChatSession{
		ctx:     ctx,
		client:  client,
		session: client.GenerativeModel("gemini-pro").StartChat(),
	}, nil
}

// SendMessage sends a request to the model as part of a chat session.
func (c *ChatSession) SendMessage(input string) (*genai.GenerateContentResponse, error) {
	return c.session.SendMessage(c.ctx, genai.Text(input))
}

// SendMessageStream is like SendMessage, but with a streaming request.
func (c *ChatSession) SendMessageStream(input string) *genai.GenerateContentResponseIterator {
	return c.session.SendMessageStream(c.ctx, genai.Text(input))
}

// ClearHistory clears chat history
func (c *ChatSession) ClearHistory() {
	c.session.History = make([]*genai.Content, 0)
}

// Close closes the genai.Client
func (c *ChatSession) Close() error {
	return c.client.Close()
}
