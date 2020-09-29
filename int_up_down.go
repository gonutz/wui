//+build windows

package wui

import (
	"math"
	"strconv"

	"github.com/gonutz/w32"
)

// TODO Typing + or - into an IntUpDown shows an error. The same might be true
// for FloatUpDowns.

func NewIntUpDown() *IntUpDown {
	return &IntUpDown{
		minValue: math.MinInt32,
		maxValue: math.MaxInt32,
	}
}

type IntUpDown struct {
	textControl
	upDownHandle  w32.HWND
	value         int32
	minValue      int32
	maxValue      int32
	onValueChange func(value int)
}

func (n *IntUpDown) create(id int) {
	// the main handle is for the edit field
	n.text = strconv.Itoa(int(n.value))
	n.textControl.create(
		id,
		w32.WS_EX_CLIENTEDGE,
		"EDIT",
		w32.WS_TABSTOP|w32.ES_NUMBER,
	)
	var visible uint
	if !n.hidden {
		visible = w32.WS_VISIBLE
	}
	upDown := w32.CreateWindowStr(
		w32.UPDOWN_CLASS,
		"",
		visible|w32.WS_CHILD|
			w32.UDS_SETBUDDYINT|w32.UDS_ALIGNRIGHT|w32.UDS_NOTHOUSANDS|
			w32.UDS_ARROWKEYS,
		n.x, n.y, n.width, n.height,
		n.parent.getHandle(), 0, n.parent.getInstance(), nil,
	)
	w32.SendMessage(upDown, w32.UDM_SETBUDDY, uintptr(n.handle), 0)
	w32.SendMessage(
		upDown,
		w32.UDM_SETRANGE32,
		uintptr(n.minValue),
		uintptr(n.maxValue),
	)
	n.upDownHandle = upDown
}

func (n *IntUpDown) SetX(x int) {
	n.SetBounds(x, n.y, n.width, n.height)
}

func (n *IntUpDown) SetY(y int) {
	n.SetBounds(n.x, y, n.width, n.height)
}

func (n *IntUpDown) SetPos(x, y int) {
	n.SetBounds(x, y, n.width, n.height)
}

func (n *IntUpDown) SetWidth(width int) {
	n.SetBounds(n.x, n.y, width, n.height)
}

func (n *IntUpDown) SetHeight(height int) {
	n.SetBounds(n.x, n.y, n.width, height)
}

func (n *IntUpDown) SetSize(width, height int) {
	n.SetBounds(n.x, n.y, width, height)
}

func (n *IntUpDown) SetBounds(x, y, width, height int) {
	n.textControl.SetBounds(x, y, width, height)
	if n.upDownHandle != 0 {
		w32.SetWindowPos(
			n.upDownHandle, 0,
			n.x, n.y, n.width, n.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
		w32.SendMessage(n.upDownHandle, w32.UDM_SETBUDDY, uintptr(n.handle), 0)
	}
}

func (n *IntUpDown) Value() int {
	if n.upDownHandle != 0 {
		n.value = int32(w32.SendMessage(n.upDownHandle, w32.UDM_GETPOS32, 0, 0))
	}
	return int(n.value)
}

func (n *IntUpDown) SetValue(v int) {
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
}

func (n *IntUpDown) MinValue() int {
	return int(n.minValue)
}

func (n *IntUpDown) SetMinValue(min int) {
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
}

func (n *IntUpDown) MaxValue() int {
	return int(n.maxValue)
}

func (n *IntUpDown) SetMaxValue(max int) {
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
}

func (n *IntUpDown) MinMaxValues() (min, max int) {
	return int(n.minValue), int(n.maxValue)
}

func (n *IntUpDown) SetMinMaxValues(min, max int) {
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
}

func (n *IntUpDown) SetOnValueChange(f func(value int)) {
	n.onValueChange = f
}

func (n *IntUpDown) SetVisible(v bool) {
	n.textControl.SetVisible(v)
	if n.upDownHandle != 0 {
		if v {
			w32.ShowWindow(n.upDownHandle, w32.SW_SHOW)
		} else {
			w32.ShowWindow(n.upDownHandle, w32.SW_HIDE)
		}
	}
}

func (n *IntUpDown) handleNotification(cmd uintptr) {
	if cmd == w32.EN_CHANGE && n.onValueChange != nil {
		n.onValueChange(n.Value())
	}
}
