//+build windows

package wui

import (
	"math"
	"strconv"
	"strings"
	"syscall"

	"github.com/gonutz/w32"
)

func NewFloatUpDown() *FloatUpDown {
	return &FloatUpDown{
		minValue:  math.Inf(-1),
		maxValue:  math.Inf(1),
		precision: 1,
	}
}

type FloatUpDown struct {
	textControl
	upDownHandle w32.HWND
	value        float64
	minValue     float64
	maxValue     float64
	precision    int
}

func (n *FloatUpDown) create(id int) {
	sanitize := func(text string) string {
		var newText string
		dotPos := -1
		for _, r := range text {
			if r == '-' && len(newText) == 0 {
				newText += "-"
			}
			if r == '.' || r == ',' {
				if dotPos == -1 {
					newText += "."
					dotPos = len(newText) - 1
				}
			}
			if '0' <= r && r <= '9' {
				newText += string(r)
				if dotPos != -1 && len(newText)-dotPos-1 >= n.precision {
					break
				}
			}
		}
		if newText == "" || newText == "-" {
			newText += "0"
		}
		if dotPos == -1 {
			newText += "." + strings.Repeat("0", n.precision)
		} else {
			newText += strings.Repeat("0", n.precision-(len(newText)-dotPos-1))
		}
		return newText
	}

	// the main handle is for the edit field
	n.text = strconv.FormatFloat(n.value, 'f', n.precision, 64)
	n.textControl.create(
		id,
		w32.WS_EX_CLIENTEDGE,
		"EDIT",
		w32.WS_TABSTOP,
	)
	w32.SetWindowSubclass(n.textControl.handle, syscall.NewCallback(func(
		window w32.HWND,
		msg uint32,
		wParam, lParam uintptr,
		subclassID uintptr,
		refData uintptr,
	) uintptr {
		switch msg {
		case w32.WM_KILLFOCUS:
			text := n.textControl.Text()
			newText := sanitize(text)
			if newText != text {
				n.textControl.SetText(newText)
			}
			return w32.DefSubclassProc(window, msg, wParam, lParam)
		case w32.WM_CHAR:
			// these are the codes sent for the respective edit operations
			const (
				selectAll = 1
				copy      = 3
				cut       = 24
				paste     = 22
			)
			if '0' <= wParam && wParam <= '9' ||
				wParam == '-' ||
				wParam == '.' || wParam == ',' ||
				wParam == w32.VK_RETURN ||
				wParam == w32.VK_BACK ||
				wParam == w32.VK_DELETE ||
				wParam == selectAll ||
				wParam == copy || wParam == cut || wParam == paste {
				return w32.DefSubclassProc(window, msg, wParam, lParam)
			}
			return 0
		default:
			return w32.DefSubclassProc(window, msg, wParam, lParam)
		}
	}), 0, 0)
	var visible uint
	if !n.hidden {
		visible = w32.WS_VISIBLE
	}
	upDown := w32.CreateWindowStr(
		w32.UPDOWN_CLASS,
		"",
		visible|w32.WS_CHILD|
			w32.UDS_ALIGNRIGHT|w32.UDS_NOTHOUSANDS|w32.UDS_ARROWKEYS,
		n.x, n.y, n.width, n.height,
		n.parent.getHandle(), w32.HMENU(id), n.parent.getInstance(), nil,
	)
	w32.SendMessage(upDown, w32.UDM_SETBUDDY, uintptr(n.handle), 0)
	n.upDownHandle = upDown
}

func (n *FloatUpDown) SetX(x int) {
	n.SetBounds(x, n.y, n.width, n.height)
}

func (n *FloatUpDown) SetY(y int) {
	n.SetBounds(n.x, y, n.width, n.height)
}

func (n *FloatUpDown) SetPos(x, y int) {
	n.SetBounds(x, y, n.width, n.height)
}

func (n *FloatUpDown) SetWidth(width int) {
	n.SetBounds(n.x, n.y, width, n.height)
}

func (n *FloatUpDown) SetHeight(height int) {
	n.SetBounds(n.x, n.y, n.width, height)
}

func (n *FloatUpDown) SetSize(width, height int) {
	n.SetBounds(n.x, n.y, width, height)
}

func (n *FloatUpDown) SetBounds(x, y, width, height int) {
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

func (n *FloatUpDown) Value() float64 {
	if n.textControl.handle != 0 {
		t := strings.Replace(n.textControl.Text(), ",", ".", 1)
		n.value, _ = strconv.ParseFloat(t, 64)
	}
	return n.value
}

func (n *FloatUpDown) SetValue(f float64) {
	n.value = f
	if n.value < n.minValue {
		n.value = n.minValue
	}
	if n.value > n.maxValue {
		n.value = n.maxValue
	}
	if n.textControl.handle != 0 {
		w32.SetWindowText(
			n.textControl.handle,
			strconv.FormatFloat(n.value, 'f', n.precision, 64),
		)
	}
}

func (n *FloatUpDown) MinValue() float64 {
	return n.minValue
}

func (n *FloatUpDown) SetMinValue(min float64) {
	if n.Value() < min {
		n.SetValue(min)
	}
	n.minValue = min
}

func (n *FloatUpDown) MaxValue() float64 {
	return n.maxValue
}

func (n *FloatUpDown) SetMaxValue(max float64) {
	if n.Value() > max {
		n.SetValue(max)
	}
	n.maxValue = max
}

func (n *FloatUpDown) MinMaxValues() (min, max int) {
	return int(n.minValue), int(n.maxValue)
}

func (n *FloatUpDown) SetMinMaxValues(min, max float64) {
	if n.Value() < min {
		n.SetValue(min)
	} else if n.Value() > max {
		n.SetValue(max)
	}
	n.minValue = min
	n.maxValue = max
}

func (n *FloatUpDown) Precision() int {
	return n.precision
}

func (n *FloatUpDown) SetPrecision(p int) {
	if p < 1 {
		p = 1
	}
	if p > 6 {
		p = 6
	}
	n.precision = p
}
