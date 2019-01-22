//+build windows

package wui

import (
	"syscall"

	"github.com/gonutz/w32"
)

func NewTextEdit() *TextEdit {
	return &TextEdit{autoHScroll: true}
}

type TextEdit struct {
	textControl
	limit       int
	autoHScroll bool
}

func (e *TextEdit) create(id int) {
	var hScroll uint
	if e.autoHScroll {
		hScroll |= w32.ES_AUTOHSCROLL
	}
	e.textControl.create(
		id, w32.WS_EX_CLIENTEDGE, "EDIT",
		w32.WS_TABSTOP|w32.WS_VSCROLL|
			w32.ES_LEFT|w32.ES_MULTILINE|w32.ES_AUTOVSCROLL|hScroll|
			w32.ES_WANTRETURN,
	)
	if e.limit != 0 {
		e.SetCharacterLimit(e.limit)
	}
	w32.SetWindowSubclass(e.handle, syscall.NewCallback(func(
		window w32.HWND,
		msg uint32,
		wParam, lParam uintptr,
		subclassID uintptr,
		refData uintptr,
	) uintptr {
		switch msg {
		case w32.WM_CHAR:
			if wParam == 1 {
				var all uintptr
				all--
				if w32.GetKeyState(w32.VK_CONTROL)&0x8000 != 0 {
					w32.SendMessage(e.handle, w32.EM_SETSEL, 0, all)
					return 0
				}
			}
			return w32.DefSubclassProc(window, msg, wParam, lParam)
		default:
			return w32.DefSubclassProc(window, msg, wParam, lParam)
		}
	}), 0, 0)
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

func (e *TextEdit) SetWordWrap(wrap bool) {
	if e.handle == 0 {
		// the ES_AUTOHSCROLL style cannot be changed at runtime
		e.autoHScroll = !wrap
	}
}

func (e *TextEdit) WordWrap() bool {
	return !e.autoHScroll
}
