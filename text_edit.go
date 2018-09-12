//+build windows

package wui

import "github.com/gonutz/w32"

func NewTextEdit() *TextEdit {
	return &TextEdit{}
}

type TextEdit struct {
	textControl
}

func (e *TextEdit) create(id int) {
	e.textControl.create(
		id, w32.WS_EX_CLIENTEDGE, "EDIT",
		w32.WS_TABSTOP|w32.WS_VSCROLL|
			w32.ES_LEFT|w32.ES_MULTILINE|w32.ES_AUTOVSCROLL|w32.ES_WANTRETURN,
	)
}
