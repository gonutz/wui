package wui

import (
	"strings"
	"testing"

	"github.com/gonutz/check"
)

func extractCursor(s string) (string, int) {
	cursor := strings.Index(s, "|")
	s = strings.Replace(s, "|", "", 1)
	return s, cursor
}

func TestExtractCursor(t *testing.T) {
	var s string
	var cursor int

	s, cursor = extractCursor("|")
	check.Eq(t, s, "")
	check.Eq(t, cursor, 0)

	s, cursor = extractCursor("abc|")
	check.Eq(t, s, "abc")
	check.Eq(t, cursor, 3)

	s, cursor = extractCursor("abc|def")
	check.Eq(t, s, "abcdef")
	check.Eq(t, cursor, 3)
}

func TestDeleteLastWordAtCursor(t *testing.T) {
	del := func(have, want string) {
		have, haveCursor := extractCursor(have)
		want, wantCursor := extractCursor(want)

		got, gotCursor := deleteWordBeforeCursor([]rune(have), haveCursor)

		check.Eq(t, got, want)
		check.Eq(t, gotCursor, wantCursor)
	}

	del("|", "|")
	del("a|", "|")
	del("ab|", "|")
	del("abc|", "|")
	del("abc |", "|")
	del("abc  |", "|")

	del(" |", "|")
	del("  |", "|")

	del(
		"abc def|",
		"abc |",
	)
	del(
		"abc de|f",
		"abc |f",
	)
	del(
		"abc d|ef",
		"abc |ef",
	)
	del(
		"abc |def",
		"|def",
	)
	del(
		"abc| def",
		"| def",
	)
	del(
		"ab|c def",
		"|c def",
	)
	del(
		"a|bc def",
		"|bc def",
	)
	del(
		"|abc def",
		"|abc def",
	)

	del("a b \t |", "a |")

	del("\r\n|", "|")
	del("\r\n\r\n|", "\r\n|")
}
