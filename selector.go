package gmir

import (
	"fmt"

	"github.com/codesoap/gmir/selector"
)

// AddDigitToSelector adds a single digit to the end of v.selector.
func (v *View) AddDigitToSelector(digit int) {
	if digit < 0 || digit > 9 {
		panic(fmt.Sprintf("'%d' is not a digit", digit))
	}
	v.selector += fmt.Sprint(digit)
}

// ClearSelector sets v.selector to an empty string.
func (v *View) ClearSelector() {
	v.selector = ""
}

// LinkURL returns the URL for v.selector. If the selector is not
// complete or does not resolve to an existing link index, ok will be
// false and url will be empty.
func (v View) LinkURL() (url string, ok bool) {
	if !selector.IsComplete(v.selector) {
		return "", false
	}
	links := v.links()
	i := selector.ToIndex(v.selector)
	if i >= len(links) {
		return "", false
	}
	return links[i].URL(), true
}
