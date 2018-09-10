//+build windows

package wui

func NewEditLine() *EditLine {
	return &EditLine{}
}

type EditLine struct {
	textControl
}
