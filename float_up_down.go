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
		min:       math.Inf(-1),
		max:       math.Inf(1),
		precision: 1,
	}
}

type FloatUpDown struct {
	textEditControl
	upDownHandle  w32.HWND
	value         float64
	min           float64
	max           float64
	precision     int
	onValueChange func(value float64)
}

var _ Control = (*FloatUpDown)(nil)

func (*FloatUpDown) canFocus() bool {
	return true
}

func (*FloatUpDown) eatsTabs() bool {
	return false
}

func (n *FloatUpDown) create(id int) {
	sanitize := func(text string) string {
		var newText string
		dotPos := -1
		for _, r := range text {
			if (r == '+' || r == '-') && len(newText) == 0 {
				newText += string(r)
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
		if newText == "" || newText == "-" || newText == "+" {
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
	n.textEditControl.create(
		id,
		w32.WS_EX_CLIENTEDGE,
		"EDIT",
		w32.WS_TABSTOP,
	)
	w32.SetWindowSubclass(n.textEditControl.handle, syscall.NewCallback(func(
		window w32.HWND,
		msg uint32,
		wParam, lParam uintptr,
		subclassID uintptr,
		refData uintptr,
	) uintptr {
		switch msg {
		case w32.WM_KILLFOCUS:
			text := n.textEditControl.Text()
			newText := sanitize(text)
			if newText != text {
				n.textEditControl.SetText(newText)
			}
			return w32.DefSubclassProc(window, msg, wParam, lParam)
		case w32.WM_CHAR:
			// these are the codes sent for the respective edit operations
			const (
				selectAll = 1  // for Ctrl+A
				copy      = 3  // Ctrl+C
				paste     = 22 // Ctrl+V
				cut       = 24 // Ctrl+X
			)
			if '0' <= wParam && wParam <= '9' ||
				wParam == '+' || wParam == '-' ||
				wParam == '.' || wParam == ',' ||
				wParam == 'i' || wParam == 'I' ||
				wParam == 'n' || wParam == 'N' ||
				wParam == 'f' || wParam == 'F' ||
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
	n.textEditControl.SetBounds(x, y, width, height)
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
	if n.textEditControl.handle != 0 {
		text := n.textEditControl.Text()
		t := strings.Replace(text, ",", ".", 1)
		if x, err := strconv.ParseFloat(t, 64); err == nil {
			n.value = x
		} else {
			lower := strings.ToLower(text)
			if strings.HasSuffix(lower, "inf") {
				if strings.Count(lower, "-")%2 == 0 {
					n.value = math.Inf(1)
				} else {
					n.value = math.Inf(-1)
				}
			}
		}
	}
	withRightPrecision := strconv.FormatFloat(n.value, 'f', n.precision, 64)
	n.value, _ = strconv.ParseFloat(withRightPrecision, 64)
	return n.value
}

// SetValue does not accept NaN (not a number). It does accept infinity, the
// user can type "+inf" or "-inf".
func (n *FloatUpDown) SetValue(f float64) {
	if math.IsNaN(f) {
		return
	}
	n.value = f
	if n.value < n.min {
		n.value = n.min
	}
	if n.value > n.max {
		n.value = n.max
	}
	if n.textEditControl.handle != 0 {
		w32.SetWindowText(
			n.textEditControl.handle,
			strconv.FormatFloat(n.value, 'f', n.precision, 64),
		)
	}
}

func (n *FloatUpDown) Min() float64 {
	return n.min
}

func (n *FloatUpDown) SetMin(min float64) {
	if n.Value() < min {
		n.SetValue(min)
	}
	n.min = min
}

func (n *FloatUpDown) Max() float64 {
	return n.max
}

func (n *FloatUpDown) SetMax(max float64) {
	if n.Value() > max {
		n.SetValue(max)
	}
	n.max = max
}

func (n *FloatUpDown) MinMax() (min, max float64) {
	return n.min, n.max
}

func (n *FloatUpDown) SetMinMax(min, max float64) {
	if n.Value() < min {
		n.SetValue(min)
	} else if n.Value() > max {
		n.SetValue(max)
	}
	n.min = min
	n.max = max
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

func (n *FloatUpDown) OnValueChange() func(value float64) {
	return n.onValueChange
}

func (n *FloatUpDown) SetOnValueChange(f func(value float64)) {
	n.onValueChange = f
}

func (n *FloatUpDown) handleNotification(cmd uintptr) {
	if cmd == w32.EN_CHANGE && n.onValueChange != nil {
		n.onValueChange(n.Value())
	}
}
