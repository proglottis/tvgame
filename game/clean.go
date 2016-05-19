package game

import (
	"strings"
)

func CleanText(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}
