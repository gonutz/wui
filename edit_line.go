//+build windows

package wui

import "github.com/gonutz/w32"

func NewEditLine() *EditLine {
	return &EditLine{limit: 0x7FFFFFFE}
}

type EditLine struct {
	textEditControl
	isPassword   bool
	passwordChar uintptr
	limit        int
	readOnly     bool
	onTextChange func()
}

var _ Control = (*EditLine)(nil)

func (*EditLine) canFocus() bool {
	return true
}

func (*EditLine) eatsTabs() bool {
	return false
}

func (e *EditLine) create(id int) {
	e.textEditControl.create(
		id,
		w32.WS_EX_CLIENTEDGE,
		"EDIT",
		w32.WS_TABSTOP|w32.ES_AUTOHSCROLL|w32.ES_PASSWORD,
	)
	e.passwordChar = w32.SendMessage(e.handle, w32.EM_GETPASSWORDCHAR, 0, 0)
	e.SetIsPassword(e.isPassword)
	e.SetCharacterLimit(e.limit)
	e.SetReadOnly(e.readOnly)
}

func (e *EditLine) SetIsPassword(isPassword bool) {
	e.isPassword = isPassword
	if e.handle != 0 {
		if e.isPassword {
			w32.SendMessage(e.handle, w32.EM_SETPASSWORDCHAR, e.passwordChar, 0)
		} else {
			w32.SendMessage(e.handle, w32.EM_SETPASSWORDCHAR, 0, 0)
		}
		w32.InvalidateRect(e.parent.getHandle(), nil, true)
	}
}

func (e *EditLine) IsPassword() bool {
	return e.isPassword
}

func (e *EditLine) SetCharacterLimit(count int) {
	if count <= 0 || count > 0x7FFFFFFE {
		count = 0x7FFFFFFE
	}
	e.limit = count
	if e.handle != 0 {
		w32.SendMessage(e.handle, w32.EM_SETLIMITTEXT, uintptr(e.limit), 0)
	}
}

func (e *EditLine) CharacterLimit() int {
	if e.handle != 0 {
		e.limit = int(w32.SendMessage(e.handle, w32.EM_GETLIMITTEXT, 0, 0))
	}
	return e.limit
}

func (e *EditLine) SetReadOnly(readOnly bool) {
	e.readOnly = readOnly
	if e.handle != 0 {
		if readOnly {
			w32.SendMessage(e.handle, w32.EM_SETREADONLY, 1, 0)
		} else {
			w32.SendMessage(e.handle, w32.EM_SETREADONLY, 0, 0)
		}
	}
}

func (e *EditLine) ReadOnly() bool {
	return e.readOnly
}

func (e *EditLine) SetOnTextChange(f func()) {
	e.onTextChange = f
}

func (e *EditLine) OnTextChange() func() {
	return e.onTextChange
}

func (e *EditLine) handleNotification(cmd uintptr) {
	if cmd == w32.EN_CHANGE && e.onTextChange != nil {
		e.onTextChange()
	}
}
