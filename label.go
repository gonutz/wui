//+build windows

package wui

import "github.com/gonutz/w32"

func NewLabel() *Label {
	return &Label{align: w32.SS_LEFT}
}

type Label struct {
	textControl
	align uint
}

func (l *Label) create(id int) {
	l.textControl.create(id, 0, "STATIC", w32.SS_CENTERIMAGE|l.align)
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

func (l *Label) IsLeftAligned() bool {
	return l.align == w32.SS_LEFT
}

func (l *Label) IsCenterAligned() bool {
	return l.align == w32.SS_CENTER
}

func (l *Label) IsRightAligned() bool {
	return l.align == w32.SS_RIGHT
}
