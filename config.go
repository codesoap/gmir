package gmir

import (
	"github.com/gdamore/tcell/v2"
)

var (
	maxTextWidth      = 72
	styleText         = tcell.StyleDefault
	styleLink         = tcell.StyleDefault.Foreground(tcell.ColorBlue)
	stylePrefromatted = tcell.StyleDefault
	styleHeading1     = tcell.StyleDefault.Bold(true)
	styleHeading2     = tcell.StyleDefault.Bold(true)
	styleHeading3     = tcell.StyleDefault.Bold(true)
	styleList         = tcell.StyleDefault
	styleQuote        = tcell.StyleDefault.Italic(true)
	styleBar          = tcell.StyleDefault.Reverse(true)
)
