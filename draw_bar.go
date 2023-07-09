package gmir

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

func (v View) drawBar(screen tcell.Screen) {
	screenWidth, screenHeight := screen.Size()

	// Prefill bar with right style:
	emitStr(screen, 0, screenHeight-1, styleBar, strings.Repeat(" ", screenWidth))

	// FIXME: Would be better to calculate percentage of last visible line on screen.
	percent := fmt.Sprintf("%.0f%%", 100*float32(v.line+1)/float32(len(v.lines)))
	leftWidth := screenWidth - len(percent)
	emitStr(screen, leftWidth, screenHeight-1, styleBar, percent)

	if v.Info != "" {
		emitStr(screen, 0, screenHeight-1, styleBar, v.Info+" ")
	} else if v.Mode == Search || v.Mode == ReverseSearch {
		v.drawSearchText(screen, leftWidth-1)
	} else if v.selector == "" {
		emitStr(screen, 0, screenHeight-1, styleBar, v.title+" ")
	} else {
		emitStr(screen, 0, screenHeight-1, styleBar, v.selector+" ")
	}
}

func (v View) drawSearchText(screen tcell.Screen, maxWidth int) {
	if maxWidth < 5 {
		return
	}
	searchWidth := runewidth.StringWidth(v.Searchterm)
	text := searchPrefix(v.Mode)
	prefixWidth := runewidth.StringWidth(text)
	maxWidth -= prefixWidth
	cursor := len(text) // Byte index of cursor within text.
	if searchWidth < maxWidth || v.Cursor < len(v.Searchterm) && searchWidth == maxWidth {
		// Text fits within maxWidth.
		text += v.Searchterm
		cursor += v.Cursor
	} else if v.Cursor == 0 {
		// Start at cursor.
		maxWidth -= 1
		endIndex := headOfText(v.Searchterm[v.Cursor:], maxWidth)
		text += v.Searchterm[v.Cursor:v.Cursor+endIndex] + "…"
	} else {
		text += "…"
		cursor = len(text)
		maxWidth -= 1
		if v.Cursor == len(v.Searchterm) {
			// Cursor is behind last character of v.Searchterm.
			maxWidth -= 1
			startIndex := tailOfText(v.Searchterm, searchWidth, maxWidth)
			text += v.Searchterm[startIndex:]
			cursor = len(text)
		} else if runewidth.StringWidth(v.Searchterm[v.Cursor:]) < maxWidth {
			// Start before cursor.
			startIndex := tailOfText(v.Searchterm, searchWidth, maxWidth)
			text += v.Searchterm[startIndex:]
			cursor += v.Cursor - startIndex
		} else {
			// Start at cursor.
			maxWidth -= 1
			endIndex := headOfText(v.Searchterm[v.Cursor:], maxWidth)
			text += v.Searchterm[v.Cursor:v.Cursor+endIndex] + "…"
		}
	}
	_, screenHeight := screen.Size()
	emitStrWithCursor(screen, 0, screenHeight-1, styleBar, text, cursor)
}

func searchPrefix(mode Mode) string {
	if mode == Search {
		return "/"
	} else if mode == ReverseSearch {
		return "?"
	}
	return ""
}

// tailOfText returns the index within text, where the tail, that fits
// within maxWidth, begins.
func tailOfText(text string, textWidth, maxWidth int) (startIndex int) {
	seenWidth := 0
	stopOnNextRealRune := false
	for i, r := range text {
		rw := runewidth.RuneWidth(r)
		seenWidth += rw
		if stopOnNextRealRune && rw > 0 {
			startIndex = i
			break
		}
		if textWidth-seenWidth <= maxWidth {
			stopOnNextRealRune = true
		}
	}
	return startIndex
}

// headOfText returns the index within text, where the head, that fits
// within maxWidth, ends.
func headOfText(text string, maxWidth int) (endIndex int) {
	seenWidth := 0
	for i, r := range text {
		seenWidth += runewidth.RuneWidth(r)
		endIndex = i
		if seenWidth > maxWidth {
			break
		}
	}
	return endIndex
}

func emitStrWithCursor(s tcell.Screen, x, y int, style tcell.Style, str string, cursor int) {
	if cursor == len(str) {
		str += " "
	}
	for i, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style.Reverse(i != cursor))
		x += w
	}
}
