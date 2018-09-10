//+build windows

package wui

import (
	"math"
	"strconv"

	"github.com/gonutz/w32"
)

type NumberUpDown struct {
	upDownHandle  w32.HWND
	editHandle    w32.HWND
	parent        *Window
	x             int
	y             int
	width         int
	height        int
	value         int32
	minValue      int32
	maxValue      int32
	disabled      bool
	onValueChange func(value int)
}

func (*NumberUpDown) isControl() {}

func NewNumberUpDown() *NumberUpDown {
	return &NumberUpDown{
		minValue: math.MinInt32,
		maxValue: math.MaxInt32,
	}
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

func (n *NumberUpDown) create(parent *Window, id int, instance w32.HINSTANCE) {
	n.parent = parent
	upDown := w32.CreateWindowStr(
		w32.UPDOWN_CLASS,
		"",
		w32.WS_VISIBLE|w32.WS_CHILD|
			w32.UDS_SETBUDDYINT|w32.UDS_ALIGNRIGHT|w32.UDS_ARROWKEYS,
		n.x, n.y, n.width, n.height,
		parent.handle, 0, instance, nil,
	)
	edit := w32.CreateWindowExStr(
		w32.WS_EX_CLIENTEDGE,
		"EDIT",
		strconv.Itoa(int(n.value)),
		w32.WS_TABSTOP|w32.WS_VISIBLE|w32.WS_CHILD|w32.ES_NUMBER,
		n.x, n.y, n.width, n.height,
		parent.handle, w32.HMENU(id), instance, nil,
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
	if parent.font != nil {
		w32.SendMessage(
			edit,
			w32.WM_SETFONT,
			uintptr(parent.font.handle),
			1,
		)
	}
}
