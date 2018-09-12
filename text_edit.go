//+build windows

package wui

import "github.com/gonutz/w32"

func NewTextEdit() *TextEdit {
	return &TextEdit{}
}

type TextEdit struct {
	textControl
	limit int
}

func (e *TextEdit) create(id int) {
	e.textControl.create(
		id, w32.WS_EX_CLIENTEDGE, "EDIT",
		w32.WS_TABSTOP|w32.WS_VSCROLL|
			w32.ES_LEFT|w32.ES_MULTILINE|w32.ES_AUTOVSCROLL|w32.ES_AUTOHSCROLL|
			w32.ES_WANTRETURN,
	)
	if e.limit != 0 {
		e.SetCharacterLimit(e.limit)
	}
}

func (e *TextEdit) SetCharacterLimit(count int) {
	e.limit = count
	if e.handle != 0 {
		w32.SendMessage(e.handle, w32.EM_SETLIMITTEXT, uintptr(e.limit), 0)
	}
}

func (e *TextEdit) CharacterLimit() int {
	if e.handle != 0 {
		e.limit = int(w32.SendMessage(e.handle, w32.EM_GETLIMITTEXT, 0, 0))
	}
	return e.limit
}
