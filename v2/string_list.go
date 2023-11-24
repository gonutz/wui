package wui

import (
	"syscall"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

func NewStringList() *StringList {
	return &StringList{selected: -1}
}

type StringList struct {
	textControl
	items    []string
	selected int
	onChange func(newIndex int)
}

var _ Control = (*StringList)(nil)

func (l *StringList) closing() {
	l.SelectedIndex()
}

func (*StringList) canFocus() bool {
	return true
}

func (*StringList) eatsTabs() bool {
	return false
}

func (l *StringList) create(id int) {
	l.textControl.create(
		id,
		w32.WS_EX_CLIENTEDGE,
		"LISTBOX",
		w32.WS_VSCROLL|w32.WS_TABSTOP|w32.LBS_NOTIFY,
	)
	for _, s := range l.items {
		l.addItem(s)
	}
	l.SetSelectedIndex(l.selected)
}

func (l *StringList) AddItem(s string) {
	l.items = append(l.items, s)
	if l.handle != 0 {
		l.addItem(s)
	}
}

func (l *StringList) addItem(s string) {
	ptr, _ := syscall.UTF16PtrFromString(s)
	w32.SendMessage(l.handle, w32.LB_ADDSTRING, 0, uintptr(unsafe.Pointer(ptr)))
}

func (l *StringList) Clear() {
	l.items = nil
	if l.handle != 0 {
		w32.SendMessage(l.handle, w32.LB_RESETCONTENT, 0, 0)
	}
}

func (l *StringList) Items() []string {
	return l.items
}

func (l *StringList) SetItems(items []string) {
	l.items = items
	if l.handle != 0 {
		w32.SendMessage(l.handle, w32.LB_RESETCONTENT, 0, 0)
		for _, s := range l.items {
			l.addItem(s)
		}
	}
}

func (l *StringList) SelectedIndex() int {
	if l.handle != 0 {
		l.selected = int(w32.SendMessage(l.handle, w32.LB_GETCURSEL, 0, 0))
	}
	return l.selected
}

// SetSelectedIndex sets the current index. Set -1 to remove any selection and
// make the list box empty. If you pass a value < -1 it will be set to -1
// instead. The index is not clamped to the number of items, you may set the
// index 10 where only 5 items are set. This lets you set the index and items at
// design time without one invalidating the other. At runtime though,
// SelectedIndex() will return -1 if you set an invalid index.
func (l *StringList) SetSelectedIndex(i int) {
	if i < -1 {
		i = -1
	}
	l.selected = i
	if l.handle != 0 {
		w32.SendMessage(l.handle, w32.LB_SETCURSEL, uintptr(i), 0)
		if l.onChange != nil {
			l.onChange(i)
		}
	}
}

func (l *StringList) OnChange() func(newIndex int) {
	return l.onChange
}

func (l *StringList) SetOnChange(f func(newIndex int)) {
	l.onChange = f
}

func (l *StringList) handleNotification(cmd uintptr) {
	if cmd == w32.LBN_SELCHANGE && l.onChange != nil {
		l.onChange(l.SelectedIndex())
	}
}
