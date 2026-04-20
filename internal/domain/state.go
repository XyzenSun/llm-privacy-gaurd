package domain

type SessionState struct {
	OriginalHistory []Message
	MaskedHistory   []Message
	ByOriginal      map[string]string
	ByPlaceholder   map[string]string
	Entries         []MaskEntry
	MessageCache    map[string]CachedMaskResult
	TurnID          int
	MapVersion      int
}

type MaskEntry struct {
	Original      string `json:"original"`
	Placeholder   string `json:"placeholder"`
	Type          string `json:"type"`
	FirstSeenTurn int    `json:"firstSeenTurn"`
	LastSeenTurn  int    `json:"lastSeenTurn"`
}

type CachedMaskResult struct {
	MessageHash    string      `json:"messageHash"`
	MapVersion     int         `json:"mapVersion"`
	MaskedContent  string      `json:"maskedContent"`
	DeltaEntries   []MaskEntry `json:"deltaEntries"`
}

type MaskDelta struct {
	NewEntries []MaskEntry `json:"newEntries"`
}

type LocalLlmResponse struct {
	RawText string
	OK      bool
}
