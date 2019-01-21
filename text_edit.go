//+build windows

package wui

import (
	"syscall"

	"github.com/gonutz/w32"
)

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
