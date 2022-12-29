package linknumber

import (
	"fmt"
	"strconv"
	"strings"
)

func IsComplete(number string) bool {
	leadingZeroes := 0
	for _, c := range number {
		if c != '0' {
			break
		}
		leadingZeroes++
	}
	completeLinkSelectorLen := leadingZeroes*2 + 1
	// If len(number) > completeLinkSelectorLen, the link can never become
	// complete. This is accepted for now.
	return len(number) == completeLinkSelectorLen
}

func FromIndex(i int) string {
	digitCnt := len(strconv.Itoa(i + 1))
	return strings.Repeat("0", digitCnt-1) + strconv.Itoa(i+1)
}

func ToIndex(number string) int {
	n, err := strconv.Atoi(number)
	if err != nil || n < 1 {
		panic(fmt.Sprint("invalid link number '%s': ", number, err))
	}
	return n - 1
}
