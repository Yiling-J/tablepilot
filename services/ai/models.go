package ai

import "golang.org/x/time/rate"

type model struct {
	model     string
	alias     string
	client    string
	maxTokens int64
	limiter   *rate.Limiter
	image     bool
}

type ModelListItem struct {
	Name  string `json:"name"`
	Image bool   `json:"image"`
}

type ModelList struct {
	Models            []ModelListItem `json:"models"`
	DefaultModel      string          `json:"default_model"`
	DefaultImageModel string          `json:"default_image_model"`
}

type GenerateListOptionsRequest struct {
	Model   string   `json:"model"`
	Prompt  string   `json:"prompt"`
	Options []string `json:"options"`
}
