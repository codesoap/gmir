package gmir

import (
	"math"

	"github.com/codesoap/gmir/parser"
	"github.com/gdamore/tcell/v2"
)

// Scroll scrolls up or down the given amount of (wrapped) lines. Scrolls
// up, if lines is negative. Never scrolls past the top or bottom line.
func (v *View) Scroll(screen tcell.Screen, lines int) {
	// TODO: Optimize, so that the same line is not wrapped multiple times.
	if lines == 0 {
		return
	}
	for lines > 0 && (v.line > 0 || v.lineOffset > 0) {
		if v.lineOffset > 0 {
			v.lineOffset--
		} else {
			v.line--
			v.lineOffset = v.maxLineOffset(screen, v.line)
		}
		lines--
	}
	for lines < 0 &&
		(v.line < len(v.lines)-1 || v.lineOffset < v.maxLineOffset(screen, v.line)) {
		if v.lineOffset == v.maxLineOffset(screen, v.line) {
			v.line++
			v.lineOffset = 0
		} else {
			v.lineOffset++
		}
		lines++
	}
	// TODO: Maybe ensure the last line will not scroll over the bottom of screen.
}

// ScrollToTop scrolls to the first line.
func (v *View) ScrollToTop(screen tcell.Screen) {
	v.line = 0
	v.lineOffset = 0
}

// ScrollToBottom scrolls to the last line.
func (v *View) ScrollToBottom(screen tcell.Screen) {
	v.line = len(v.lines) - 1
	v.lineOffset = v.maxLineOffset(screen, v.line)
}

// ScrollToNextHeading scrolls to the first line of the next heading.
func (v *View) ScrollToNextHeading(screen tcell.Screen) {
	if v.line == len(v.lines)-1 {
		return
	}
	for i, line := range v.lines[v.line+1:] {
		if isHeading(line) {
			v.line += i + 1
			v.lineOffset = 0
			return
		}
	}
}

// ScrollToPrevHeading scrolls to the first line of the previous heading.
func (v *View) ScrollToPrevHeading(screen tcell.Screen) {
	if v.line == 0 {
		return
	}
	for i := v.line - 1; i >= 0; i-- {
		if isHeading(v.lines[i]) {
			v.line = i
			v.lineOffset = 0
			return
		}
	}
}

// ScrollToNthHeading scrolls to the first line of the nth heading.
func (v *View) ScrollToNthHeading(screen tcell.Screen, n int) {
	for i, line := range v.lines {
		if isHeading(line) {
			n--
			if n < 0 {
				v.line = i
				v.lineOffset = 0
				return
			}
		}
	}
}

func isHeading(line parser.Line) bool {
	_, isHeading1 := line.(parser.Heading1Line)
	_, isHeading2 := line.(parser.Heading2Line)
	_, isHeading3 := line.(parser.Heading3Line)
	return isHeading1 || isHeading2 || isHeading3
}

// ScrollDownToSearchMatch scrolls to the next line, that matches
// v.Searchpattern. If the current line matches v.Searchpattern, nothing
// is done.
//
// Returns false, if neither the current line nor any line after matches
// v.Searchpattern.
func (v *View) ScrollDownToSearchMatch(screen tcell.Screen) bool {
	return v.scrollDownToSearchMatch(screen, false)
}

// ScrollDownToSearchMatch scrolls to the next line, that matches
// v.Searchpattern.
//
// Returns false, if none of the lines after the current one matches
// v.Searchpattern.
func (v *View) ScrollDownToNextSearchMatch(screen tcell.Screen) bool {
	return v.scrollDownToSearchMatch(screen, true)
}

// ScrollUpToSearchMatch scrolls to the previous line, that matches
// v.Searchpattern. If the current line matches v.Searchpattern, nothing
// is done.
//
// Returns false, if neither the current line nor any line after matches
// v.Searchpattern.
func (v *View) ScrollUpToSearchMatch(screen tcell.Screen) bool {
	return v.scrollUpToSearchMatch(screen, false)
}

// ScrollUpToSearchMatch scrolls to the previous line, that matches
// v.Searchpattern.
//
// Returns false, if none of the lines after the current one matches
// v.Searchpattern.
func (v *View) ScrollUpToNextSearchMatch(screen tcell.Screen) bool {
	return v.scrollUpToSearchMatch(screen, true)
}

func (v *View) scrollDownToSearchMatch(screen tcell.Screen, skipFirst bool) bool {
	if v.Searchpattern == nil {
		return false
	}
	screenWidth, _ := screen.Size()
	_, _, textWidth := v.columnWidths(screenWidth)
	for i, line := range v.lines[v.line:] {
		matches := v.Searchpattern.FindAllStringIndex(line.Text(), -1)
		if len(matches) == 0 {
			continue
		}
		if wrappable, isWrappable := line.(parser.WrappableLine); isWrappable {
			wrapIndexes := wrappable.WrapIndexes(textWidth)
			for _, offset := range lineOffsetsWithMatches(wrapIndexes, matches) {
				if i > 0 || (skipFirst && offset > v.lineOffset) || (!skipFirst && offset >= v.lineOffset) {
					v.line += i
					v.lineOffset = offset
					return true
				}
			}
		} else if (skipFirst && i > 0) || (!skipFirst && i >= 0) {
			v.line += i
			v.lineOffset = 0
			return true
		}
	}
	return false
}

func (v *View) scrollUpToSearchMatch(screen tcell.Screen, skipFirst bool) bool {
	if v.Searchpattern == nil {
		return false
	}
	screenWidth, _ := screen.Size()
	_, _, textWidth := v.columnWidths(screenWidth)
	for lineIndex := v.line; lineIndex >= 0; lineIndex-- {
		line := v.lines[lineIndex]
		matches := v.Searchpattern.FindAllStringIndex(line.Text(), -1)
		if len(matches) == 0 {
			continue
		}
		if wrappable, isWrappable := line.(parser.WrappableLine); isWrappable {
			wrapIndexes := wrappable.WrapIndexes(textWidth)
			matchingLineOffsets := lineOffsetsWithMatches(wrapIndexes, matches)
			for i := len(matchingLineOffsets) - 1; i >= 0; i-- {
				offset := matchingLineOffsets[i]
				if lineIndex < v.line || (skipFirst && offset < v.lineOffset) || (!skipFirst && offset <= v.lineOffset) {
					v.line = lineIndex
					v.lineOffset = offset
					return true
				}
			}
		} else if (skipFirst && lineIndex < v.line) || (!skipFirst && lineIndex <= v.line) {
			v.line = lineIndex
			v.lineOffset = 0
			return true
		}
	}
	return false
}

func lineOffsetsWithMatches(wrapIndexes []int, matches [][]int) []int {
	matchStops := append(wrapIndexes, math.MaxInt)
	offsets := make([]int, 0)
	offset := 0
	for offset < len(matchStops) {
		for _, match := range matches {
			if match[0] < matchStops[offset] &&
				(offset == 0 || match[0] >= matchStops[offset-1]) &&
				(len(offsets) == 0 || offsets[len(offsets)-1] != offset) {
				offsets = append(offsets, offset)
			}
		}
		offset++
	}
	return offsets
}
