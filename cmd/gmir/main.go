package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/codesoap/gmir"
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
0-9      : Enter link number
Esc      : Reset link number
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
		switch ev := s.PollEvent().(type) {
		case *tcell.EventResize:
			s.Sync()
			redraw(v, s)
			v.FixLineOffset(s)
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyUp:
				v.Scroll(s, 1)
				redraw(v, s)
			case tcell.KeyDown:
				v.Scroll(s, -1)
				redraw(v, s)
			case tcell.KeyPgUp:
				_, height := s.Size()
				v.Scroll(s, height/2)
				redraw(v, s)
			case tcell.KeyPgDn:
				_, height := s.Size()
				v.Scroll(s, -height/2)
				redraw(v, s)
			case tcell.KeyEsc:
				v.ClearLinknumber()
				redraw(v, s)
			case tcell.KeyRune:
				switch ev.Rune() {
				case 'q':
					s.Fini()
					os.Exit(0)
				case '1', '2', '3', '4', '5', '6', '7', '8', '9', '0':
					digit, _ := strconv.Atoi(string(ev.Rune()))
					v.AddDigitToLinknumber(digit)
					if url, ok := v.LinkURL(); ok {
						s.Fini()
						fmt.Println(url)
						os.Exit(0)
					}
					redraw(v, s)
				}
			}
		}
	}
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
