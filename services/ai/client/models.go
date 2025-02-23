package client

import "github.com/invopop/jsonschema"

type Message struct {
	Role    string
	Content string
}

func UserMessage(content string) *Message {
	return &Message{
		Role:    "user",
		Content: content,
	}
}

func AssistantMessage(content string) *Message {
	return &Message{
		Role:    "assistant",
		Content: content,
	}
}

func SystemMessage(content string) *Message {
	return &Message{
		Role:    "system",
		Content: content,
	}
}

type ChatRequest struct {
	Messages        []*Message
	Temperature     float64
	Schema          *jsonschema.Schema
	Model           string
	MaxOutputTokens int64
	PresencePenalty float64
}

type ChatResponse struct {
	Content string `json:"content"`
	Tokens  int64  `json:"tokens"`
}

type ImageGenerateRequest struct {
	Prompt string
	Model  string
}

type ImageResponse struct {
	URL     string `json:"url"`
	Content string `json:"content"`
	Tokens  int    `json:"tokens"`
}
