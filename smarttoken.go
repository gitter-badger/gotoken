package gotoken

import (
	"unicode"
	"unicode/utf8"
)

const runeClassUndef = -1
const runeClassLetter = 0
const runeClassDigit = 1
const runeClassPunct = 2
const runeClassOther = 3

const policyDepth = 0
const policyCount = 1

const indexPolicy = 0
const indexMaxLength = 1
const indexMaxDepth = 2
const indexMinLength = 3
const indexMinDepth = 4
const indexMaxCount = 1

// SmartToken is a tokenizer for SmartToken algorithm.
type SmartToken struct {
	rangeTableList  []*unicode.RangeTable
	rangeTableIndex int
	runeClass       int

	policy []int
}

// AddRangeTable pushes new language into tokenizer.
func (st *SmartToken) AddRangeTable(rt *unicode.RangeTable) {
	st.rangeTableList = append(st.rangeTableList, rt)
}

// SetDepthPolicy tells tokenizer to use depth-based tokenization policy.
func (st *SmartToken) SetDepthPolicy(maxLength int, maxDepth int, minLength int, minDepth int) {
	st.policy = []int{policyDepth, maxLength, maxDepth, minLength, minDepth}
}

// SetCountPolicy tells tokenizer to use count-based tokenization policy.
func (st *SmartToken) SetCountPolicy(maxCount int) {
	st.policy = []int{policyCount, maxCount}
}

func (st *SmartToken) getDepth(length int) int {
	if st.policy[indexPolicy] == policyDepth {
		if length <= st.policy[indexMaxLength] {
			return st.policy[indexMaxDepth]
		} else if length >= st.policy[indexMinLength] {
			return st.policy[indexMinDepth]
		}
		return st.policy[indexMinDepth] +
			(st.policy[indexMaxDepth]-st.policy[indexMinDepth])*
				(st.policy[indexMinLength]-length)/
				(st.policy[indexMinLength]-st.policy[indexMaxLength])
	}
	return length
}

func (st *SmartToken) flush() {
	st.runeClass = runeClassUndef
	st.rangeTableIndex = -1
}

func (st *SmartToken) getRuneClass(r rune) int {
	switch {
	case unicode.IsLetter(r):
		return runeClassLetter
	case unicode.IsDigit(r):
		return runeClassDigit
	case unicode.IsPunct(r):
		return runeClassPunct
	}
	return runeClassOther
}

func (st *SmartToken) getTableIndex(r rune) int {
	for index, rangeTable := range st.rangeTableList {
		if unicode.In(r, rangeTable) {
			return index
		}
	}
	return -1
}

func (st *SmartToken) pushRune(r rune) bool {
	result := false
	newRuneClass := st.getRuneClass(r)

	if newRuneClass == runeClassLetter {
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

func (st *SmartToken) getSubtokens(token string, tokens map[string]int, distribution map[int]int) {
	st.flush()
	cb := makeCircularBuffer(st.getDepth(utf8.RuneCountInString(token)) + 1)
	for index, r := range token {
		if st.pushRune(r) {
			cb.push(index)
			if cb.full() {
				left, rightArray := cb.extract()
				for depth, right := range rightArray {
					tokens[token[left:right]] = depth
					distribution[depth]++
				}
			}
		}
	}
	cb.push(len(token))
	for !cb.empty() {
		left, rightArray := cb.extract()
		for depth, right := range rightArray {
			tokens[token[left:right]] = depth
			distribution[depth]++
		}
		cb.pop()
	}
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
