//+build windows

package wui

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
