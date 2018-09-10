//+build windows

package wui

import "github.com/gonutz/w32"

func NewButton() *Button {
	return &Button{}
}

type Button struct {
	textControl
	onClick func()
}

func (b *Button) SetOnClick(f func()) {
	b.onClick = f
}

func (b *Button) create(id int) {
	b.textControl.create(id, 0, "BUTTON", w32.WS_TABSTOP|w32.BS_DEFPUSHBUTTON)
}
