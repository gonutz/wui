//+build windows

package wui

import "github.com/gonutz/w32"

type Button struct {
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
	onClick  func()
}

func NewButton() *Button {
	return &Button{}
}

func (b *Button) SetOnClick(f func()) *Button {
	b.onClick = f
	return b
}

func (b *Button) create(parent *Window, id int, instance w32.HINSTANCE) {
	if b.handle == 0 {
		b.handle = w32.CreateWindowExStr(
			0,
			"BUTTON",
			b.text,
			w32.WS_VISIBLE|w32.WS_CHILD|w32.WS_TABSTOP|w32.BS_DEFPUSHBUTTON,
			b.x, b.y, b.width, b.height,
			parent.handle, w32.HMENU(id), instance, nil,
		)
	}
	b.afterCreate(parent)
}
