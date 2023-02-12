package parser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"golang.org/x/text/unicode/norm"
)

var (
	reLinkLine                = regexp.MustCompile(`^=>\s+(\S+)\s+(.+)\s*$`)
	rePreformattingToggleLine = regexp.MustCompile("^```")
	reHeading1Line            = regexp.MustCompile(`^#\s*(.+)\s*$`)
	reHeading2Line            = regexp.MustCompile(`^##\s*(.+)\s*$`)
	reHeading3Line            = regexp.MustCompile(`^###\s*(.+)\s*$`)
	reListLine                = regexp.MustCompile(`^\*\s+(.+)\s*$`)
	reQuoteLine               = regexp.MustCompile(`^>\s*(.+)\s*$`)
)

type Line interface {
	Text() string
}

type WrappableLine interface {
	Line
	WrapIndexes(width int) []int // Byte-indexes before which the line shall be wrapped.
	IndentWidth() int            // Number of blanks to leave at the start of a wrapped line.
}

type TextLine struct{ text string }
type LinkLine struct{ url, name string }
type PreformattedLine struct{ text string }
type Heading1Line struct{ text string }
type Heading2Line struct{ text string }
type Heading3Line struct{ text string }
type ListLine struct{ text string }
type QuoteLine struct{ text string }

func (t TextLine) Text() string         { return t.text }
func (l LinkLine) Text() string         { return fmt.Sprintf("=> %s (%s)", l.name, l.url) }
func (p PreformattedLine) Text() string { return p.text }
func (h Heading1Line) Text() string     { return "# " + h.text }
func (h Heading2Line) Text() string     { return "## " + h.text }
func (h Heading3Line) Text() string     { return "### " + h.text }
func (l ListLine) Text() string         { return "* " + l.text }
func (q QuoteLine) Text() string        { return "> " + q.text }

func (t TextLine) WrapIndexes(width int) []int { return wrapIndexes(t.Text(), width, t.IndentWidth()) }
func (l LinkLine) WrapIndexes(width int) []int { return wrapIndexes(l.Text(), width, l.IndentWidth()) }
func (h Heading1Line) WrapIndexes(width int) []int {
	return wrapIndexes(h.Text(), width, h.IndentWidth())
}
func (h Heading2Line) WrapIndexes(width int) []int {
	return wrapIndexes(h.Text(), width, h.IndentWidth())
}
func (h Heading3Line) WrapIndexes(width int) []int {
	return wrapIndexes(h.Text(), width, h.IndentWidth())
}
func (l ListLine) WrapIndexes(width int) []int  { return wrapIndexes(l.Text(), width, l.IndentWidth()) }
func (q QuoteLine) WrapIndexes(width int) []int { return wrapIndexes(q.Text(), width, q.IndentWidth()) }

func (t TextLine) IndentWidth() int     { return 0 }
func (l LinkLine) IndentWidth() int     { return 3 }
func (h Heading1Line) IndentWidth() int { return 2 }
func (h Heading2Line) IndentWidth() int { return 3 }
func (h Heading3Line) IndentWidth() int { return 4 }
func (l ListLine) IndentWidth() int     { return 2 }
func (q QuoteLine) IndentWidth() int    { return 2 }

func (l LinkLine) URL() string { return l.url }

func wrapIndexes(text string, width, indent int) []int {
	if indent < 0 {
		panic("Bug: Tried to wrap with negative indent.")
	} else if width <= indent {
		return nil
	}
	trimmedText := strings.TrimRight(text, " ")
	remainingWidth := runewidth.StringWidth(trimmedText)
	if remainingWidth <= int(width) {
		return nil
	}
	wrapIndexes := make([]int, 0)

	// Assuming that the prefix (e.g. '=> ' for LinkLine) should never be wrapped.
	i := indent
	remainingWidth -= indent // We know that all indent bytes are runes of width 1.
	for remainingWidth > width-indent {
		if isSingleRune(trimmedText[i:]) {
			// There is a single rune on the last line, which is wider than width-indent.
			break
		}
		oldI := i
		i += indexOfNextWrapChar(text[i:], width-indent) + 1
		remainingWidth -= runewidth.StringWidth(trimmedText[oldI:i])
		wrapIndexes = append(wrapIndexes, i)
	}
	return wrapIndexes
}

func isSingleRune(text string) bool {
	for i := range text {
		if i > 0 {
			return false
		}
	}
	return true
}

// indexOfNextWrapChar takes a text which MUST NOT fit within width and
// returns the index of the byte after which the next line wrap should
// occur.
func indexOfNextWrapChar(text string, width int) int {
	startIndex := lastSpaceWithinWidth(text, width+1)
	if startIndex == -1 {
		startIndex = lastIndexWithinWidth(text, width)
	}
	index := startIndex
	for i, r := range text[startIndex+1:] {
		if runewidth.RuneWidth(r) != 0 && r != ' ' {
			break
		}
		index = startIndex + i + utf8.RuneLen(r)
	}
	return index
}

func lastSpaceWithinWidth(text string, width int) int {
	lastSpaceIndex := -1
	seenWidth := 0
	for i, r := range text {
		rw := runewidth.RuneWidth(r)
		seenWidth += rw
		if seenWidth > width {
			break
		}
		if r == ' ' {
			lastSpaceIndex = i
		}
	}
	return lastSpaceIndex
}

// lastIndexWithinWidth returns the index of the last byte within text
// that fits within width. However, if not even the first rune fits, the
// index of the last byte of the first rune will be returned.
func lastIndexWithinWidth(text string, width int) int {
	seenWidth := 0
	for i, r := range text {
		seenWidth += runewidth.RuneWidth(r)
		if seenWidth > width {
			if i == 0 {
				return utf8.RuneLen(r) - 1
			}
			return i - 1
		}
	}
	panic("text fits within width")
}

// Parse parses the GMI from the given reader. All text will be
// normalized to the NFC form.
func Parse(in io.Reader) ([]Line, error) {
	preformatted := false
	out := make([]Line, 0)
	nfcIn := norm.NFC.Reader(in)
	s := bufio.NewScanner(nfcIn)
	for s.Scan() {
		// TODO: A replacing io.Reader would probably be more performant
		//       than strings.ReplaceAll().
		// Tabs are replaced, because they don't work well with tcell.
		line := strings.ReplaceAll(s.Text(), "\t", "    ")
		if rePreformattingToggleLine.MatchString(line) {
			preformatted = !preformatted
			continue
		}
		if preformatted {
			out = append(out, PreformattedLine{line})
			continue
		}
		if m := reLinkLine.FindStringSubmatch(line); m != nil {
			out = append(out, LinkLine{m[1], m[2]})
			continue
		}
		if m := reHeading3Line.FindStringSubmatch(line); m != nil {
			out = append(out, Heading3Line{m[1]})
			continue
		}
		if m := reHeading2Line.FindStringSubmatch(line); m != nil {
			out = append(out, Heading2Line{m[1]})
			continue
		}
		if m := reHeading1Line.FindStringSubmatch(line); m != nil {
			out = append(out, Heading1Line{m[1]})
			continue
		}
		if m := reListLine.FindStringSubmatch(line); m != nil {
			out = append(out, ListLine{m[1]})
			continue
		}
		if m := reQuoteLine.FindStringSubmatch(line); m != nil {
			out = append(out, QuoteLine{m[1]})
			continue
		}
		out = append(out, TextLine{strings.TrimSpace(line)})
	}
	return out, s.Err()
}
