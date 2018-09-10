//+build windows

package wui

import "github.com/gonutz/w32"

type EditLine struct {
	handle   w32.HWND
	parent   *Window
	x        int
	y        int
	width    int
	height   int
	hidden   bool
	disabled bool
	text     string
	font     *Font
}

func NewEditLine() *EditLine {
	return &EditLine{}
}
