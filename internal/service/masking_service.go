package service

import (
	"llm-privacy-gaurd/internal/domain"
	"llm-privacy-gaurd/internal/llm"
	"llm-privacy-gaurd/internal/repository"
	"sort"
)

type MaskingService struct {
	mappingRepo *repository.MaskMappingRepository
	trustedLLM  llm.TrustedLLMClient
}

func NewMaskingService(mappingRepo *repository.MaskMappingRepository, trustedLLM llm.TrustedLLMClient) *MaskingService {
	return &MaskingService{
		mappingRepo: mappingRepo,
		trustedLLM:  trustedLLM,
	}
}

type MaskResult struct {
	MaskedContent string
	ByPlaceholder map[string]string
	NewEntries    []domain.MaskEntry
}

func (s *MaskingService) Mask(sessionID, content string, knownMappings map[string]string, turnID int) (*MaskResult, error) {
	preMasked := s.applyKnownMappings(content, knownMappings)
	//禁用正则简单判断，提升效果
	// if s.looksFullyCovered(preMasked) {
	// 	return &MaskResult{
	// 		MaskedContent: preMasked,
	// 		ByPlaceholder: knownMappings,
	// 	}, nil
	// }

	entries, err := s.trustedLLM.Detect(content, knownMappings, nil, preMasked)
	if err != nil {
		return nil, err
	}

	delta := s.validateEntries(entries, knownMappings)
	if len(delta.NewEntries) == 0 {
		return &MaskResult{
			MaskedContent: preMasked,
			ByPlaceholder: knownMappings,
		}, nil
	}

	updated := make(map[string]string)
	for k, v := range knownMappings {
		updated[k] = v
	}
	for _, e := range delta.NewEntries {
		updated[e.Original] = e.Placeholder
	}

	finalMasked := s.applyKnownMappings(content, updated)

	dbMappings := make([]domain.MaskMapping, len(delta.NewEntries))
	for i, e := range delta.NewEntries {
		dbMappings[i] = domain.MaskMapping{
			ID:            newID(),
			SessionID:     sessionID,
			Original:      e.Original,
			Placeholder:   e.Placeholder,
			Type:          e.Type,
			FirstSeenTurn: turnID,
			LastSeenTurn:  turnID,
		}
	}
	if err := s.mappingRepo.UpsertMany(dbMappings); err != nil {
		return nil, err
	}

	return &MaskResult{
		MaskedContent: finalMasked,
		ByPlaceholder: updated,
		NewEntries:    delta.NewEntries,
	}, nil
}

func (s *MaskingService) Unmask(content string, byPlaceholder map[string]string) string {
	items := make([]struct {
		placeholder string
		original    string
	}, 0, len(byPlaceholder))
	for p, o := range byPlaceholder {
		items = append(items, struct {
			placeholder string
			original    string
		}{p, o})
	}
	sort.Slice(items, func(i, j int) bool {
		return len(items[i].placeholder) > len(items[j].placeholder)
	})

	result := content
	for _, item := range items {
		result = replaceAll(result, item.placeholder, item.original)
	}
	return result
}

func (s *MaskingService) GetMappings(sessionID string) (map[string]string, error) {
	mappings, err := s.mappingRepo.ListBySession(sessionID)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, m := range mappings {
		result[m.Original] = m.Placeholder
	}
	return result, nil
}

func (s *MaskingService) applyKnownMappings(content string, mappings map[string]string) string {
	items := make([]struct {
		original    string
		placeholder string
	}, 0, len(mappings))
	for o, p := range mappings {
		items = append(items, struct {
			original    string
			placeholder string
		}{o, p})
	}
	sort.Slice(items, func(i, j int) bool {
		return len(items[i].original) > len(items[j].original)
	})

	result := content
	for _, item := range items {
		result = replaceAll(result, item.original, item.placeholder)
	}
	return result
}
//简易的正则判断，判断是否为邮箱、手机号、身份证号、银行卡号、IP地址，降低了识别效果，暂时废弃
func (s *MaskingService) looksFullyCovered(content string) bool {
	emailPattern := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	phonePattern := `1[3-9]\d{9}`
	idPattern := `\d{17}[\dXx]`
	bankPattern := `\d{16,19}`
	ipPattern := `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`

	for _, pattern := range []string{emailPattern, phonePattern, idPattern, bankPattern, ipPattern} {
		if matchPattern(content, pattern) {
			return false
		}
	}
	return true
}

func (s *MaskingService) validateEntries(entries []llm.MaskEntry, knownMappings map[string]string) *domain.MaskDelta {
	delta := &domain.MaskDelta{NewEntries: []domain.MaskEntry{}}
	for _, e := range entries {
		if e.Original == "" || e.Placeholder == "" || e.Type == "" {
			continue
		}
		if existing, exists := knownMappings[e.Original]; exists {
			if existing != e.Placeholder {
				return nil
			}
			continue
		}
		delta.NewEntries = append(delta.NewEntries, domain.MaskEntry{
			Original:    e.Original,
			Placeholder: e.Placeholder,
			Type:        e.Type,
		})
	}
	return delta
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := findString(s, old)
		if idx == -1 {
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result + s
}

func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func matchPattern(s, pattern string) bool {
	for i := 0; i <= len(s)-len(pattern); i++ {
		match := true
		for j, c := range pattern {
			if c == '.' || c == '+' || c == '*' || c == '[' || c == ']' || c == '-' || c == '\\' {
				continue
			}
			if i+j >= len(s) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}