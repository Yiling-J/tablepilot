package provider

type Model struct {
	Model     string `json:"model"`
	Alias     string `json:"alias"`
	Client    string `json:"client"`
	MaxTokens int64  `json:"max_tokens"`
	Rpm       int    `json:"rpm"`
	Default   bool   `json:"default"`
	Image     bool   `bool:"image"`
}

type Provider struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Key      string  `json:"key"`
	BaseURL  string  `json:"base_url"`
	Models   []Model `json:"models"`
	Editable bool    `json:"editable"`
}
