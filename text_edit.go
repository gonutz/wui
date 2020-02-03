//+build windows

package wui

import (
	"syscall"
	"unicode"

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
		w32.WS_TABSTOP|w32.WS_VSCROLL|w32.WS_HSCROLL|
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
			if wParam == 1 && w32.GetKeyState(w32.VK_CONTROL)&0x8000 != 0 {
				// Ctrl+A was pressed - select all text.
				e.SelectAll()
				return 0
			}
			if wParam == 127 {
				// Ctrl+Backspace was pressed, if there is currently a selected
				// active, delete it. If there is just the cursor, delete the
				// last word before the cursor.
				text := []rune(e.Text())
				start, end := e.CursorPosition()
				if start != end {
					// There is a selection, delete it.
					e.SetText(string(append(text[:start], text[end:]...)))
					e.SetCursorPosition(start)
				} else {
					// No selection, delete the last word before the cursor.
					newText, newCursor := deleteWordBeforeCursor(text, start)
					e.SetText(newText)
					e.SetCursorPosition(newCursor)
				}
				return 0
			}
			return w32.DefSubclassProc(window, msg, wParam, lParam)
		default:
			return w32.DefSubclassProc(window, msg, wParam, lParam)
		}
	}), 0, 0)
}

func deleteWordBeforeCursor(text []rune, cursor int) (newText string, newCursor int) {
	prefix := text[:cursor]
	n := len(prefix)

	if n >= 2 && prefix[n-2] == '\r' && prefix[n-1] == '\n' {
		prefix = prefix[:n-2]
	} else if n <= 1 {
		prefix = nil
	} else {
		if unicode.IsSpace(prefix[n-1]) {
			for n > 0 && unicode.IsSpace(prefix[n-1]) {
				prefix = prefix[:n-1]
				n--
			}
		}
		for n > 0 && !unicode.IsSpace(prefix[n-1]) {
			prefix = prefix[:n-1]
			n--
		}
	}

	newText = string(append(prefix, text[cursor:]...))
	newCursor = len(prefix)
	return
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
