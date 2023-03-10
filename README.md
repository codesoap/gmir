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
- Jumping between headings is possible.
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
Up, k       : Scroll up one line
Down, j     : Scroll down one line
Page up, u  : Scroll up half a page
Page down, d: Scroll down half a page
g           : Go to the top
G           : Go to the bottom
h           : Go to next heading
H           : Go to previous heading
/           : Start search
?           : Start reverse search
n           : Go to next search match
p           : Go to previous search match
0-9         : Enter link number
Esc         : Clear input
q        : Quit
```

# TODO
Here are some ideas on what could be added in the future, in no
particular order:
- Enable passing forwards and backwards links as command line arguments.
