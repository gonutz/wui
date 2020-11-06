//+build windows

package wui

import "github.com/gonutz/w32"

func NewLabel() *Label {
	return &Label{}
}

type Label struct {
	textControl
	alignment TextAlignment
}

var _ Control = (*Label)(nil)

func (*Label) canFocus() bool {
	return false
}

func (*Label) eatsTabs() bool {
	return false
}

type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignCenter
	AlignRight
)

func (a TextAlignment) String() string {
	// NOTE that these strings are used in the designer to get their
	// representations as Go code so they must always correspond to their
	// constant names and be prefixed with the package name.
	switch a {
	case AlignLeft:
		return "wui.AlignLeft"
	case AlignCenter:
		return "wui.AlignCenter"
	case AlignRight:
		return "wui.AlignRight"
	default:
		return "unknown TextAlignment"
	}
}

func (l *Label) create(id int) {
	l.textControl.create(id, 0, "STATIC", w32.SS_CENTERIMAGE|alignStyle(l.alignment))
}

func alignStyle(a TextAlignment) uint {
	if a == AlignCenter {
		return w32.SS_CENTER
	}
	if a == AlignRight {
		return w32.SS_RIGHT
	}
	return w32.SS_LEFT
}

func (l *Label) SetAlignment(a TextAlignment) {
	l.alignment = a
	if l.handle != 0 {
		style := uint(w32.GetWindowLongPtr(l.handle, w32.GWL_STYLE))
		style = style &^ w32.SS_LEFT &^ w32.SS_CENTER &^ w32.SS_RIGHT
		w32.SetWindowLongPtr(
			l.handle,
			w32.GWL_STYLE,
			uintptr(style|alignStyle(l.alignment)),
		)
		w32.InvalidateRect(l.handle, nil, true)
	}
}

func (l *Label) Alignment() TextAlignment {
	return l.alignment
}
