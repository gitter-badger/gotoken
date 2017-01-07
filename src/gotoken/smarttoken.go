package gotoken

import (
	"strings"
	"unicode/utf8"
)

// SmartToken - Tokenizer for "SmartToken" algorithm.
type SmartToken struct {
	spaces    string
	languages []string
	limit     int
}

const undefinedLanguage = -2
const unidentifiedLanguage = -1

func (t *SmartToken) getLanguage(r rune) int {
	for position, language := range t.languages {
		if strings.ContainsRune(language, r) {
			return position
		}
	}
	return unidentifiedLanguage
}

func (t *SmartToken) getSubTokensLightweight(token []rune, subtokens map[string]bool) {
	currentLanguage := undefinedLanguage
	startPosition := -1
	for position, r := range token {
		detectedLanguage := t.getLanguage(r)
		if detectedLanguage != currentLanguage {
			if startPosition != -1 {
				subtokens[string(token[startPosition:position])] = true
			}
			startPosition = position
			currentLanguage = detectedLanguage
		}
	}
	subtokens[string(token[startPosition:])] = true
}

func (t *SmartToken) getSubTokensRegular(token []rune, subtokens map[string]bool) {
	if len(token) != 0 {
		currentLanguage := undefinedLanguage
		for position, r := range token {
			detectedLanguage := t.getLanguage(r)
			if detectedLanguage != currentLanguage {
				subtokens[string(token[:position])] = true
				if currentLanguage != -2 {
					t.getSubTokensRegular(token[position:], subtokens)
				}
				currentLanguage = detectedLanguage
			}
		}
		subtokens[string(token)] = true
	}
}

func (t *SmartToken) getSubTokens(token string, subtokens map[string]bool) {
	if utf8.RuneCountInString(token) > t.limit {
		t.getSubTokensLightweight([]rune(token), subtokens)
	} else {
		t.getSubTokensRegular([]rune(token), subtokens)
	}
}

// GetTokens - Extract tokens from the string.
func (t *SmartToken) GetTokens(source string) map[string]bool {
	answer := make(map[string]bool)
	tokens := strings.FieldsFunc(source, func(r rune) bool {
		return strings.ContainsRune(t.spaces, r)
	})
	for _, token := range tokens {
		t.getSubTokens(token, answer)
	}
	return answer
}
