package parser_test

import (
	"strings"
	"testing"

	"github.com/codesoap/gmir/parser"
)

// wrapTestCases are cases where input is a line that will be wrapped
// to the given width.
var wrapTestCases = []struct {
	input               string
	width               int
	expectedWrapIndexes []int
}{
	{
		"a b",
		3,
		[]int{},
	},
	{
		"a bc",
		3,
		[]int{2},
	},
	{
		"a b c",
		3,
		[]int{4},
	},
	{
		"abcd e",
		3,
		[]int{3},
	},
	{
		"=> ab c", // => c (ab)
		5,
		[]int{5, 7},
	},
	{
		">  ",
		3,
		[]int{},
	},
	{
		"日本語",
		1,
		[]int{3, 6},
	},
	{
		"日本語",
		3,
		[]int{3, 6},
	},
	{
		"日本語",
		4,
		[]int{6},
	},
	{
		"日 本語",
		3,
		[]int{4, 7},
	},
}

func TestWrap(t *testing.T) {
	for _, testCase := range wrapTestCases {
		t.Logf("Testing with '%s'.", testCase.input)
		lines, err := parser.Parse(strings.NewReader(testCase.input))
		if err != nil {
			t.Errorf("Could not parse input: %v", err)
			continue
		} else if len(lines) != 1 {
			t.Errorf("Found %d lines instead of one.", len(lines))
			continue
		}
		wrappable, ok := lines[0].(parser.WrappableLine)
		if !ok {
			t.Errorf("Given line is not wrappable.")
			continue
		}
		wrapIndexes := wrappable.WrapIndexes(testCase.width)
		if len(wrapIndexes) != len(testCase.expectedWrapIndexes) {
			t.Errorf("Wrong amount of wraps: Got %d, expected %d.",
				len(wrapIndexes),
				len(testCase.expectedWrapIndexes))
			continue
		}
		for i, wrapIndex := range wrapIndexes {
			if wrapIndex != testCase.expectedWrapIndexes[i] {
				t.Errorf("Got wrap index %d but expected %d.",
					wrapIndex,
					testCase.expectedWrapIndexes[i])
				continue
			}
		}
	}
}
