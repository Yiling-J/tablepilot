package client

import (
	"fmt"

	"github.com/invopop/jsonschema"
)

type ContentType string

const (
	ContentTypeText  ContentType = "text"
	ContentTypeImage ContentType = "image"
)

type Content struct {
	Type ContentType
	Data string
}

type Message struct {
	Role    string
	Content []Content
}

func UserMessage(content string) *Message {
	return &Message{
		Role: "user",
		Content: []Content{
			{Type: ContentTypeText, Data: content},
		},
	}
}

func UserMessageWithImages(content string, images map[string]string) *Message {
	m := &Message{
		Role: "user",
		Content: []Content{
			{Type: ContentTypeText, Data: content},
		},
	}
	for id, img := range images {
		m.Content = append(m.Content, Content{
			Type: ContentTypeText,
			Data: fmt.Sprintf("\nBelow is the image with ID: <%s>", id),
		})
		m.Content = append(m.Content, Content{
			Type: ContentTypeImage,
			Data: img,
		})
	}
	return m
}

func AssistantMessage(content string) *Message {
	return &Message{
		Role: "assistant",
		Content: []Content{
			{Type: ContentTypeText, Data: content},
		},
	}
}

func SystemMessage(content string) *Message {
	return &Message{
		Role: "system",
		Content: []Content{
			{Type: ContentTypeText, Data: content},
		},
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
