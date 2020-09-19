//+build windows

package wui

import "github.com/gonutz/w32"

func NewButton() *Button {
	return &Button{}
}

type anchor int

const (
	anchorNone anchor = iota
	anchorRight
	anchorLeftAndRight
	anchorLeftAndCenter
	anchorRightAndCenter
	anchorCenter
	anchorBottom
	anchorTopAndBottom
)

type Button struct {
	textControl
	hAnchor anchor
	vAnchor anchor
	onClick func()
}

func (b *Button) AnchorLeft() {
	b.hAnchor = anchorNone
}

func (b *Button) AnchorRight() {
	b.hAnchor = anchorRight
}

func (b *Button) AnchorLeftAndRight() {
	b.hAnchor = anchorLeftAndRight
}

func (b *Button) AnchorHorizontalCenter() {
	b.hAnchor = anchorCenter
}

func (b *Button) AnchorLeftAndCenter() {
	b.hAnchor = anchorLeftAndCenter
}

func (b *Button) AnchorRightAndCenter() {
	b.hAnchor = anchorRightAndCenter
}

func (b *Button) AnchorTop() {
	b.vAnchor = anchorNone
}

func (b *Button) AnchorBottom() {
	b.vAnchor = anchorBottom
}

func (b *Button) AnchorTopAndBottom() {
	b.vAnchor = anchorTopAndBottom
}

func (b *Button) OnClick() func() {
	return b.onClick
}

func (b *Button) SetOnClick(f func()) {
	b.onClick = f
}

func (b *Button) create(id int) {
	b.textControl.create(id, 0, "BUTTON", w32.WS_TABSTOP|w32.BS_DEFPUSHBUTTON)
}
