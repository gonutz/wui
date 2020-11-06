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

var _ Control = (*ComboBox)(nil)

func (*ComboBox) canFocus() bool {
	return true
}

func (*ComboBox) eatsTabs() bool {
	return false
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

func (e *ComboBox) AddItem(s string) {
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

func (e *ComboBox) SetItems(items []string) {
	e.items = items
	if e.handle != 0 {
		w32.SendMessage(e.handle, w32.CB_RESETCONTENT, 0, 0)
		for _, s := range e.items {
			e.addItem(s)
		}
	}
}

func (e *ComboBox) SelectedIndex() int {
	if e.handle != 0 {
		e.selected = int(w32.SendMessage(e.handle, w32.CB_GETCURSEL, 0, 0))
	}
	return e.selected
}

// SetSelectedIndex sets the current index. Set -1 to remove any selection and
// make the combo box empty. If you pass a value < -1 it will be set to -1
// instead. The index is not clamped to the number of items, you may set the
// index 10 where only 5 items are set. This lets you set the index and items at
// design time without one invalidating the other. At runtime though,
// SelectedIndex() will return -1 if you set an invalid index.
func (e *ComboBox) SetSelectedIndex(i int) {
	if i < -1 {
		i = -1
	}
	e.selected = i
	if e.handle != 0 {
		w32.SendMessage(e.handle, w32.CB_SETCURSEL, uintptr(i), 0)
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
