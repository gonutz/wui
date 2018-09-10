//+build windows

package wui

import "github.com/gonutz/w32"

func NewEditLine() *EditLine {
	return &EditLine{}
}

type EditLine struct {
	textControl
}

func (e *EditLine) create(id int) {
	e.textControl.create(id, w32.WS_EX_CLIENTEDGE, "EDIT", w32.WS_TABSTOP)
}
