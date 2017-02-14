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

// SmartTokenInfo provides basic information about the token.
type SmartTokenInfo struct {
	DetectedLanguage int
	DetectedBase     [2]int
}

// SmartToken is a tokenizer for SmartToken algorithm.
type SmartToken struct {
	rangeTableList          []*unicode.RangeTable
	previousRangeTableIndex int
	currentRangeTableIndex  int
	previousRuneClass       RuneClass
	currentRuneClass        RuneClass
	policy                  TokenizationPolicy
}

func (st *SmartToken) detectBase(bs []interface{}, left int, right int) [2]int {
	return [2]int{bs[left].(int) - bs[0].(int), bs[right].(int) - bs[0].(int)}
}

func (st *SmartToken) detectLanguage(bs []interface{}, rc []interface{}, rt []interface{}) (int, [2]int) {
	// fmt.Printf("%v %v\n", rc, rt)
	length := len(rc)
	switch {
	case length == 1:
		if rc[0] == Letter {
			return rt[0].(int), st.detectBase(bs, 0, 1) // "hello", "привет"
		}
		return -1, [2]int{0, 0} // "123"
	case length == 2:
		switch {
		case rc[0] == Letter && rc[1] == Letter:
			return rt[1].(int), st.detectBase(bs, 0, 2) // "mailка"
		case rc[0] == Letter && rc[1] != Letter:
			return rt[0].(int), st.detectBase(bs, 0, 1) // "привет---"
		case rc[0] != Letter && rc[1] == Letter:
			return rt[1].(int), st.detectBase(bs, 1, 2) // "---привет"
		default:
			return -1, [2]int{0, 0} // "---123"
		}
	case length == 3:
		switch {
		case rc[0] == Letter && rc[1] == Letter && rc[2] == Letter:
			return -1, [2]int{0, 0} // "mailприветhello"
		case rc[0] == Letter && rc[1] == Letter && rc[2] != Letter:
			return rt[1].(int), st.detectBase(bs, 0, 2) // "mailка---"
		case rc[0] == Letter && rc[1] != Letter && rc[2] == Letter:
			if rt[0] == rt[2] {
				return rt[2].(int), st.detectBase(bs, 0, 3) // "карабас-барабас"
			}
			return rt[2].(int), st.detectBase(bs, 2, 3) // "css-стили"
		case rc[0] != Letter && rc[1] == Letter && rc[2] == Letter:
			return rt[2].(int), st.detectBase(bs, 2, 3) // "---mailка"
		case rc[0] == Letter && rc[1] != Letter && rc[2] != Letter:
			return rt[0].(int), st.detectBase(bs, 0, 1) // "привет---123"
		case rc[0] != Letter && rc[1] == Letter && rc[2] != Letter:
			return rt[1].(int), st.detectBase(bs, 1, 2) // "---привет---"
		case rc[0] != Letter && rc[1] != Letter && rc[2] == Letter:
			return rt[2].(int), st.detectBase(bs, 2, 3) // "123---привет"
		default:
			return -1, [2]int{0, 0} // "123---123"
		}
	default:
		return -1, [2]int{0, 0}
	}
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
func (st *SmartToken) TokenizeString(source string) map[string]SmartTokenInfo {
	tokens := make(map[string]SmartTokenInfo) // Token -> Info.
	// distribution := make(map[int]int) // Depth -> Count(Token).

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
				st.getSubtokens(source[offset:index], tokens)
			}
			break
		}
	}
	if state == stateToken {
		st.getSubtokens(source[offset:], tokens)
	}
	return tokens
}

func (st *SmartToken) getSubtokens(token string, tokens map[string]SmartTokenInfo) {
	st.flush()
	depth := st.policy.GetDepth(utf8.RuneCountInString(token)) + 1
	blockSizeBuffer := gocontainers.NewCircularBuffer(depth)
	runeClassBuffer := gocontainers.NewCircularBuffer(depth - 1)
	rangeTableBuffer := gocontainers.NewCircularBuffer(depth - 1)

	for index, r := range token {
		if st.pushRune(r) {
			blockSizeBuffer.PushBack(index)
			if st.previousRuneClass != Undef {
				runeClassBuffer.PushBack(st.previousRuneClass)
				if st.previousRuneClass == Letter {
					rangeTableBuffer.PushBack(st.previousRangeTableIndex)
				} else {
					rangeTableBuffer.PushBack(-1)
				}
			}
			if blockSizeBuffer.Full() {
				array := blockSizeBuffer.ToArray()
				left := blockSizeBuffer.Front()
				for depth, right := range array[1:] {
					var info SmartTokenInfo
					info.DetectedLanguage, info.DetectedBase =
						st.detectLanguage(array, runeClassBuffer.ToArray()[0:depth+1], rangeTableBuffer.ToArray()[0:depth+1])
					tokens[token[left.(int):right.(int)]] = info
				}
			}
		}
	}

	blockSizeBuffer.PushBack(len(token))
	runeClassBuffer.PushBack(st.currentRuneClass)
	if st.currentRuneClass == Letter {
		rangeTableBuffer.PushBack(st.currentRangeTableIndex)
	} else {
		rangeTableBuffer.PushBack(-1)
	}

	// fmt.Printf("token = %v, bs = %v, rc = %v, rt = %v\n", token, blockSizeBuffer.ToArray(), runeClassBuffer.ToArray(), rangeTableBuffer.ToArray())

	for !blockSizeBuffer.Empty() {
		array := blockSizeBuffer.ToArray()
		left := blockSizeBuffer.Front()
		for depth, right := range array[1:] {
			var info SmartTokenInfo
			info.DetectedLanguage, info.DetectedBase =
				st.detectLanguage(array, runeClassBuffer.ToArray()[0:depth+1], rangeTableBuffer.ToArray()[0:depth+1])
			tokens[token[left.(int):right.(int)]] = info
		}
		blockSizeBuffer.PopFront()
		runeClassBuffer.PopFront()
		rangeTableBuffer.PopFront()
	}
}

func (st *SmartToken) flush() {
	st.previousRuneClass = Undef
	st.currentRuneClass = Undef
	st.previousRangeTableIndex = -1
	st.currentRangeTableIndex = -1
}

// very dirty!!!
func (st *SmartToken) pushRune(r rune) bool {
	result := false
	newRuneClass := st.getRuneClass(r)

	if newRuneClass == Letter {
		newRangeTableIndex := st.getTableIndex(r)
		if newRangeTableIndex != st.currentRangeTableIndex {
			st.previousRangeTableIndex = st.currentRangeTableIndex
			st.currentRangeTableIndex = newRangeTableIndex
			st.previousRuneClass = st.currentRuneClass
			st.currentRuneClass = newRuneClass
			result = true
		} else if newRuneClass != st.currentRuneClass {
			st.previousRangeTableIndex = st.currentRangeTableIndex
			st.currentRangeTableIndex = -1
			st.previousRuneClass = st.currentRuneClass
			st.currentRuneClass = newRuneClass
			result = true
		}
	} else if newRuneClass != st.currentRuneClass {
		st.previousRangeTableIndex = st.currentRangeTableIndex
		st.currentRangeTableIndex = -1
		st.previousRuneClass = st.currentRuneClass
		st.currentRuneClass = newRuneClass
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
