//+build windows

package wui

import "github.com/gonutz/w32"

type Label struct {
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
	align    uint
}

func NewLabel() *Label {
	return &Label{
		align: w32.SS_LEFT,
	}
}

func (l *Label) setAlign(align uint) *Label {
	l.align = align
	if l.handle != 0 {
		style := uint(w32.GetWindowLongPtr(l.handle, w32.GWL_STYLE))
		style = style &^ w32.SS_LEFT &^ w32.SS_CENTER &^ w32.SS_RIGHT
		w32.SetWindowLongPtr(l.handle, w32.GWL_STYLE, uintptr(style|l.align))
		w32.InvalidateRect(l.handle, nil, true)
	}
	return l
}

func (l *Label) SetLeftAlign() *Label {
	return l.setAlign(w32.SS_LEFT)
}

func (l *Label) SetCenterAlign() *Label {
	return l.setAlign(w32.SS_CENTER)
}

func (l *Label) SetRightAlign() *Label {
	return l.setAlign(w32.SS_RIGHT)
}
