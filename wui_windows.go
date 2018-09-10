//+build windows

package wui

import (
	"strconv"

	"github.com/gonutz/w32"
)

func createControl_(
	control Control,
	parent *Window,
	id int,
	instance w32.HINSTANCE,
) {
	switch c := control.(type) {
	case *NumberUpDown:

	case *Label:
		var visible uint
		if !c.hidden {
			visible = w32.WS_VISIBLE
		}
		c.handle = w32.CreateWindowExStr(
			0,
			"STATIC",
			c.text,
			visible|w32.WS_CHILD|w32.SS_CENTERIMAGE|c.align,
			c.x, c.y, c.width, c.height,
			parent.handle, w32.HMENU(id), instance, nil,
		)
		if c.font != nil {
			c.font.create()
			w32.SendMessage(
				c.handle,
				w32.WM_SETFONT,
				uintptr(c.font.handle),
				1,
			)
		} else if parent.font != nil {
			w32.SendMessage(
				c.handle,
				w32.WM_SETFONT,
				uintptr(parent.font.handle),
				1,
			)
		}
	case *Paintbox:
		c.handle = w32.CreateWindowExStr(
			0,
			"STATIC",
			"",
			w32.WS_VISIBLE|w32.WS_CHILD|w32.SS_OWNERDRAW,
			c.x, c.y, c.width, c.height,
			parent.handle, w32.HMENU(id), instance, nil,
		)
		if parent.font != nil {
			w32.SendMessage(
				c.handle,
				w32.WM_SETFONT,
				uintptr(parent.font.handle),
				1,
			)
		}
	case *Checkbox:
		c.handle = w32.CreateWindowExStr(
			0,
			"BUTTON",
			c.text,
			w32.WS_VISIBLE|w32.WS_CHILD|w32.WS_TABSTOP|w32.BS_AUTOCHECKBOX,
			c.x, c.y, c.width, c.height,
			parent.handle, w32.HMENU(id), instance, nil,
		)
		w32.SendMessage(c.handle, w32.BM_SETCHECK, toCheckState(c.checked), 0)
		if parent.font != nil {
			w32.SendMessage(
				c.handle,
				w32.WM_SETFONT,
				uintptr(parent.font.handle),
				1,
			)
		}
	case *EditLine:
		var visible uint
		if !c.hidden {
			visible = w32.WS_VISIBLE
		}
		c.handle = w32.CreateWindowExStr(
			w32.WS_EX_CLIENTEDGE,
			"EDIT",
			c.text,
			visible|w32.WS_CHILD|w32.WS_TABSTOP,
			c.x, c.y, c.width, c.height,
			parent.handle, w32.HMENU(id), instance, nil,
		)
		if parent.font != nil {
			w32.SendMessage(
				c.handle,
				w32.WM_SETFONT,
				uintptr(parent.font.handle),
				1,
			)
		}
	case *ProgressBar:
		var visible uint
		if !c.hidden {
			visible = w32.WS_VISIBLE
		}
		c.handle = w32.CreateWindowExStr(
			w32.WS_EX_CLIENTEDGE,
			w32.PROGRESS_CLASS,
			"",
			visible|w32.WS_CHILD,
			c.x, c.y, c.width, c.height,
			parent.handle, w32.HMENU(id), instance, nil,
		)
		w32.SendMessage(c.handle, w32.PBM_SETRANGE32, 0, maxProgressBarValue)
		c.SetValue(c.value)
		if parent.font != nil {
			w32.SendMessage(
				c.handle,
				w32.WM_SETFONT,
				uintptr(parent.font.handle),
				1,
			)
		}
	default:
		panic("unhandled control type")
	}
}
