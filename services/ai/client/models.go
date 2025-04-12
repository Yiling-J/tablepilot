package client

import (
	"fmt"
	"sort"

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

func sortAndRun(m map[string]string, fn func(k string, v string)) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fn(k, m[k])
	}
}

func UserMessageWithImages(content string, images map[string]string) *Message {
	m := &Message{
		Role: "user",
		Content: []Content{
			{Type: ContentTypeText, Data: content},
		},
	}
	sortAndRun(images, func(k, v string) {
		m.Content = append(m.Content, Content{
			Type: ContentTypeText,
			Data: fmt.Sprintf("\nBelow is the image with ID: <%s>", k),
		})
		m.Content = append(m.Content, Content{
			Type: ContentTypeImage,
			Data: v,
		})
	})
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
	ImageModel      string
	MaxOutputTokens int64
	PresencePenalty float64
}

type ChatResponse struct {
	Content string `json:"content"`
	Tokens  int64  `json:"tokens"`
}

type ImageGenResponse struct {
	Images map[string][]byte `json:"images"` // map[image id]image b64 data
	Tokens int64             `json:"tokens"`
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
