package readline

import (
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

var (
	line string

	// Byte-Index of line on which the cursor is located. Will always
	// be at the beginning of a unicode character or len(line).
	cursor int

	history      []string
	historyIndex int
)

type Status int

const (
	Reading = Status(iota)
	Done
	Aborted
)

func Input() string {
	return line
}

func Cursor() int {
	return cursor
}

func History() ([]string, int) {
	return history, historyIndex
}

// ProcessKey processes a single key input. If the key changed the
// status to Done or Aborted, the current line will be cleared and
// either added to the history or discarded.
func ProcessKey(ev *tcell.EventKey) Status {
	// TODO: Implement history entry selection.
	/*
		TODO: Implement all these shortcuts from https://github.com/peterh/liner:

		Ctrl-A, Home         : Move cursor to beginning of line
		Ctrl-E, End          : Move cursor to end of line
		Ctrl-B, Left         : Move cursor one character left
		Ctrl-F, Right        : Move cursor one character right
		Ctrl-Left, Alt-B     : Move cursor to previous word
		Ctrl-Right, Alt-F    : Move cursor to next word
		Ctrl-D, Del          : (if line is not empty) Delete character under cursor
		Ctrl-C, Esc          : Abort
		Ctrl-H, BackSpace    : Delete character before cursor
		Ctrl-W, Alt-BackSpace: Delete word leading up to cursor
		Alt-D                : Delete word following cursor
		Ctrl-K               : Delete from cursor to end of line
		Ctrl-U               : Delete from start of line to cursor
		Ctrl-P, Up           : Previous match from history
		Ctrl-N, Down         : Next match from history
	*/
	switch ev.Key() {
	case tcell.KeyLeft:
		goLeft()
		return Reading
	case tcell.KeyRight:
		goRight()
		return Reading
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if cursor > 0 {
			goLeft()
			deleteCharUnderCursor()
		}
		return Reading
	case tcell.KeyDelete:
		deleteCharUnderCursor()
		return Reading
	case tcell.KeyEnter:
		if line != "" {
			history = append(history, line)
			historyIndex = len(history) - 1
			line = ""
			cursor = 0
		}
		return Done
	case tcell.KeyEsc:
		line = ""
		cursor = 0
		return Aborted
	case tcell.KeyRune:
		insertRune(ev.Rune())
		return Reading
	}
	return Reading
}

func goLeft() {
	if cursor == 0 {
		return
	}
	var prevRuneIndex int
	for i := range line {
		if i >= cursor {
			break
		}
		prevRuneIndex = i
	}
	cursor = prevRuneIndex
}

func goRight() {
	if cursor == len(line) {
		return
	}
	var nextRuneIndex int
	for i := range line {
		nextRuneIndex = i
		if i > cursor {
			break
		}
	}
	if nextRuneIndex == cursor {
		cursor = len(line)
	} else {
		cursor = nextRuneIndex
	}
}

func insertRune(input rune) {
	if cursor == len(line) {
		line += string(input)
	} else {
		line = line[:cursor] + string(input) + line[cursor:]
	}
	cursor += utf8.RuneLen(input)
}

func deleteCharUnderCursor() {
	if cursor == len(line) {
		return
	}
	endOfDelete := cursor
	for i, r := range line[cursor:] {
		if i == 0 {
			continue
		}
		rw := runewidth.RuneWidth(r)
		if rw > 0 {
			endOfDelete = cursor + i
			break
		}
	}
	if endOfDelete == cursor {
		line = line[:cursor]
	} else {
		line = line[:cursor] + line[endOfDelete:]
	}
}
