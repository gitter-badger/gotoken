package gotoken

import "unicode"

const runeClassUndef = -1
const runeClassLetter = 0
const runeClassDigit = 1
const runeClassPunct = 2

type SmartToken struct {
	rangeTableList  []*unicode.RangeTable
	rangeTableIndex int
	runeClass       int
}

func (st *SmartToken) AddRangeTable(rt *unicode.RangeTable) {
	st.rangeTableList = append(st.rangeTableList, rt)
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
	return runeClassUndef
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

func (st *SmartToken) GetSubtokensWithDepth(token string, depth int) map[string]bool {
	st.flush()
	subTokens := make(map[string]bool)
	cb := makeCircularBuffer(depth + 1)
	for index, r := range token {
		if st.pushRune(r) {
			cb.push(index)
			if cb.full() {
				left, rightArray := cb.extract()
				for _, right := range rightArray {
					subTokens[token[left:right]] = true
				}
			}
		}
	}
	cb.push(len(token))
	for !cb.empty() {
		left, rightArray := cb.extract()
		for _, right := range rightArray {
			subTokens[token[left:right]] = true
		}
		cb.pop()
	}
	return subTokens
}
