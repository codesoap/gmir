package gmir

import (
	"fmt"
	"io"
	"strings"

	"github.com/codesoap/gmir/linknumber"
	"github.com/codesoap/gmir/parser"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// TODO: right scroll
// TODO: Better name than lineOffset/LineOffset.
// TODO: Search term, forward-/backward-link.

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

// A View represents the whole state related to a document, including
// it's content and scroll position.
type View struct {
	lines []parser.Line
	line  int // Index in lines of the first displayed line.

	// Number of wrapped lines to skip within the first displayed line.
	lineOffset int

	// Linknumber while it is being typed.
	linknumber string
}

func NewView(in io.Reader) (View, error) {
	lines, err := parser.Parse(in)
	if err != nil {
		return View{}, err
	}
	return View{lines, 0, 0, ""}, nil
}

func (v View) links() []parser.LinkLine {
	links := make([]parser.LinkLine, 0)
	for _, line := range v.lines {
		if link, isLink := line.(parser.LinkLine); isLink {
			links = append(links, link)
		}
	}
	return links
}

/*
Draw draws the given view to screen. The screen is separted like this:

	   7cols   9cols     10cols     16cols
	┌───────┬─────────┬──────────┬────────────────┐
	│ left  │ link    │ rendered │ right space    │
	│ space │ numbers │ gmi      │                │
	│       │         │          │                │
	├───────┴─────────┴──────────┴────────────────┤
	│ bar                                         │
	└─────────────────────────────────────────────┘

The rendered gmi aims to be a width that is comfortable to read. The
right space may be intruded by preformatted lines. The link numbers
column will always be wide enough to fit the largest link number in the
document.
*/
func (v View) Draw(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()
	leftSpace, linkColWidth, textWidth := v.columnWidths(screenWidth)
	if screenWidth < linkColWidth+8 || screenHeight < 2 {
		// Screen too small.
		return
	}
	v.drawLinkAndGMIColumn(screen, leftSpace, linkColWidth, textWidth)
	v.drawBar(screen)
	screen.Show()
}

// FixLineOffset fixes v.lineOffset to ensure it does not go over the
// amount of actually available wrapped lines. Use after screen resize.
func (v *View) FixLineOffset(screen tcell.Screen) {
	if v.lineOffset == 0 {
		return
	}
	line := v.lines[v.line]
	if wrappable, isWrappable := line.(parser.WrappableLine); isWrappable {
		screenWidth, _ := screen.Size()
		_, _, textWidth := v.columnWidths(screenWidth)
		maxLineOffset := len(wrappable.WrapIndexes(textWidth))
		if v.lineOffset > maxLineOffset {
			v.lineOffset = maxLineOffset
		}
	} else {
		panic("lineOffset for non wrappable line found.")
	}
}

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

// maxLineOffset returns the maximum legal line offset of line. Returns
// 0 if line is not wrappable.
func (v View) maxLineOffset(screen tcell.Screen, line int) int {
	if wrappable, isWrappable := v.lines[line].(parser.WrappableLine); isWrappable {
		screenWidth, _ := screen.Size()
		_, _, textWidth := v.columnWidths(screenWidth)
		return len(wrappable.WrapIndexes(textWidth))
	}
	return 0
}

func (v View) columnWidths(screenWidth int) (leftSpace, linkColWidth, textWidth int) {
	links := v.links()
	if len(links) > 0 {
		linkColWidth = len(linknumber.FromIndex(len(links)-1)) + 1
	}
	if screenWidth >= maxTextWidth+linkColWidth {
		textWidth = maxTextWidth
		space := screenWidth - (maxTextWidth + linkColWidth)
		if space > linkColWidth+2 {
			leftSpace = (space / 2) - linkColWidth
		}
	} else {
		textWidth = screenWidth - linkColWidth
	}
	return leftSpace, linkColWidth, textWidth
}

func (v View) drawLinkAndGMIColumn(screen tcell.Screen, offset, linkColWidth, textWidth int) {
	_, screenHeight := screen.Size()
	drawnLines, linkIndex := 0, -1
	for i := 0; i < len(v.lines) && drawnLines < screenHeight-1; i++ {
		_, isLink := v.lines[i].(parser.LinkLine)
		if isLink {
			linkIndex++
		}
		if i < v.line {
			continue
		}
		if isLink {
			linkSelector := linknumber.FromIndex(linkIndex)
			linkSelector = strings.Repeat(" ", linkColWidth-len(linkSelector)-1) + linkSelector
			emitStr(screen, offset, drawnLines, styleText, linkSelector)
		}
		drawnLines = v.drawLine(screen, i, drawnLines, offset+linkColWidth, textWidth)
	}
}

// drawLine draws the given line, wrapping it if necessary and returns
// the amount of lines written to screen.
func (v View) drawLine(screen tcell.Screen, lineIndex, drawnLines, offset, textWidth int) int {
	line := v.lines[lineIndex]
	style := styleFor(line)
	if wrappable, isWrappable := line.(parser.WrappableLine); isWrappable {
		wrappedLines := linesOfWrappable(wrappable, textWidth)
		for j, wrappedLine := range wrappedLines {
			if lineIndex == v.line && j < v.lineOffset {
				continue
			}
			// FIXME: Could stop drawing once screenHeight-1 is reached.
			if j == 0 {
				emitStr(screen, offset, drawnLines, style, wrappedLine)
			} else {
				emitStr(screen, offset+wrappable.IndentWidth(), drawnLines, style, wrappedLine)
			}
			drawnLines++
		}
	} else {
		emitStr(screen, offset, drawnLines, stylePrefromatted, line.Text())
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

func (v View) drawBar(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()

	// FIXME: Would be better to calculate percentage of last visible line on screen.
	percent := fmt.Sprintf("%.0f%%", 100*float32(v.line+1)/float32(len(v.lines)))

	text := ""
	if len(v.linknumber)+len(percent) >= screenWidth {
		text = v.linknumber + " " + percent
	} else {
		text = v.linknumber +
			strings.Repeat(" ", screenWidth-len(percent)-len(v.linknumber)) +
			percent
	}
	emitStr(screen, 0, screenHeight-1, styleBar, text)
}
