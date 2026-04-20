package llm

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type TrustedLLMClient interface {
	Detect(prompt string, knownMappings map[string]string, recentContext []Message, newMessage string) ([]MaskEntry, error)
}

type CloudLLMClient interface {
	Chat(messages []Message) (string, error)
}

type MaskEntry struct {
	Original    string `json:"original"`
	Placeholder string `json:"placeholder"`
	Type        string `json:"type"`
}