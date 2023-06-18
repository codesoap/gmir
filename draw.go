package gmir

import (
	"strings"

	"github.com/codesoap/gmir/parser"
	"github.com/codesoap/gmir/selector"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

/*
Draw draws the given view to screen. The screen is separated like this:

	   7cols    11cols     10cols          18cols
	┌───────┬───────────┬──────────┬──────────────────┐
	│ left  │ selectors │ rendered │    right space   │
	│ space │           │ gmi      │                  │
	│       │           │          │                  │
	├───────┴───────────┴──────────┴──────────────────┤
	│ bar                                             │
	└─────────────────────────────────────────────────┘

The rendered gmi aims to be a width that is comfortable to read. The
right space may be intruded by preformatted lines. The selector column
will always be wide enough to fit the largest selector in the document.
*/
func (v View) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()
	leftSpace, selectorColWidth, textWidth := v.columnWidths(screenWidth)
	if screenWidth < selectorColWidth+8 || screenHeight < 2 {
		// Screen too small.
		return
	}
	maxColOffset := v.maxLineWidth(textWidth) - textWidth
	if v.ColOffset > maxColOffset {
		leftSpace -= maxColOffset
	} else {
		leftSpace -= v.ColOffset
	}
	v.drawSelectorAndGMIColumn(screen, leftSpace, selectorColWidth, textWidth)
	v.drawBar(screen)
	screen.Show()
}

func (v View) maxLineWidth(wrappedTextWidth int) int {
	maxLineWidth := wrappedTextWidth
	for _, line := range v.lines {
		if _, isWrappable := line.(parser.WrappableLine); !isWrappable {
			if lineWidth := runewidth.StringWidth(line.Text()); lineWidth > maxLineWidth {
				maxLineWidth = lineWidth
			}
		}
	}
	return maxLineWidth
}

func (v View) drawSelectorAndGMIColumn(screen tcell.Screen, offset, selectorColWidth, textWidth int) {
	_, screenHeight := screen.Size()
	drawnLines, selectorIndex := 0, -1
	for i := 0; i < len(v.lines) && drawnLines < screenHeight-1; i++ {
		_, isLink := v.lines[i].(parser.LinkLine)
		if isLink {
			selectorIndex++
		}
		if i < v.line {
			continue
		}
		if isLink {
			selector := selector.FromIndex(selectorIndex)
			selector = strings.Repeat(" ", selectorColWidth-len(selector)-1) + selector
			emitStr(screen, offset, drawnLines, styleText, selector)
		}
		drawnLines = v.drawLine(screen, i, drawnLines, offset+selectorColWidth, textWidth)
	}
}

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}
}

// drawLine draws the given line, wrapping it if necessary and returns
// the amount of lines written to screen.
func (v View) drawLine(screen tcell.Screen, lineIndex, drawnLines, offset, textWidth int) int {
	line := v.lines[lineIndex]
	style := styleFor(line)
	var highlights [][]int
	if v.Searchpattern != nil {
		highlights = v.Searchpattern.FindAllStringIndex(line.Text(), -1)
	}
	if wrappable, isWrappable := line.(parser.WrappableLine); isWrappable {
		wrappedLines := linesOfWrappable(wrappable, textWidth)
		for j, wrappedLine := range wrappedLines {
			if lineIndex != v.line || j >= v.lineOffset {
				// FIXME: Could stop drawing once screenHeight-1 is reached.
				emitStrWithHighlights(screen, offset, drawnLines, style, wrappedLine, highlights)
				drawnLines++
			}
			if j == 0 {
				offset += wrappable.IndentWidth()
			}
			highlights = subFromHighlights(highlights, len(wrappedLine))
		}
	} else {
		emitStrWithHighlights(screen, offset, drawnLines, stylePrefromatted, line.Text(), highlights)
		drawnLines++
	}
	return drawnLines
}

func styleFor(line parser.Line) tcell.Style {
	switch line.(type) {
	case parser.TextLine:
		return styleText
	case parser.LinkLine:
		return styleLink
	case parser.PreformattedLine:
		return stylePrefromatted
	case parser.Heading1Line:
		return styleHeading1
	case parser.Heading2Line:
		return styleHeading2
	case parser.Heading3Line:
		return styleHeading3
	case parser.ListLine:
		return styleList
	case parser.QuoteLine:
		return styleQuote
	}
	panic("unknown line type")
}

func linesOfWrappable(wrappable parser.WrappableLine, width int) []string {
	wrapIndexes := wrappable.WrapIndexes(width)
	lines := make([]string, len(wrapIndexes)+1)
	if len(wrapIndexes) == 0 {
		lines[0] = wrappable.Text()
	} else {
		previousWrapIndex := 0
		for i, wrapIndex := range wrapIndexes {
			lines[i] = wrappable.Text()[previousWrapIndex:wrapIndex]
			previousWrapIndex = wrapIndex
		}
		lines[len(lines)-1] = wrappable.Text()[previousWrapIndex:]
	}
	return lines
}

func subFromHighlights(highlights [][]int, x int) [][]int {
	newHighlights := make([][]int, len(highlights), len(highlights))
	for i, h := range highlights {
		newHighlights[i] = []int{h[0] - x, h[1] - x}
	}
	return newHighlights
}

func emitStrWithHighlights(s tcell.Screen, x, y int, style tcell.Style, str string, highlights [][]int) {
	for i, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style.Reverse(withinHighlight(i, highlights)))
		x += w
	}
}

func withinHighlight(i int, highlights [][]int) bool {
	for _, h := range highlights {
		if i >= h[0] && i < h[1] {
			return true
		}
	}
	return false
}
