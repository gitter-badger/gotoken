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
	output map[string]int
}

func TestTokenizeStringDepthPolicy(t *testing.T) {
	ga := goassert.New(t)

	var st SmartToken
	st.AddRangeTable(unicode.Latin)
	st.AddRangeTable(unicode.Cyrillic)
	st.SetDepthPolicy(10, 10, 18, 2)

	testSet := []tokenizerTestSet{
		tokenizerTestSet{
			input: "hello", // Single word.
			output: map[string]int{
				"hello": 0,
			},
		},
		tokenizerTestSet{
			input: "hello world", // Token separation.
			output: map[string]int{
				"hello": 0,
				"world": 0,
			},
		},
		tokenizerTestSet{ // KnownLanguage -> KnownLanguage.
			input: "helloпривет",
			output: map[string]int{
				"hello":       0,
				"привет":      0,
				"helloпривет": 1,
			},
		},
		tokenizerTestSet{ // KnownLanguage -> UnknownLanguage.
			input: "hello你好",
			output: map[string]int{
				"hello":   0,
				"你好":      0,
				"hello你好": 1,
			},
		},
		tokenizerTestSet{ // UnknownLanguage -> KnownLanguage.
			input: "你好привет",
			output: map[string]int{
				"你好":       0,
				"привет":   0,
				"你好привет": 1,
			},
		},
		tokenizerTestSet{ // Similar language separation.
			input: "aаaа",
			output: map[string]int{
				"a":    0,
				"а":    0,
				"aа":   1,
				"аa":   1,
				"aаa":  2,
				"аaа":  2,
				"aаaа": 3,
			},
		},
		tokenizerTestSet{
			input: "hello123", // Language -> Number.
			output: map[string]int{
				"hello":    0,
				"123":      0,
				"hello123": 1,
			},
		},
		tokenizerTestSet{
			input: "hello...", // Language -> Delimiter.
			output: map[string]int{
				"hello":    0,
				"...":      0,
				"hello...": 1,
			},
		},
		tokenizerTestSet{
			input: "hello☭", // Language -> Other.
			output: map[string]int{
				"hello":  0,
				"☭":      0,
				"hello☭": 1,
			},
		},
		tokenizerTestSet{
			input: "123hello", // Number -> Language.
			output: map[string]int{
				"123":      0,
				"hello":    0,
				"123hello": 1,
			},
		},
		tokenizerTestSet{
			input: "123...", // Number -> Delimiter.
			output: map[string]int{
				"123":    0,
				"...":    0,
				"123...": 1,
			},
		},
		tokenizerTestSet{
			input: "123☭", // Number -> Other.
			output: map[string]int{
				"123":  0,
				"☭":    0,
				"123☭": 1,
			},
		},
		tokenizerTestSet{
			input: "...hello", // Delimiter -> Language.
			output: map[string]int{
				"...":      0,
				"hello":    0,
				"...hello": 1,
			},
		},
		tokenizerTestSet{
			input: "...123", // Delimiter -> Number.
			output: map[string]int{
				"...":    0,
				"123":    0,
				"...123": 1,
			},
		},
		tokenizerTestSet{
			input: "...☭", // Delimiter -> Other.
			output: map[string]int{
				"...":  0,
				"☭":    0,
				"...☭": 1,
			},
		},
		tokenizerTestSet{
			input: "☭hello", // Other -> Language.
			output: map[string]int{
				"☭":      0,
				"hello":  0,
				"☭hello": 1,
			},
		},
		tokenizerTestSet{
			input: "☭123", // Other -> Number.
			output: map[string]int{
				"☭":    0,
				"123":  0,
				"☭123": 1,
			},
		},
		tokenizerTestSet{
			input: "☭...", // Other -> Delimiter.
			output: map[string]int{
				"☭":    0,
				"...":  0,
				"☭...": 1,
			},
		},
		tokenizerTestSet{
			input: "a.b.c.d", // Left side.
			output: map[string]int{
				"a":       0,
				"b":       0,
				"c":       0,
				"d":       0,
				".":       0,
				"a.":      1,
				"b.":      1,
				"c.":      1,
				".b":      1,
				".c":      1,
				".d":      1,
				"a.b":     2,
				"b.c":     2,
				"c.d":     2,
				".b.":     2,
				".c.":     2,
				"a.b.":    3,
				"b.c.":    3,
				".b.c":    3,
				".c.d":    3,
				"a.b.c":   4,
				"b.c.d":   4,
				".b.c.":   4,
				"a.b.c.":  5,
				".b.c.d":  5,
				"a.b.c.d": 6,
			},
		},
		tokenizerTestSet{
			input: "aaa.bbb.ccc.ddd", // Middle side.
			output: map[string]int{
				"aaa":         0,
				"bbb":         0,
				"ccc":         0,
				"ddd":         0,
				".":           0,
				"aaa.":        1,
				"bbb.":        1,
				"ccc.":        1,
				".bbb":        1,
				".ccc":        1,
				".ddd":        1,
				"aaa.bbb":     2,
				"bbb.ccc":     2,
				"ccc.ddd":     2,
				".bbb.":       2,
				".ccc.":       2,
				"aaa.bbb.":    3,
				"bbb.ccc.":    3,
				".bbb.ccc":    3,
				".ccc.ddd":    3,
				"aaa.bbb.ccc": 4,
				"bbb.ccc.ddd": 4,
				".bbb.ccc.":   4,
			},
		},
		tokenizerTestSet{
			input: "aaa...bbb...ccc...ddd", // Right side.
			output: map[string]int{
				"aaa":    0,
				"bbb":    0,
				"ccc":    0,
				"ddd":    0,
				"...":    0,
				"aaa...": 1,
				"bbb...": 1,
				"ccc...": 1,
				"...bbb": 1,
				"...ccc": 1,
				"...ddd": 1,
			},
		},
	}

	for _, test := range testSet {
		result := st.TokenizeString(test.input)
		ga.Assert(reflect.DeepEqual(result, test.output), fmt.Sprintf("wrong tokenization of '%v' -> %v", test.input, result))
	}
}
