//+build windows

package wui

import (
	"syscall"
	"unsafe"

	"github.com/gonutz/w32"
)

func NewComboBox() *ComboBox {
	return &ComboBox{selected: -1}
}

type ComboBox struct {
	textControl
	items    []string
	selected int
	onChange func(newIndex int)
}

func (e *ComboBox) create(id int) {
	e.textControl.create(
		id,
		w32.WS_EX_CLIENTEDGE,
		"COMBOBOX",
		w32.WS_TABSTOP|w32.CBS_DROPDOWNLIST,
	)
	for _, s := range e.items {
		e.addItem(s)
	}
	e.SetSelectedIndex(e.selected)
}

func (e *ComboBox) Add(s string) {
	e.items = append(e.items, s)
	if e.handle != 0 {
		e.addItem(s)
	}
}

func (e *ComboBox) addItem(s string) {
	ptr, _ := syscall.UTF16PtrFromString(s)
	w32.SendMessage(e.handle, w32.CB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr)))
}

func (e *ComboBox) Clear() {
	e.items = nil
	if e.handle != 0 {
		w32.SendMessage(e.handle, w32.CB_RESETCONTENT, 0, 0)
	}
}

func (e *ComboBox) Items() []string {
	return e.items
}

func (e *ComboBox) SelectedIndex() int {
	if e.handle != 0 {
		e.selected = int(w32.SendMessage(e.handle, w32.CB_GETCURSEL, 0, 0))
	}
	return e.selected
}

func (e *ComboBox) SetSelectedIndex(i int) {
	oldI := e.selected
	if 0 <= i && i < len(e.items) {
		e.selected = i
	} else {
		e.selected = -1
	}
	if e.handle != 0 {
		w32.SendMessage(e.handle, w32.CB_SETCURSEL, uintptr(i), 0)
	}
	if i != oldI && e.onChange != nil {
		e.onChange(i)
	}
}

func (e *ComboBox) SetOnChange(f func(newIndex int)) {
	e.onChange = f
}

func (e *ComboBox) handleNotification(cmd uintptr) {
	if cmd == w32.CBN_SELCHANGE && e.onChange != nil {
		e.onChange(e.SelectedIndex())
	}
}
