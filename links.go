package gmir

import (
	"fmt"

	"github.com/codesoap/gmir/linknumber"
)

// AddDigitToLinknumber adds a single digit to the end of v.linknumber.
func (v *View) AddDigitToLinknumber(digit int) {
	if digit < 0 || digit > 9 {
		panic(fmt.Sprintf("'%d' is not a digit", digit))
	}
	v.linknumber += fmt.Sprint(digit)
}

// ClearLinknumber sets v.linknumber to an empty string.
func (v *View) ClearLinknumber() {
	v.linknumber = ""
}

// LinkURL returns the URL for v.linknumber. If the number is not
// complete or does not resolve to an existing link index, ok will be
// false and url will be empty.
func (v View) LinkURL() (url string, ok bool) {
	if !linknumber.IsComplete(v.linknumber) {
		return "", false
	}
	links := v.links()
	i := linknumber.ToIndex(v.linknumber)
	if i >= len(links) {
		return "", false
	}
	return links[i].URL(), true
}
