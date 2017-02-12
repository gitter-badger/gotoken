package gotoken

import (
	"unicode"
	"unicode/utf8"

	"github.com/rvncerr/gocontainers"
)

//go:generate stringer -tsype=RuneClass -output=rune_class_string_gen.go
type RuneClass int

const (
	Undef  RuneClass = -1
	Letter RuneClass = iota
	Digit
	Punct
	Other
)

// SmartToken is a tokenizer for SmartToken algorithm.
type SmartToken struct {
	rangeTableList  []*unicode.RangeTable
	rangeTableIndex int
	runeClass       RuneClass
	policy          TokenizationPolicy
}

func NewDepthTokenizer(maxLength int, maxDepth int, minLength int, minDepth int) *SmartToken {
	policy := NewPolicyDepth(maxLength, maxDepth, minLength, minDepth)
	return &SmartToken{
		policy: policy,
	}
}

// AddRangeTable pushes new language into tokenizer.
func (st *SmartToken) AddRangeTable(rt *unicode.RangeTable) {
	st.rangeTableList = append(st.rangeTableList, rt)
}

// SetPolicy tells tokenizer how to calculate token sizes
func (st *SmartToken) SetPolicy(p TokenizationPolicy) {
	st.policy = p
}

// TokenizeString starts SmartToken tokenization process on a string.
func (st *SmartToken) TokenizeString(source string) map[string]int {
	tokens := make(map[string]int)    // Token -> Depth.
	distribution := make(map[int]int) // Depth -> Count(Token).

	const stateSpace = 0
	const stateToken = 1

	offset := 0
	state := stateSpace
	for index, r := range source {
		switch state {
		case stateSpace:
			if !unicode.IsSpace(r) {
				state = stateToken
				offset = index
			}
			break
		case stateToken:
			if unicode.IsSpace(r) {
				state = stateSpace
				st.getSubtokens(source[offset:index], tokens, distribution)
			}
			break
		}
	}
	if state == stateToken {
		st.getSubtokens(source[offset:], tokens, distribution)
	}
	return tokens
}

func (st *SmartToken) getSubtokens(token string, tokens map[string]int, distribution map[int]int) {
	st.flush()
	depth := st.policy.GetDepth(utf8.RuneCountInString(token)) + 1
	cb := gocontainers.NewCircularBuffer(depth)
	for index, r := range token {
		if st.pushRune(r) {
			cb.PushBack(index)
			if cb.Full() {
				array := cb.ToArray()
				left := cb.Front()
				for depth, right := range array[1:] {
					tokens[token[left.(int):right.(int)]] = depth
					distribution[depth]++
				}
			}
		}
	}

	cb.PushBack(len(token))
	for !cb.Empty() {
		array := cb.ToArray()
		left := cb.Front()
		for depth, right := range array[1:] {
			tokens[token[left.(int):right.(int)]] = depth
			distribution[depth]++
		}
		cb.PopFront()
	}
}

func (st *SmartToken) flush() {
	st.runeClass = Undef
	st.rangeTableIndex = -1
}

func (st *SmartToken) pushRune(r rune) bool {
	result := false
	newRuneClass := st.getRuneClass(r)

	if newRuneClass == Letter {
		newRangeTableIndex := st.getTableIndex(r)
		if newRangeTableIndex != st.rangeTableIndex {
			st.rangeTableIndex = newRangeTableIndex
			result = true
		}
	}
	if newRuneClass != st.runeClass {
		st.runeClass = newRuneClass
		result = true
	}
	return result
}

func (st *SmartToken) getRuneClass(r rune) RuneClass {
	switch {
	case unicode.IsLetter(r):
		return Letter
	case unicode.IsDigit(r):
		return Digit
	case unicode.IsPunct(r):
		return Punct
	}
	return Other
}

func (st *SmartToken) getTableIndex(r rune) int {
	for index, rangeTable := range st.rangeTableList {
		if unicode.In(r, rangeTable) {
			return index
		}
	}
	return -1
}
