package gmir

import (
	"fmt"
	"io"
	"regexp"

	"github.com/codesoap/gmir/linknumber"
	"github.com/codesoap/gmir/parser"
	"github.com/gdamore/tcell/v2"
)

// TODO: Select from search history.
// TODO: Better name than lineOffset/LineOffset.
// TODO: forward-/backward-link.
// FIXME: Search term and scroll position are kept in two places. Make it one.

type Mode int

const (
	Regular       = Mode(iota)
	Search        // Typing a search term.
	ReverseSearch // Typing a search term for reverse search.
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

	// Linknumber while it is being typed.
	linknumber string

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
		lines: lines,
		Mode:  Regular,
	}, nil
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
