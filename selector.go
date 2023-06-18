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

// SelectorNumber returns the currently selected index.
func (v View) SelectorIndex() int {
	return selector.ToIndex(v.selector)
}

// SelectorIsValid returns true, if the selector is complete and
// resolves to a valid selectable.
func (v View) SelectorIsValid() bool {
	return selector.IsComplete(v.selector) && v.selectorIsInRange()
}

func (v View) selectorIsInRange() bool {
	i := selector.ToIndex(v.selector)
	switch v.selectable {
	case link:
		return i < len(v.links())
	case heading:
		return i < len(v.headings())
	}
	panic("unknown selectable")
}

// LinkURL returns the URL for v.selector.
func (v View) LinkURL() string {
	i := selector.ToIndex(v.selector)
	return v.links()[i].URL()
}
