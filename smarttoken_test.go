package gotoken

import (
	"fmt"
	"reflect"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
)

type tokenizerTestSet struct {
	input  string
	output map[string]SmartTokenInfo
}

func runTokenizerTestSetDepth(testSet []tokenizerTestSet, t *testing.T) {
	assert := assert.New(t)
	st := NewDepthTokenizer(10, 10, 18, 2)
	st.AddRangeTable(unicode.Latin)
	st.AddRangeTable(unicode.Cyrillic)

	for _, test := range testSet {
		result := st.TokenizeString(test.input)
		assert.True(reflect.DeepEqual(result, test.output), fmt.Sprintf("wrong tokenization of '%v' -> %v", test.input, result))
	}
}

func TestTokenizerSimple(t *testing.T) {
	testSet := []tokenizerTestSet{
		tokenizerTestSet{
			input: "hello", // Single word.
			output: map[string]SmartTokenInfo{
				"hello": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
			},
		},
		tokenizerTestSet{
			input: "hello world", // Token separation.
			output: map[string]SmartTokenInfo{
				"hello": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
				"world": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
			},
		},
	}
	runTokenizerTestSetDepth(testSet, t)
}

func TestTokenizerTransitions(t *testing.T) {
	testSet := []tokenizerTestSet{
		tokenizerTestSet{ // KnownLanguage -> KnownLanguage.
			input: "helloпривет",
			output: map[string]SmartTokenInfo{
				"hello":       SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
				"привет":      SmartTokenInfo{DetectedLanguage: 1, DetectedBase: [2]int{0, 12}},
				"helloпривет": SmartTokenInfo{DetectedLanguage: 1, DetectedBase: [2]int{0, 17}},
			},
		},
		tokenizerTestSet{ // KnownLanguage -> UnknownLanguage.
			input: "hello你好",
			output: map[string]SmartTokenInfo{
				"hello":   SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
				"你好":      SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 6}},
				"hello你好": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 11}}, // wtf??
			},
		},
		tokenizerTestSet{ // UnknownLanguage -> KnownLanguage.
			input: "你好привет",
			output: map[string]SmartTokenInfo{
				"你好":       SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 6}},
				"привет":   SmartTokenInfo{DetectedLanguage: 1, DetectedBase: [2]int{0, 12}},
				"你好привет": SmartTokenInfo{DetectedLanguage: 1, DetectedBase: [2]int{0, 18}},
			},
		},
		tokenizerTestSet{
			input: "hello123", // Language -> Number.
			output: map[string]SmartTokenInfo{
				"hello":    SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
				"123":      SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"hello123": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
			},
		},
		tokenizerTestSet{
			input: "hello...", // Language -> Delimiter.
			output: map[string]SmartTokenInfo{
				"hello":    SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
				"...":      SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"hello...": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
			},
		},
		tokenizerTestSet{
			input: "hello☭", // Language -> Other.
			output: map[string]SmartTokenInfo{
				"hello":  SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
				"☭":      SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"hello☭": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
			},
		},
		tokenizerTestSet{
			input: "123hello", // Number -> Language.
			output: map[string]SmartTokenInfo{
				"123":      SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"hello":    SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
				"123hello": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{3, 8}},
			},
		},
		tokenizerTestSet{
			input: "123...", // Number -> Delimiter.
			output: map[string]SmartTokenInfo{
				"123":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"...":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"123...": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
			},
		},
		tokenizerTestSet{
			input: "123☭", // Number -> Other.
			output: map[string]SmartTokenInfo{
				"123":  SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"☭":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"123☭": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
			},
		},
		tokenizerTestSet{
			input: "...hello", // Delimiter -> Language.
			output: map[string]SmartTokenInfo{
				"...":      SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"hello":    SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
				"...hello": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{3, 8}},
			},
		},
		tokenizerTestSet{
			input: "...123", // Delimiter -> Number.
			output: map[string]SmartTokenInfo{
				"...":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"123":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"...123": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
			},
		},
		tokenizerTestSet{
			input: "...☭", // Delimiter -> Other.
			output: map[string]SmartTokenInfo{
				"...":  SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"☭":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"...☭": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
			},
		},
		tokenizerTestSet{
			input: "☭hello", // Other -> Language.
			output: map[string]SmartTokenInfo{
				"☭":      SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"hello":  SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 5}},
				"☭hello": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{3, 8}},
			},
		},
		tokenizerTestSet{
			input: "☭123", // Other -> Number.
			output: map[string]SmartTokenInfo{
				"☭":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"123":  SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"☭123": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
			},
		},
		tokenizerTestSet{
			input: "☭...", // Other -> Delimiter.
			output: map[string]SmartTokenInfo{
				"☭":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"...":  SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"☭...": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
			},
		},
	}
	runTokenizerTestSetDepth(testSet, t)
}

func TestTokenizerLanguages(t *testing.T) {
	testSet := []tokenizerTestSet{
		tokenizerTestSet{ // Similar language separation.
			input: "aаaа",
			output: map[string]SmartTokenInfo{
				"a":    SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 1}},
				"а":    SmartTokenInfo{DetectedLanguage: 1, DetectedBase: [2]int{0, 2}},
				"aа":   SmartTokenInfo{DetectedLanguage: 1, DetectedBase: [2]int{0, 3}},
				"аa":   SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"aаa":  SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"аaа":  SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"aаaа": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
			},
		},
		tokenizerTestSet{
			input: "你好。再见。", // Chinese punctuation.
			output: map[string]SmartTokenInfo{
				"你好":     SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 6}},
				"再见":     SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 6}},
				"。":      SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"你好。":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 6}},
				"再见。":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 6}},
				"。再见":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{3, 9}},
				"你好。再见":  SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 15}},
				"。再见。":   SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{3, 9}},
				"你好。再见。": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
			},
		},
	}
	runTokenizerTestSetDepth(testSet, t)
}

func TestTokenizerDepth(t *testing.T) {
	testSet := []tokenizerTestSet{
		tokenizerTestSet{
			input: "a.b.c.d", // Left side.
			output: map[string]SmartTokenInfo{
				"a":       SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 1}},
				"b":       SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 1}},
				"c":       SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 1}},
				"d":       SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 1}},
				".":       SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"a.":      SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 1}},
				"b.":      SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 1}},
				"c.":      SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 1}},
				".b":      SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{1, 2}},
				".c":      SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{1, 2}},
				".d":      SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{1, 2}},
				"a.b":     SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"b.c":     SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"c.d":     SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				".b.":     SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{1, 2}},
				".c.":     SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{1, 2}},
				"a.b.":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"b.c.":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				".b.c":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				".c.d":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"a.b.c":   SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"b.c.d":   SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				".b.c.":   SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"a.b.c.":  SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				".b.c.d":  SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"a.b.c.d": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
			},
		},
		tokenizerTestSet{
			input: "aaa.bbb.ccc.ddd", // Middle side.
			output: map[string]SmartTokenInfo{
				"aaa":         SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"bbb":         SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"ccc":         SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"ddd":         SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				".":           SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"aaa.":        SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"bbb.":        SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"ccc.":        SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				".bbb":        SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{1, 4}},
				".ccc":        SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{1, 4}},
				".ddd":        SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{1, 4}},
				"aaa.bbb":     SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 7}},
				"bbb.ccc":     SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 7}},
				"ccc.ddd":     SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 7}},
				".bbb.":       SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{1, 4}},
				".ccc.":       SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{1, 4}},
				"aaa.bbb.":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"bbb.ccc.":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				".bbb.ccc":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				".ccc.ddd":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"aaa.bbb.ccc": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"bbb.ccc.ddd": SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				".bbb.ccc.":   SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
			},
		},
		tokenizerTestSet{
			input: "aaa...bbb...ccc...ddd", // Right side.
			output: map[string]SmartTokenInfo{
				"aaa":    SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"bbb":    SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"ccc":    SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"ddd":    SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"...":    SmartTokenInfo{DetectedLanguage: -1, DetectedBase: [2]int{0, 0}},
				"aaa...": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"bbb...": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"ccc...": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{0, 3}},
				"...bbb": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{3, 6}},
				"...ccc": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{3, 6}},
				"...ddd": SmartTokenInfo{DetectedLanguage: 0, DetectedBase: [2]int{3, 6}},
			},
		},
	}
	runTokenizerTestSetDepth(testSet, t)
}
