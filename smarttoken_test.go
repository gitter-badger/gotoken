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
				"hello": SmartTokenInfo{DetectedLanguage: 0},
			},
		},
		tokenizerTestSet{
			input: "hello world", // Token separation.
			output: map[string]SmartTokenInfo{
				"hello": SmartTokenInfo{DetectedLanguage: 0},
				"world": SmartTokenInfo{DetectedLanguage: 0},
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
				"hello":       SmartTokenInfo{DetectedLanguage: 0},
				"привет":      SmartTokenInfo{DetectedLanguage: 1},
				"helloпривет": SmartTokenInfo{DetectedLanguage: 1},
			},
		},
		tokenizerTestSet{ // KnownLanguage -> UnknownLanguage.
			input: "hello你好",
			output: map[string]SmartTokenInfo{
				"hello":   SmartTokenInfo{DetectedLanguage: 0},
				"你好":      SmartTokenInfo{DetectedLanguage: -1},
				"hello你好": SmartTokenInfo{DetectedLanguage: -1},
			},
		},
		tokenizerTestSet{ // UnknownLanguage -> KnownLanguage.
			input: "你好привет",
			output: map[string]SmartTokenInfo{
				"你好":       SmartTokenInfo{DetectedLanguage: -1},
				"привет":   SmartTokenInfo{DetectedLanguage: 1},
				"你好привет": SmartTokenInfo{DetectedLanguage: 1},
			},
		},
		tokenizerTestSet{
			input: "hello123", // Language -> Number.
			output: map[string]SmartTokenInfo{
				"hello":    SmartTokenInfo{DetectedLanguage: 0},
				"123":      SmartTokenInfo{DetectedLanguage: -1},
				"hello123": SmartTokenInfo{DetectedLanguage: 0},
			},
		},
		tokenizerTestSet{
			input: "hello...", // Language -> Delimiter.
			output: map[string]SmartTokenInfo{
				"hello":    SmartTokenInfo{DetectedLanguage: 0},
				"...":      SmartTokenInfo{DetectedLanguage: -1},
				"hello...": SmartTokenInfo{DetectedLanguage: 0},
			},
		},
		tokenizerTestSet{
			input: "hello☭", // Language -> Other.
			output: map[string]SmartTokenInfo{
				"hello":  SmartTokenInfo{DetectedLanguage: 0},
				"☭":      SmartTokenInfo{DetectedLanguage: -1},
				"hello☭": SmartTokenInfo{DetectedLanguage: 0},
			},
		},
		tokenizerTestSet{
			input: "123hello", // Number -> Language.
			output: map[string]SmartTokenInfo{
				"123":      SmartTokenInfo{DetectedLanguage: -1},
				"hello":    SmartTokenInfo{DetectedLanguage: 0},
				"123hello": SmartTokenInfo{DetectedLanguage: 0},
			},
		},
		tokenizerTestSet{
			input: "123...", // Number -> Delimiter.
			output: map[string]SmartTokenInfo{
				"123":    SmartTokenInfo{DetectedLanguage: -1},
				"...":    SmartTokenInfo{DetectedLanguage: -1},
				"123...": SmartTokenInfo{DetectedLanguage: -1},
			},
		},
		tokenizerTestSet{
			input: "123☭", // Number -> Other.
			output: map[string]SmartTokenInfo{
				"123":  SmartTokenInfo{DetectedLanguage: -1},
				"☭":    SmartTokenInfo{DetectedLanguage: -1},
				"123☭": SmartTokenInfo{DetectedLanguage: -1},
			},
		},
		tokenizerTestSet{
			input: "...hello", // Delimiter -> Language.
			output: map[string]SmartTokenInfo{
				"...":      SmartTokenInfo{DetectedLanguage: -1},
				"hello":    SmartTokenInfo{DetectedLanguage: 0},
				"...hello": SmartTokenInfo{DetectedLanguage: 0},
			},
		},
		tokenizerTestSet{
			input: "...123", // Delimiter -> Number.
			output: map[string]SmartTokenInfo{
				"...":    SmartTokenInfo{DetectedLanguage: -1},
				"123":    SmartTokenInfo{DetectedLanguage: -1},
				"...123": SmartTokenInfo{DetectedLanguage: -1},
			},
		},
		tokenizerTestSet{
			input: "...☭", // Delimiter -> Other.
			output: map[string]SmartTokenInfo{
				"...":  SmartTokenInfo{DetectedLanguage: -1},
				"☭":    SmartTokenInfo{DetectedLanguage: -1},
				"...☭": SmartTokenInfo{DetectedLanguage: -1},
			},
		},
		tokenizerTestSet{
			input: "☭hello", // Other -> Language.
			output: map[string]SmartTokenInfo{
				"☭":      SmartTokenInfo{DetectedLanguage: -1},
				"hello":  SmartTokenInfo{DetectedLanguage: 0},
				"☭hello": SmartTokenInfo{DetectedLanguage: 0},
			},
		},
		tokenizerTestSet{
			input: "☭123", // Other -> Number.
			output: map[string]SmartTokenInfo{
				"☭":    SmartTokenInfo{DetectedLanguage: -1},
				"123":  SmartTokenInfo{DetectedLanguage: -1},
				"☭123": SmartTokenInfo{DetectedLanguage: -1},
			},
		},
		tokenizerTestSet{
			input: "☭...", // Other -> Delimiter.
			output: map[string]SmartTokenInfo{
				"☭":    SmartTokenInfo{DetectedLanguage: -1},
				"...":  SmartTokenInfo{DetectedLanguage: -1},
				"☭...": SmartTokenInfo{DetectedLanguage: -1},
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
				"a":    SmartTokenInfo{DetectedLanguage: 0},
				"а":    SmartTokenInfo{DetectedLanguage: 1},
				"aа":   SmartTokenInfo{DetectedLanguage: 1},
				"аa":   SmartTokenInfo{DetectedLanguage: 0},
				"aаa":  SmartTokenInfo{DetectedLanguage: -1},
				"аaа":  SmartTokenInfo{DetectedLanguage: -1},
				"aаaа": SmartTokenInfo{DetectedLanguage: -1},
			},
		},
		tokenizerTestSet{
			input: "你好。再见。", // Chinese punctuation.
			output: map[string]SmartTokenInfo{
				"你好":     SmartTokenInfo{DetectedLanguage: -1},
				"再见":     SmartTokenInfo{DetectedLanguage: -1},
				"。":      SmartTokenInfo{DetectedLanguage: -1},
				"你好。":    SmartTokenInfo{DetectedLanguage: -1},
				"再见。":    SmartTokenInfo{DetectedLanguage: -1},
				"。再见":    SmartTokenInfo{DetectedLanguage: -1},
				"你好。再见":  SmartTokenInfo{DetectedLanguage: -1},
				"。再见。":   SmartTokenInfo{DetectedLanguage: -1},
				"你好。再见。": SmartTokenInfo{DetectedLanguage: -1},
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
				"a":       SmartTokenInfo{DetectedLanguage: 0},
				"b":       SmartTokenInfo{DetectedLanguage: 0},
				"c":       SmartTokenInfo{DetectedLanguage: 0},
				"d":       SmartTokenInfo{DetectedLanguage: 0},
				".":       SmartTokenInfo{DetectedLanguage: -1},
				"a.":      SmartTokenInfo{DetectedLanguage: 0},
				"b.":      SmartTokenInfo{DetectedLanguage: 0},
				"c.":      SmartTokenInfo{DetectedLanguage: 0},
				".b":      SmartTokenInfo{DetectedLanguage: 0},
				".c":      SmartTokenInfo{DetectedLanguage: 0},
				".d":      SmartTokenInfo{DetectedLanguage: 0},
				"a.b":     SmartTokenInfo{DetectedLanguage: 0},
				"b.c":     SmartTokenInfo{DetectedLanguage: 0},
				"c.d":     SmartTokenInfo{DetectedLanguage: 0},
				".b.":     SmartTokenInfo{DetectedLanguage: 0},
				".c.":     SmartTokenInfo{DetectedLanguage: 0},
				"a.b.":    SmartTokenInfo{DetectedLanguage: -1},
				"b.c.":    SmartTokenInfo{DetectedLanguage: -1},
				".b.c":    SmartTokenInfo{DetectedLanguage: -1},
				".c.d":    SmartTokenInfo{DetectedLanguage: -1},
				"a.b.c":   SmartTokenInfo{DetectedLanguage: -1},
				"b.c.d":   SmartTokenInfo{DetectedLanguage: -1},
				".b.c.":   SmartTokenInfo{DetectedLanguage: -1},
				"a.b.c.":  SmartTokenInfo{DetectedLanguage: -1},
				".b.c.d":  SmartTokenInfo{DetectedLanguage: -1},
				"a.b.c.d": SmartTokenInfo{DetectedLanguage: -1},
			},
		},
		tokenizerTestSet{
			input: "aaa.bbb.ccc.ddd", // Middle side.
			output: map[string]SmartTokenInfo{
				"aaa":         SmartTokenInfo{DetectedLanguage: 0},
				"bbb":         SmartTokenInfo{DetectedLanguage: 0},
				"ccc":         SmartTokenInfo{DetectedLanguage: 0},
				"ddd":         SmartTokenInfo{DetectedLanguage: 0},
				".":           SmartTokenInfo{DetectedLanguage: -1},
				"aaa.":        SmartTokenInfo{DetectedLanguage: 0},
				"bbb.":        SmartTokenInfo{DetectedLanguage: 0},
				"ccc.":        SmartTokenInfo{DetectedLanguage: 0},
				".bbb":        SmartTokenInfo{DetectedLanguage: 0},
				".ccc":        SmartTokenInfo{DetectedLanguage: 0},
				".ddd":        SmartTokenInfo{DetectedLanguage: 0},
				"aaa.bbb":     SmartTokenInfo{DetectedLanguage: 0},
				"bbb.ccc":     SmartTokenInfo{DetectedLanguage: 0},
				"ccc.ddd":     SmartTokenInfo{DetectedLanguage: 0},
				".bbb.":       SmartTokenInfo{DetectedLanguage: 0},
				".ccc.":       SmartTokenInfo{DetectedLanguage: 0},
				"aaa.bbb.":    SmartTokenInfo{DetectedLanguage: -1},
				"bbb.ccc.":    SmartTokenInfo{DetectedLanguage: -1},
				".bbb.ccc":    SmartTokenInfo{DetectedLanguage: -1},
				".ccc.ddd":    SmartTokenInfo{DetectedLanguage: -1},
				"aaa.bbb.ccc": SmartTokenInfo{DetectedLanguage: -1},
				"bbb.ccc.ddd": SmartTokenInfo{DetectedLanguage: -1},
				".bbb.ccc.":   SmartTokenInfo{DetectedLanguage: -1},
			},
		},
		tokenizerTestSet{
			input: "aaa...bbb...ccc...ddd", // Right side.
			output: map[string]SmartTokenInfo{
				"aaa":    SmartTokenInfo{DetectedLanguage: 0},
				"bbb":    SmartTokenInfo{DetectedLanguage: 0},
				"ccc":    SmartTokenInfo{DetectedLanguage: 0},
				"ddd":    SmartTokenInfo{DetectedLanguage: 0},
				"...":    SmartTokenInfo{DetectedLanguage: -1},
				"aaa...": SmartTokenInfo{DetectedLanguage: 0},
				"bbb...": SmartTokenInfo{DetectedLanguage: 0},
				"ccc...": SmartTokenInfo{DetectedLanguage: 0},
				"...bbb": SmartTokenInfo{DetectedLanguage: 0},
				"...ccc": SmartTokenInfo{DetectedLanguage: 0},
				"...ddd": SmartTokenInfo{DetectedLanguage: 0},
			},
		},
	}
	runTokenizerTestSetDepth(testSet, t)
}
