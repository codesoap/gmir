`gmir` is a reader for gmi files (of the Gemini protocol).

**`gmir` is still in its early development and behavior might change
again.**

The goal of `gmir` is to make reading gmi files more pleasant than with
a pager like `less`, while also offering link selection to integrate
better with other Gemini software:
- Words are not broken when wrapping lines.
- Preformatted text is never wrapped.
- Indentation is added when wrapping e.g. list lines.
- Syntax like headings and links are highlighted.
- Selecting links is possible. The URL of the selection will be printed
  to stdout.

# Installation
```
go install github.com/codesoap/gmir/cmd/gmir@latest
```

The `gmir` binary is now located at `~/go/bin/gmir`. If you use Go
version 1.15 or older, use `go get` instead.

# Usage
```
$ gmir -h
Usage:
gmir [FILE]
If FILE is not given, standard input is read.

Key bindings:
Up       : Scroll up one line
Down     : Scroll down one line
Page up  : Scroll up half a page
Page down: Scroll down half a page
0-9      : Enter link number
Esc      : Reset link number
q        : Quit
```

# TODO
Here are some ideas on what could be added in the future, in no
particular order:
- Enable searching.
	- I wan't "readline-style" keyboard shortcuts and a history to be available. [peterh/liner](https://github.com/peterh/liner) and [chzyer/readline](https://github.com/chzyer/readline) implement this, but don't seem to be usable with `tcell` (see [chzyer/readline #180](https://github.com/chzyer/readline/issues/180), [tcell #179](https://github.com/gdamore/tcell/issues/179) and [tcell #146](https://github.com/gdamore/tcell/issues/146)). Please let me know if you know other relevant libraries or have a different idea.
- Viewing a table of contents and jumping to selected headings.
- Enable passing forwards and backwards links as command line arguments.
