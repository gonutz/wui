package wui

import "github.com/gonutz/w32/v2"

func NewTextEdit() *TextEdit {
	return &TextEdit{
		autoHScroll: true,
		limit:       0x7FFFFFFE,
	}
}

type TextEdit struct {
	textEditControl
	limit        int
	autoHScroll  bool
	writesTabs   bool
	readOnly     bool
	onTextChange func()
}

var _ Control = (*TextEdit)(nil)

func (e *TextEdit) closing() {
	e.Text()
}

func (*TextEdit) canFocus() bool {
	return true
}

func (e *TextEdit) OnTabFocus() func() {
	return e.onTabFocus
}

func (e *TextEdit) SetOnTabFocus(f func()) {
	e.onTabFocus = f
}

func (e *TextEdit) eatsTabs() bool {
	return e.writesTabs
}

func (e *TextEdit) create(id int) {
	var hScroll uint
	if e.autoHScroll {
		hScroll = w32.ES_AUTOHSCROLL | w32.WS_HSCROLL
	}
	e.textEditControl.create(
		id, w32.WS_EX_CLIENTEDGE, "EDIT",
		w32.WS_TABSTOP|w32.WS_VSCROLL|
			w32.ES_LEFT|w32.ES_MULTILINE|w32.ES_AUTOVSCROLL|hScroll|
			w32.ES_WANTRETURN,
	)
	if e.limit != 0 {
		e.SetCharacterLimit(e.limit)
	}
	e.SetReadOnly(true)
}

func (e *TextEdit) SetCharacterLimit(count int) {
	if count <= 0 || count > 0x7FFFFFFE {
		count = 0x7FFFFFFE
	}
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

func (e *TextEdit) SetWordWrap(wrap bool) {
	if e.handle == 0 {
		// the ES_AUTOHSCROLL style cannot be changed at runtime
		e.autoHScroll = !wrap
	}
}

func (e *TextEdit) WordWrap() bool {
	return !e.autoHScroll
}

func (e *TextEdit) WritesTabs() bool {
	return e.writesTabs
}

func (e *TextEdit) SetWritesTabs(tabs bool) {
	e.writesTabs = tabs
}

func (e *TextEdit) SetOnTextChange(f func()) {
	e.onTextChange = f
}

func (e *TextEdit) OnTextChange() func() {
	return e.onTextChange
}

func (e *TextEdit) handleNotification(cmd uintptr) {
	if cmd == w32.EN_CHANGE && e.onTextChange != nil {
		e.onTextChange()
	}
}

func (e *TextEdit) SetReadOnly(readOnly bool) {
	e.readOnly = readOnly
	if e.handle != 0 {
		if readOnly {
			w32.SendMessage(e.handle, w32.EM_SETREADONLY, 1, 0)
		} else {
			w32.SendMessage(e.handle, w32.EM_SETREADONLY, 0, 0)
		}
	}
}

func (e *TextEdit) ReadOnly() bool {
	return e.readOnly
}
