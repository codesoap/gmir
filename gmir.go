package gmir

import (
	"fmt"
	"io"
	"regexp"

	"github.com/codesoap/gmir/parser"
	"github.com/codesoap/gmir/selector"
	"github.com/gdamore/tcell/v2"
)

// TODO: Select from search history.
// TODO: Better name than lineOffset/LineOffset.
// TODO: forward-/backward-link.
// FIXME: Search term and scroll position are kept in two places. Make it one.

type Mode int
type selectable int

const (
	Regular       = Mode(iota)
	Search        // Typing a search term.
	ReverseSearch // Typing a search term for reverse search.
)

const (
	link    = selectable(iota)
	heading // Selecting headings within the table of contents.
)

// A View represents the whole state related to a document, including
// its content and scroll position.
type View struct {
	lines []parser.Line
	line  int // Index in lines of the first displayed line.

	// Number of wrapped lines to skip within the first displayed line.
	lineOffset int

	// Number of columns to shift the content to the left. Useful for
	// viewing preformatted text, that is wider than the screen. The
	// shifting will be limited by the widest preformatted line in lines.
	ColOffset int

	selectable selectable
	selector   string // Selector while it is being typed.

	Mode          Mode
	Searchterm    string         // The search term while it is being typed.
	Cursor        int            // Index of first byte of cursored rune in Searchterm. May be up to len(Searchterm).
	Searchpattern *regexp.Regexp // The active search pattern.

	// If not "", this info is displayed in the bar. Useful for infos like
	// "Invalid search pattern" or "No match found".
	Info string
}

func NewView(in io.Reader) (View, error) {
	lines, err := parser.Parse(in)
	if err != nil {
		return View{}, err
	} else if len(lines) == 0 {
		return View{}, fmt.Errorf("given GMI is empty")
	}
	return View{
		lines:      lines,
		Mode:       Regular,
		selectable: link,
	}, nil
}

// TOCView returns a copy of v only containing headings and with
// selectable set to heading. This copy is suitable for use as a table
// of contents.
func (v View) TOCView() View {
	return View{
		lines:      v.headings(),
		Mode:       Regular,
		selectable: heading,
	}
}

func (v View) IsEmpty() bool {
	return len(v.lines) == 0
}

// ShowURLs enables the display of URLs for link lines.
func (v *View) ShowURLs() {
	parser.ShowURLs = true
}

// HideURLs disables the display of URLs for link lines.
func (v *View) HideURLs() {
	parser.ShowURLs = false
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

func (v View) headings() []parser.Line {
	headings := make([]parser.Line, 0)
	for _, line := range v.lines {
		if isHeading(line) {
			headings = append(headings, line)
		}
	}
	return headings
}

func (v View) isSelectable(line parser.Line) bool {
	switch v.selectable {
	case link:
		_, isLink := line.(parser.LinkLine)
		return isLink
	case heading:
		return isHeading(line)
	}
	panic("unknown selectable")
}

// FixLineOffset fixes v.lineOffset to ensure it does not go over the
// amount of actually available wrapped lines. Use after screen resize
// or any other event that changes the amount of wraps for a line.
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

func (v View) columnWidths(screenWidth int) (leftSpace, selectorColWidth, textWidth int) {
	var selectableCount int
	switch v.selectable {
	case link:
		selectableCount = len(v.links())
	case heading:
		selectableCount = len(v.headings())
	default:
		panic("unknown selectable")
	}
	if selectableCount > 0 {
		selectorColWidth = len(selector.FromIndex(selectableCount-1)) + 1
	}
	if screenWidth >= maxTextWidth+selectorColWidth {
		textWidth = maxTextWidth
		space := screenWidth - (maxTextWidth + selectorColWidth)
		if space > selectorColWidth+2 {
			leftSpace = (space / 2) - selectorColWidth
		}
	} else {
		textWidth = screenWidth - selectorColWidth
	}
	return leftSpace, selectorColWidth, textWidth
}
