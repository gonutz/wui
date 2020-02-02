//+build windows

package wui

import (
	"unicode/utf8"
	"unsafe"

	"github.com/gonutz/w32"
)

func NewEditLine() *EditLine {
	return &EditLine{}
}

type EditLine struct {
	textControl
	isPass       bool
	passChar     uintptr
	limit        int
	cursorStart  int
	cursorEnd    int
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
	if e.cursorStart != 0 || e.cursorEnd != 0 {
		e.setCursor(e.cursorStart, e.cursorEnd)
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

// CursorPosition returns the current cursor position, respectively the current
// selection.
//
// If no selection is active, the returned start and end values are the same.
// They then indicate the index of the UTF-8 character in this EditLine's Text()
// before which the caret is currently set.

// If a selection is active, start is the index of the first selected UTF-8
// character in Text() and end is the position one character after the end of
// the selection. The selected text is thus
//     e.Text()[start:end]
func (e *EditLine) CursorPosition() (start, end int) {
	if e.handle != 0 {
		w32.SendMessage(
			e.handle,
			w32.EM_GETSEL,
			uintptr(unsafe.Pointer(&e.cursorStart)),
			uintptr(unsafe.Pointer(&e.cursorEnd)),
		)
	}
	return e.cursorStart, e.cursorEnd
}

func (e *EditLine) SetCursorPosition(pos int) {
	e.setCursor(pos, pos)
}

func (e *EditLine) SetSelection(start, end int) {
	e.setCursor(start, end)
}

func (e *EditLine) setCursor(start, end int) {
	e.cursorStart = start
	e.cursorEnd = end

	if e.handle != 0 {
		w32.SendMessage(
			e.handle,
			w32.EM_SETSEL,
			uintptr(e.cursorStart),
			uintptr(e.cursorEnd),
		)
	} else {
		e.clampCursorToText()
	}
}

func (e *EditLine) clampCursorToText() {
	// If called before we have a window, we have to handle clamping of the
	// positions ourselves.
	n := utf8.RuneCountInString(e.Text())
	if e.cursorStart < 0 {
		e.cursorStart = 0
	}
	if e.cursorStart > n {
		e.cursorStart = n
	}

	if e.cursorEnd < 0 {
		e.cursorEnd = 0
	}
	if e.cursorEnd > n {
		e.cursorEnd = n
	}

	if e.cursorEnd < e.cursorStart {
		e.cursorStart, e.cursorEnd = e.cursorEnd, e.cursorStart
	}
}

func (e *EditLine) SetOnTextChange(f func()) {
	e.onTextChange = f
}

func (e *EditLine) OnTextChange() func() {
	return e.onTextChange
}
