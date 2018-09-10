//+build windows

package wui

import (
	"math"
	"strconv"

	"github.com/gonutz/w32"
)

func NewNumberUpDown() *NumberUpDown {
	return &NumberUpDown{
		minValue: math.MinInt32,
		maxValue: math.MaxInt32,
	}
}

type NumberUpDown struct {
	upDownHandle  w32.HWND
	editHandle    w32.HWND
	x             int
	y             int
	width         int
	height        int
	parent        container
	value         int32
	minValue      int32
	maxValue      int32
	disabled      bool
	hidden        bool
	onValueChange func(value int)
}

func (*NumberUpDown) isControl() {}

func (n *NumberUpDown) setParent(parent container) {
	n.parent = parent
}

func (n *NumberUpDown) create(id int) {
	var visible uint
	if !n.hidden {
		visible = w32.WS_VISIBLE
	}
	upDown := w32.CreateWindowStr(
		w32.UPDOWN_CLASS,
		"",
		visible|w32.WS_CHILD|
			w32.UDS_SETBUDDYINT|w32.UDS_ALIGNRIGHT|w32.UDS_ARROWKEYS,
		n.x, n.y, n.width, n.height,
		n.parent.getHandle(), 0, n.parent.getInstance(), nil,
	)
	edit := w32.CreateWindowExStr(
		w32.WS_EX_CLIENTEDGE,
		"EDIT",
		strconv.Itoa(int(n.value)),
		visible|w32.WS_CHILD|w32.WS_TABSTOP|w32.ES_NUMBER,
		n.x, n.y, n.width, n.height,
		n.parent.getHandle(), w32.HMENU(id), n.parent.getInstance(), nil,
	)
	if n.disabled {
		w32.EnableWindow(edit, false)
	}
	w32.SendMessage(upDown, w32.UDM_SETBUDDY, uintptr(edit), 0)
	w32.SendMessage(
		upDown,
		w32.UDM_SETRANGE32,
		uintptr(n.minValue),
		uintptr(n.maxValue),
	)
	n.upDownHandle = upDown
	n.editHandle = edit
	// TODO handle font
	//if n.parent.font != nil {
	//	w32.SendMessage(
	//		edit,
	//		w32.WM_SETFONT,
	//		uintptr(n.parent.font.handle),
	//		1,
	//	)
	//}
}

func (n *NumberUpDown) parentFontChanged() {
	// TODO
}

func (n *NumberUpDown) Enabled() bool {
	return !n.disabled
}

func (n *NumberUpDown) SetEnabled(e bool) *NumberUpDown {
	n.disabled = !e
	if n.editHandle != 0 {
		w32.EnableWindow(n.editHandle, e)
	}
	return n
}

func (n *NumberUpDown) Value() int {
	if n.upDownHandle != 0 {
		n.value = int32(w32.SendMessage(n.upDownHandle, w32.UDM_GETPOS32, 0, 0))
	}
	return int(n.value)
}

func (n *NumberUpDown) SetValue(v int) *NumberUpDown {
	n.value = int32(v)
	if n.value < n.minValue {
		n.value = n.minValue
	}
	if n.value > n.maxValue {
		n.value = n.maxValue
	}
	if n.upDownHandle != 0 {
		w32.SendMessage(n.upDownHandle, w32.UDM_SETPOS32, 0, uintptr(v))
	}
	return n
}

func (n *NumberUpDown) SetMinValue(min int) *NumberUpDown {
	if n.Value() < min {
		n.SetValue(min)
	}
	n.minValue = int32(min)
	if n.upDownHandle != 0 {
		w32.SendMessage(
			n.upDownHandle,
			w32.UDM_SETRANGE32,
			uintptr(n.minValue),
			uintptr(n.maxValue),
		)
	}
	return n
}

func (n *NumberUpDown) SetMaxValue(max int) *NumberUpDown {
	if n.Value() > max {
		n.SetValue(max)
	}
	n.maxValue = int32(max)
	if n.upDownHandle != 0 {
		w32.SendMessage(
			n.upDownHandle,
			w32.UDM_SETRANGE32,
			uintptr(n.minValue),
			uintptr(n.maxValue),
		)
	}
	return n
}

func (n *NumberUpDown) SetMinMaxValues(min, max int) *NumberUpDown {
	if n.Value() < min {
		n.SetValue(min)
	} else if n.Value() > max {
		n.SetValue(max)
	}
	n.minValue = int32(min)
	n.maxValue = int32(max)
	if n.upDownHandle != 0 {
		w32.SendMessage(
			n.upDownHandle,
			w32.UDM_SETRANGE32,
			uintptr(n.minValue),
			uintptr(n.maxValue),
		)
	}
	return n
}

func (n *NumberUpDown) SetBounds(x, y, width, height int) *NumberUpDown {
	n.x = x
	n.y = y
	n.width = width
	n.height = height
	if n.editHandle != 0 {
		w32.SetWindowPos(
			n.editHandle, 0,
			n.x, n.y, n.width, n.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
		w32.SetWindowPos(
			n.upDownHandle, 0,
			n.x, n.y, n.width, n.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
		w32.SendMessage(
			n.upDownHandle,
			w32.UDM_SETBUDDY,
			uintptr(n.editHandle),
			0,
		)
	}
	return n
}

func (n *NumberUpDown) SetOnValueChange(f func(value int)) *NumberUpDown {
	n.onValueChange = f
	return n
}
