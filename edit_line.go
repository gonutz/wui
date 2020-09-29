//+build windows

package wui

import "github.com/gonutz/w32"

func NewEditLine() *EditLine {
	return &EditLine{}
}

type EditLine struct {
	textControl
	isPass       bool
	passChar     uintptr
	limit        int
	onTextChange func()
}

func (e *EditLine) create(id int) {
	e.textControl.create(
		id,
		w32.WS_EX_CLIENTEDGE,
		"EDIT",
		w32.WS_TABSTOP|w32.ES_AUTOHSCROLL|w32.ES_PASSWORD,
	)
	e.passChar = w32.SendMessage(e.handle, w32.EM_GETPASSWORDCHAR, 0, 0)
	e.SetPassword(e.isPass)
	if e.limit != 0 {
		e.SetCharacterLimit(e.limit)
	}
}

func (e *EditLine) SetPassword(isPass bool) {
	e.isPass = isPass
	if e.handle != 0 {
		if e.isPass {
			w32.SendMessage(e.handle, w32.EM_SETPASSWORDCHAR, e.passChar, 0)
		} else {
			w32.SendMessage(e.handle, w32.EM_SETPASSWORDCHAR, 0, 0)
		}
		w32.InvalidateRect(e.parent.getHandle(), nil, true)
	}
}

func (e *EditLine) IsPassword() bool {
	return e.isPass
}

func (e *EditLine) SetCharacterLimit(count int) {
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
