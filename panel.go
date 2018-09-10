//+build windows

package wui

import "github.com/gonutz/w32"

type Panel struct {
	handle   w32.HWND
	parent   *Window
	x        int
	y        int
	width    int
	height   int
	hidden   bool
	disabled bool
}

func (*Panel) isContainer() {}
