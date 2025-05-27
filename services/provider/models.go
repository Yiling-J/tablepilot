package provider

type Model struct {
	Model     string `json:"model"`
	Alias     string `json:"alias"`
	Client    string `json:"client"`
	MaxTokens int64  `json:"max_tokens"`
	Rpm       int    `json:"rpm"`
	Default   bool   `json:"default"`
	Image     bool   `json:"image"`
}

type ProviderType string

const (
	ProviderTypeOpenAI           ProviderType = "OpenAI"
	ProviderTypeOpenGemini       ProviderType = "Gemini"
	ProviderTypeOpenAIcompatible ProviderType = "OpenAI-Compatible"
	ProviderTypeAnthropic        ProviderType = "Anthropic"
	ProviderOpenRouter           ProviderType = "OpenRouter"
)

type Provider struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Key      string  `json:"key"`
	BaseURL  string  `json:"base_url"`
	Models   []Model `json:"models"`
	Editable bool    `json:"editable"`
	Enabled  bool    `json:"enabled"`
}
