package main

import (
	"strings"
)

func CleanText(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}
