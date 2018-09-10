//+build windows

package wui

import "github.com/gonutz/w32"

type Checkbox struct {
	handle   w32.HWND
	parent   *Window
	x        int
	y        int
	width    int
	height   int
	hidden   bool
	disabled bool
	text     string
	font     *Font
	checked  bool
	onChange func(bool)
}

func NewCheckbox() *Checkbox {
	return &Checkbox{}
}

func (c *Checkbox) Checked() bool {
	return c.checked
}

func (c *Checkbox) SetChecked(checked bool) *Checkbox {
	if checked == c.checked {
		return c
	}
	c.checked = checked
	if c.handle != 0 {
		w32.SendMessage(c.handle, w32.BM_SETCHECK, toCheckState(c.checked), 0)
	}
	if c.onChange != nil {
		c.onChange(c.checked)
	}
	return c
}

func toCheckState(checked bool) uintptr {
	if checked {
		return w32.BST_CHECKED
	}
	return w32.BST_UNCHECKED
}

func (c *Checkbox) SetOnChange(f func(checked bool)) *Checkbox {
	c.onChange = f
	return c
}
