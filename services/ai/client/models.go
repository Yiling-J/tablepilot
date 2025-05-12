package client

import (
	"fmt"
	"sort"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go"
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

func UserMessageWithSingleImage(content string, image string) *Message {
	m := &Message{
		Role: "user",
		Content: []Content{
			{Type: ContentTypeText, Data: content},
		},
	}
	m.Content = append(m.Content, Content{
		Type: ContentTypeImage,
		Data: image,
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

type ChatToolParam struct {
	Name        string
	Type        string
	Enum        []string
	Description string
}

type ChatTool struct {
	Name        string
	Description string
	Parameters  []ChatToolParam
}

func (c *ChatTool) OpenAITool() openai.ChatCompletionToolParam {
	props := map[string]any{}
	required := []string{}
	for _, param := range c.Parameters {
		p := map[string]any{"description": param.Description}
		switch param.Type {
		case "[]string":
			p["type"] = "array"
			p["items"] = map[string]string{
				"type": "string",
			}
		default:
			p["type"] = param.Type
		}
		if len(param.Enum) > 0 {
			p["enum"] = param.Enum
		}
		required = append(required, param.Name)
		props[param.Name] = p
	}
	return openai.ChatCompletionToolParam{
		Function: openai.FunctionDefinitionParam{
			Name:        c.Name,
			Description: openai.String(c.Description),
			Parameters: openai.FunctionParameters{
				"type":       "object",
				"properties": props,
				"required":   required,
			},
		},
	}
}

type ChatRequest struct {
	Messages        []*Message
	Temperature     float64
	Schema          *jsonschema.Schema
	Tools           []ChatTool
	Model           string
	ImageModel      string
	MaxOutputTokens int64
	PresencePenalty float64
}

type ChatResponse struct {
	Content string `json:"content"`
	Tokens  int64  `json:"tokens"`
}

type FunctionCall struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type FunctionCallResponse struct {
	Text          string         `json:"text"`
	FunctionCalls []FunctionCall `json:"function_calls"`
	Tokens        int64          `json:"tokens"`
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
