package wui

import (
	"syscall"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

func NewStringTable(header1 string, headers ...string) *StringTable {
	return &StringTable{
		headers:  append([]string{header1}, headers...),
		items:    make([]string, 0, 1),
		selected: -1,
	}
}

type StringTable struct {
	textControl
	headers           []string
	items             []string
	createdRows       int
	onSelectionChange func()
	selected          int
}

var _ Control = (*StringTable)(nil)

func (*StringTable) canFocus() bool {
	return true
}

func (s *StringTable) OnTabFocus() func() {
	return s.onTabFocus
}

func (s *StringTable) SetOnTabFocus(f func()) {
	s.onTabFocus = f
}

func (*StringTable) eatsTabs() bool {
	return false
}

func (c *StringTable) create(id int) {
	c.textControl.create(
		id,
		w32.WS_EX_CLIENTEDGE,
		"SysListView32",
		w32.WS_TABSTOP|w32.LVS_REPORT|w32.LVS_SINGLESEL|w32.LVS_NOSORTHEADER|
			w32.LVS_SHOWSELALWAYS,
	)
	w32.SendMessage(c.handle, w32.LVM_SETEXTENDEDLISTVIEWSTYLE, 0,
		w32.LVS_EX_FULLROWSELECT|w32.LVS_EX_DOUBLEBUFFER|w32.LVS_EX_GRIDLINES)

	hdc := w32.GetDC(c.handle)
	defer w32.ReleaseDC(c.handle, hdc)
	var maxW int32
	for i := range c.headers {
		size, ok := w32.GetTextExtentPoint32(hdc, c.headers[i])
		if ok && size.CX > maxW {
			maxW = size.CX
		}
	}
	for i := range c.headers {
		header, _ := syscall.UTF16PtrFromString(c.headers[i])
		var w int32 = 5
		size, ok := w32.GetTextExtentPoint32(hdc, c.headers[i])
		if ok {
			w = size.CX
		}
		w32.SendMessage(c.handle, w32.LVM_INSERTCOLUMN, uintptr(i), uintptr(unsafe.Pointer(
			&w32.LVCOLUMN{
				Mask:     w32.LVCF_FMT | w32.LVCF_WIDTH | w32.LVCF_TEXT | w32.LVCF_SUBITEM,
				Fmt:      w32.LVCFMT_CENTER,
				Cx:       w + 12, // we need a margin or the headers will not be fully displayed
				PszText:  header,
				ISubItem: int32(i + 1),
			})))
	}
	for i, item := range c.items {
		c.SetCell(c.indexToCol(i), c.indexToRow(i), item)
	}
}

func (c *StringTable) SetCell(col, row int, s string) {
	if col < 0 || col >= len(c.headers) {
		return
	}
	if c.handle == 0 {
		// add the item to our internal item list
		i := c.toItemIndex(col, row)
		if i >= len(c.items) {
			if i < cap(c.items) {
				c.items = c.items[:i+1]
			} else {
				c.items = append(c.items, make([]string, i-len(c.items)+1)...)
			}
		}
		c.items[i] = s
	} else {
		// make sure there are enough rows available
		for c.createdRows <= row {
			w32.SendMessage(
				c.handle,
				w32.LVM_INSERTITEM,
				0,
				uintptr(unsafe.Pointer(&w32.LVITEM{IItem: int32(c.createdRows)})),
			)
			c.createdRows++
		}
		// set the cell's text
		text, _ := syscall.UTF16PtrFromString(s)
		w32.SendMessage(c.handle, w32.LVM_SETITEMTEXT, uintptr(row),
			uintptr(unsafe.Pointer(&w32.LVITEM{
				Mask:     w32.LVIF_TEXT,
				PszText:  text,
				ISubItem: int32(col),
			})))
	}
}

func (c *StringTable) toItemIndex(col, row int) int {
	return col + row*len(c.headers)
}

func (c *StringTable) indexToCol(i int) int {
	return i % len(c.headers)
}

func (c *StringTable) indexToRow(i int) int {
	return i / len(c.headers)
}

func (c *StringTable) lockOnSelectionChange() (unlock func()) {
	lock := c.onSelectionChange
	c.onSelectionChange = nil
	return func() {
		c.onSelectionChange = lock
	}
}

func (c *StringTable) DeleteRow(row int) {
	rows := c.RowCount()
	if 0 <= row && row < rows {
		curSel := c.selected
		if curSel >= row {
			defer func() {
				c.selected = -1
				if curSel == rows-1 {
					c.newItemSelected(curSel - 1)
				} else {
					c.newItemSelected(curSel)
				}
			}()
		}
		defer c.lockOnSelectionChange()()
		if c.handle != 0 {
			w32.SendMessage(c.handle, w32.LVM_DELETEITEM, uintptr(row), 0)
			c.createdRows--
			if c.createdRows > 0 && c.HasFocus() {
				// make sure the selection is still active
				press := func(key uintptr) {
					w32.SendMessage(c.handle, w32.WM_KEYDOWN, key, 0)
					w32.SendMessage(c.handle, w32.WM_KEYUP, key, 0)
				}
				if c.createdRows == 1 {
					press(w32.VK_UP)
				} else if row == 0 {
					press(w32.VK_DOWN)
					press(w32.VK_UP)
				} else {
					press(w32.VK_UP)
					press(w32.VK_DOWN)
				}
			}
		}
	}
}

func (c *StringTable) RowCount() int {
	if c.handle == 0 {
		return (len(c.items) + len(c.headers) - 1) / len(c.headers)
	} else {
		return c.createdRows
	}
}

func (c *StringTable) ColCount() int {
	return len(c.headers)
}

func (c *StringTable) SelectedRow() int {
	return c.selected
}

func (c *StringTable) newItemSelected(i int) {
	if i != c.selected {
		c.selected = i
		if c.onSelectionChange != nil {
			c.onSelectionChange()
		}
	}
}

func (c *StringTable) itemDeselected() {
	c.newItemSelected(-1)
}

func (c *StringTable) Clear() {
	defer c.itemDeselected()
	defer c.lockOnSelectionChange()()

	for i := c.RowCount() - 1; i >= 0; i-- {
		w32.SendMessage(c.handle, w32.LVM_DELETEITEM, uintptr(i), 0)
	}
	c.createdRows = 0
}

func (c *StringTable) SetOnSelectionChange(f func()) {
	c.onSelectionChange = f
}

func (c *StringTable) OnSelectionChange() func() {
	return c.onSelectionChange
}
