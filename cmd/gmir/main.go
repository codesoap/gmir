package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"

	"github.com/codesoap/gmir"
	"github.com/codesoap/gmir/readline"
	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
)

var (
	uFlag bool
	tFlag string
)

func showUsageInfo() {
	fmt.Fprintln(flag.CommandLine.Output(), `Usage:
gmir [-u] [-t TITLE] [FILE]
If FILE is not given, standard input is read.

Options:
-u  Hide URLs of links by default
-t  Set a title that is displayed in the bar.

Key bindings:
Up, k     : Scroll up one line
Down, j   : Scroll down one line
Right, l  : Scroll right one column; reset with Esc
u         : Scroll up half a page
d         : Scroll down half a page
Page up, b: Scroll up a full page
Page down,
f, Space  : Scroll down a full page
g         : Go to the top
G         : Go to the bottom
h         : Go to next heading
H         : Go to previous heading
t         : Show table of contents
/         : Start search
?         : Start reverse search
n         : Go to next search match
p         : Go to previous search match
0-9       : Select link or table of contents entry
Esc       : Clear input and right scroll or exit table of contents
v         : Hide link URLs
V         : Show link URLs
q         : Quit`)
}

type views struct {
	doc     gmir.View // The main view.
	toc     gmir.View // The view with the table of contents.
	showTOC bool
}

func (vs *views) activeView() *gmir.View {
	if vs.showTOC {
		return &vs.toc
	}
	return &vs.doc
}

func init() {
	flag.Usage = showUsageInfo
	flag.BoolVar(&uFlag, "u", false, "Hide URLs on link lines by default")
	flag.StringVar(&tFlag, "t", "", "Set a title that is displayed in the bar")
	flag.Parse()
}

func main() {
	in := getInput()
	defer in.Close()
	doc, err := gmir.NewView(in, tFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not parse input:", err)
		os.Exit(1)
	}
	if uFlag {
		doc.HideURLs()
	}

	encoding.Register()
	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	if e := s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}
	doc.Draw(s)
	vs := views{
		doc: doc,
		toc: doc.TOCView(),
	}
	for {
		processEvent(s.PollEvent(), &vs, s)
		s.Clear()
		vs.activeView().Draw(s)
	}
}

func processEvent(event tcell.Event, vs *views, s tcell.Screen) {
	v := vs.activeView()
	switch ev := event.(type) {
	case *tcell.EventResize:
		s.Sync()
		vs.doc.FixLineOffset(s)
		vs.toc.FixLineOffset(s)
	case *tcell.EventKey:
		switch v.Mode {
		case gmir.Regular:
			processKeyEvent(ev, vs, s)
		case gmir.Search, gmir.ReverseSearch:
			switch readline.ProcessKey(ev) {
			case readline.Reading:
				v.Searchterm = readline.Input()
				v.Cursor = readline.Cursor()
			case readline.Done:
				v.Searchterm = ""
				v.Cursor = 0
				history, historyIndex := readline.History()
				re, err := regexp.Compile(history[historyIndex])
				if err != nil {
					v.Info = "Invalid pattern"
				} else {
					v.Searchpattern = re
					if v.Mode == gmir.ReverseSearch {
						if !v.ScrollUpToSearchMatch(s) {
							v.Info = "Pattern not found."
						}
					} else {
						if !v.ScrollDownToSearchMatch(s) {
							v.Info = "Pattern not found."
						}
					}
				}
				v.Mode = gmir.Regular
			case readline.Aborted:
				v.Searchterm = ""
				v.Searchpattern = nil
				v.Cursor = 0
				v.Mode = gmir.Regular
			}
		}
	}
}

func processKeyEvent(ev *tcell.EventKey, vs *views, s tcell.Screen) {
	v := vs.activeView()
	v.Info = ""
	switch ev.Key() {
	case tcell.KeyUp:
		v.Scroll(s, 1)
	case tcell.KeyDown:
		v.Scroll(s, -1)
	case tcell.KeyRight:
		v.ColOffset += 1
	case tcell.KeyPgUp:
		_, height := s.Size()
		v.Scroll(s, height-1)
	case tcell.KeyPgDn:
		_, height := s.Size()
		v.Scroll(s, -height+1)
	case tcell.KeyEsc:
		v.ColOffset = 0
		v.ClearSelector()
		vs.showTOC = false
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q':
			s.Fini()
			os.Exit(0)
		case 'k':
			v.Scroll(s, 1)
		case 'j':
			v.Scroll(s, -1)
		case 'l':
			v.ColOffset += 1
		case 'u':
			_, height := s.Size()
			v.Scroll(s, height/2)
		case 'd':
			_, height := s.Size()
			v.Scroll(s, -height/2)
		case 'b':
			_, height := s.Size()
			v.Scroll(s, height-1)
		case 'f', ' ':
			_, height := s.Size()
			v.Scroll(s, -height+1)
		case 'g':
			v.ScrollToTop(s)
		case 'G':
			v.ScrollToBottom(s)
		case 'h':
			v.ScrollToNextHeading(s)
		case 'H':
			v.ScrollToPrevHeading(s)
		case 't':
			if vs.toc.IsEmpty() {
				v.Info = "Table of contents is empty"
			} else {
				vs.showTOC = true
			}
		case 'p':
			if !v.ScrollUpToNextSearchMatch(s) {
				v.Info = "No previous match found."
			}
		case 'n':
			if !v.ScrollDownToNextSearchMatch(s) {
				v.Info = "No further match found."
			}
		case '/':
			v.Mode = gmir.Search
			v.ClearSelector()
		case '?':
			v.Mode = gmir.ReverseSearch
			v.ClearSelector()
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			digit, _ := strconv.Atoi(string(ev.Rune()))
			v.AddDigitToSelector(digit)
			if v.SelectorIsValid() {
				if vs.showTOC {
					vs.showTOC = false
					vs.doc.ScrollToNthHeading(s, v.SelectorIndex())
					v.ClearSelector()
				} else {
					s.Fini()
					fmt.Println(v.LinkURL())
					os.Exit(0)
				}
			}
		case 'v':
			v.HideURLs()
			v.FixLineOffset(s)
		case 'V':
			v.ShowURLs()
			v.FixLineOffset(s)
		}
	}
}

func getInput() io.ReadCloser {
	if len(flag.Args()) > 1 {
		fmt.Fprintln(os.Stderr, "Too many arguments.")
		os.Exit(1)
	} else if len(flag.Args()) == 1 {
		file, err := os.Open(flag.Args()[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not open given file:", err)
			os.Exit(1)
		}
		return file
	}
	return os.Stdin
}
