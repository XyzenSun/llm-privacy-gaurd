package llm

import (
	"encoding/json"
	"strings"
)

func ParseMaskEntries(raw string) ([]MaskEntry, error) {
	var entries []MaskEntry
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "[") {
		start := strings.Index(raw, "[")
		end := strings.LastIndex(raw, "]")
		if start != -1 && end != -1 && end > start {
			raw = raw[start : end+1]
		}
	}
	if err := json.Unmarshal([]byte(raw), &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
