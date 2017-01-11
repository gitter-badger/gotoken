package gotoken

import (
	"fmt"
	"reflect"
	"testing"
	"unicode"

	"github.com/rvncerr/goassert"
)

func TestGetDepth(t *testing.T) {

	const leftX = 20
	const leftY = 100
	const rightX = 120
	const rightY = 10

	ga := goassert.New(t)

	var st SmartToken
	st.SetDepthPolicy(leftX, leftY, rightX, rightY)

	for i := 1; i <= leftX; i++ {
		ga.Assert(st.getDepth(i) == leftY, "depth policy left part")
	}

	memory := leftY
	for i := leftX + 1; i <= rightX; i++ {
		ga.Assert(st.getDepth(i) <= memory, "depth policy middle part")
		memory = st.getDepth(i)
	}

	for i := rightX; i <= rightX+20; i++ {
		ga.Assert(st.getDepth(i) == rightY, "depth policy right part")
	}
}

type tokenizerTestSet struct {
	input  string
	output map[string]bool
}

func TestTokenizeString(t *testing.T) {
	ga := goassert.New(t)

	var st SmartToken
	st.AddRangeTable(unicode.Latin)
	st.AddRangeTable(unicode.Cyrillic)
	st.SetDepthPolicy(10, 10, 18, 2)

	testSet := []tokenizerTestSet{
		tokenizerTestSet{
			input: "hello", // Single word.
			output: map[string]bool{
				"hello": true,
			},
		},
		tokenizerTestSet{
			input: "hello world", // Token separation.
			output: map[string]bool{
				"hello": true,
				"world": true,
			},
		},
		tokenizerTestSet{ // KnownLanguage -> KnownLanguage.
			input: "helloпривет",
			output: map[string]bool{
				"hello":       true,
				"привет":      true,
				"helloпривет": true,
			},
		},
		tokenizerTestSet{ // KnownLanguage -> UnknownLanguage.
			input: "hello你好",
			output: map[string]bool{
				"hello":   true,
				"你好":      true,
				"hello你好": true,
			},
		},
		tokenizerTestSet{ // UnknownLanguage -> KnownLanguage.
			input: "你好привет",
			output: map[string]bool{
				"你好":       true,
				"привет":   true,
				"你好привет": true,
			},
		},
		tokenizerTestSet{ // Similar language separation.
			input: "aаaа",
			output: map[string]bool{
				"a":    true,
				"а":    true,
				"aа":   true,
				"аa":   true,
				"aаa":  true,
				"аaа":  true,
				"aаaа": true,
			},
		},
		tokenizerTestSet{
			input: "hello123", // Language -> Number.
			output: map[string]bool{
				"hello":    true,
				"123":      true,
				"hello123": true,
			},
		},
		tokenizerTestSet{
			input: "hello...", // Language -> Delimiter.
			output: map[string]bool{
				"hello":    true,
				"...":      true,
				"hello...": true,
			},
		},
		tokenizerTestSet{
			input: "hello☭", // Language -> Other.
			output: map[string]bool{
				"hello":  true,
				"☭":      true,
				"hello☭": true,
			},
		},
		tokenizerTestSet{
			input: "123hello", // Number -> Language.
			output: map[string]bool{
				"123":      true,
				"hello":    true,
				"123hello": true,
			},
		},
		tokenizerTestSet{
			input: "123...", // Number -> Delimiter.
			output: map[string]bool{
				"123":    true,
				"...":    true,
				"123...": true,
			},
		},
		tokenizerTestSet{
			input: "123☭", // Number -> Other.
			output: map[string]bool{
				"123":  true,
				"☭":    true,
				"123☭": true,
			},
		},
		tokenizerTestSet{
			input: "...hello", // Delimiter -> Language.
			output: map[string]bool{
				"...":      true,
				"hello":    true,
				"...hello": true,
			},
		},
		tokenizerTestSet{
			input: "...123", // Delimiter -> Number.
			output: map[string]bool{
				"...":    true,
				"123":    true,
				"...123": true,
			},
		},
		tokenizerTestSet{
			input: "...☭", // Delimiter -> Other.
			output: map[string]bool{
				"...":  true,
				"☭":    true,
				"...☭": true,
			},
		},
		tokenizerTestSet{
			input: "☭hello", // Other -> Language.
			output: map[string]bool{
				"☭":      true,
				"hello":  true,
				"☭hello": true,
			},
		},
		tokenizerTestSet{
			input: "☭123", // Other -> Number.
			output: map[string]bool{
				"☭":    true,
				"123":  true,
				"☭123": true,
			},
		},
		tokenizerTestSet{
			input: "☭...", // Other -> Delimiter.
			output: map[string]bool{
				"☭":    true,
				"...":  true,
				"☭...": true,
			},
		},
		tokenizerTestSet{
			input: "a.b.c.d",
			output: map[string]bool{
				"a":       true,
				"b":       true,
				"c":       true,
				"d":       true,
				".":       true,
				"a.":      true,
				"b.":      true,
				"c.":      true,
				".b":      true,
				".c":      true,
				".d":      true,
				"a.b":     true,
				"b.c":     true,
				"c.d":     true,
				".b.":     true,
				".c.":     true,
				"a.b.":    true,
				"b.c.":    true,
				".b.c":    true,
				".c.d":    true,
				"a.b.c":   true,
				"b.c.d":   true,
				".b.c.":   true,
				"a.b.c.":  true,
				".b.c.d":  true,
				"a.b.c.d": true,
			},
		},
		tokenizerTestSet{
			input: "aaa.bbb.ccc.ddd",
			output: map[string]bool{
				"aaa":         true,
				"bbb":         true,
				"ccc":         true,
				"ddd":         true,
				".":           true,
				"aaa.":        true,
				"bbb.":        true,
				"ccc.":        true,
				".bbb":        true,
				".ccc":        true,
				".ddd":        true,
				"aaa.bbb":     true,
				"bbb.ccc":     true,
				"ccc.ddd":     true,
				".bbb.":       true,
				".ccc.":       true,
				"aaa.bbb.":    true,
				"bbb.ccc.":    true,
				".bbb.ccc":    true,
				".ccc.ddd":    true,
				"aaa.bbb.ccc": true,
				"bbb.ccc.ddd": true,
				".bbb.ccc.":   true,
			},
		},
		tokenizerTestSet{
			input: "aaa...bbb...ccc...ddd",
			output: map[string]bool{
				"aaa":    true,
				"bbb":    true,
				"ccc":    true,
				"ddd":    true,
				"...":    true,
				"aaa...": true,
				"bbb...": true,
				"ccc...": true,
				"...bbb": true,
				"...ccc": true,
				"...ddd": true,
			},
		},
	}

	for _, test := range testSet {
		result := st.TokenizeString(test.input)
		ga.Assert(reflect.DeepEqual(result, test.output), fmt.Sprintf("wrong tokenization of '%v' -> %v", test.input, result))
	}
}
