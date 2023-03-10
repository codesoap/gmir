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

func showUsageInfo() {
	fmt.Fprintln(flag.CommandLine.Output(), `Usage:
gmir [FILE]
If FILE is not given, standard input is read.

Key bindings:
Up       : Scroll up one line
Down     : Scroll down one line
Page up  : Scroll up half a page
Page down: Scroll down half a page
g        : Go to the top
G        : Go to the bottom
h        : Go to next heading
H        : Go to previous heading
/        : Start search
?        : Start reverse search
n        : Go to next search match
p        : Go to previous search match
0-9      : Enter link number
Esc      : Clear input
q        : Quit`)
}

func main() {
	flag.Usage = showUsageInfo
	flag.Parse()

	in := getInput()
	defer in.Close()
	v, err := gmir.NewView(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not parse input:", err)
		os.Exit(1)
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
	v.Draw(s)
	for {
		processEvent(s.PollEvent(), &v, s)
	}
}

func processEvent(event tcell.Event, v *gmir.View, s tcell.Screen) {
	switch ev := event.(type) {
	case *tcell.EventResize:
		s.Sync()
		v.FixLineOffset(s)
		redraw(*v, s)
	case *tcell.EventKey:
		switch v.Mode {
		case gmir.Regular:
			processKeyEvent(ev, v, s)
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
			redraw(*v, s)
		}
	}
}

func processKeyEvent(ev *tcell.EventKey, v *gmir.View, s tcell.Screen) {
	v.Info = ""
	switch ev.Key() {
	case tcell.KeyUp:
		v.Scroll(s, 1)
	case tcell.KeyDown:
		v.Scroll(s, -1)
	case tcell.KeyPgUp:
		_, height := s.Size()
		v.Scroll(s, height/2)
	case tcell.KeyPgDn:
		_, height := s.Size()
		v.Scroll(s, -height/2)
	case tcell.KeyEsc:
		v.ClearLinknumber()
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'q':
			s.Fini()
			os.Exit(0)
		case 'g':
			v.ScrollToTop(s)
		case 'G':
			v.ScrollToBottom(s)
		case 'h':
			v.ScrollToNextHeading(s)
		case 'H':
			v.ScrollToPrevHeading(s)
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
			v.ClearLinknumber()
		case '?':
			v.Mode = gmir.ReverseSearch
			v.ClearLinknumber()
		case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
			digit, _ := strconv.Atoi(string(ev.Rune()))
			v.AddDigitToLinknumber(digit)
			if url, ok := v.LinkURL(); ok {
				s.Fini()
				fmt.Println(url)
				os.Exit(0)
			}
		}
	}
	redraw(*v, s)
}

func redraw(v gmir.View, s tcell.Screen) {
	s.Clear()
	v.Draw(s)
}

func getInput() io.ReadCloser {
	if len(os.Args) > 2 {
		fmt.Fprintln(os.Stderr, "Too many arguments.")
		os.Exit(1)
	} else if len(os.Args) == 2 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not open given file:", err)
			os.Exit(1)
		}
		return file
	}
	return os.Stdin
}
