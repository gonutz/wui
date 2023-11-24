//go:build windows
// +build windows

package wui

import "github.com/gonutz/w32/v2"

func NewCheckBox() *CheckBox {
	return &CheckBox{}
}

type CheckBox struct {
	textControl
	checked  bool
	onChange func(bool)
}

var _ Control = (*CheckBox)(nil)

func (*CheckBox) canFocus() bool {
	return true
}

func (c *CheckBox) OnTabFocus() func() {
	return c.onTabFocus
}

func (c *CheckBox) SetOnTabFocus(f func()) {
	c.onTabFocus = f
}

func (*CheckBox) eatsTabs() bool {
	return false
}

func (c *CheckBox) create(id int) {
	c.textControl.create(id, 0, "BUTTON", w32.WS_TABSTOP|w32.BS_AUTOCHECKBOX)
	w32.SendMessage(c.handle, w32.BM_SETCHECK, toCheckState(c.checked), 0)
}

func (c *CheckBox) Checked() bool {
	return c.checked
}

func (c *CheckBox) SetChecked(checked bool) {
	if checked == c.checked {
		return
	}
	c.checked = checked
	if c.handle != 0 {
		w32.SendMessage(c.handle, w32.BM_SETCHECK, toCheckState(c.checked), 0)
	}
	if c.onChange != nil {
		c.onChange(c.checked)
	}
	return
}

func toCheckState(checked bool) uintptr {
	if checked {
		return w32.BST_CHECKED
	}
	return w32.BST_UNCHECKED
}

func (c *CheckBox) SetOnChange(f func(checked bool)) {
	c.onChange = f
}

func (c *CheckBox) handleNotification(cmd uintptr) {
	if cmd == w32.BN_CLICKED {
		c.checked = w32.SendMessage(c.handle, w32.BM_GETCHECK, 0, 0) == w32.BST_CHECKED
		if c.onChange != nil {
			c.onChange(c.checked)
		}
	}
}
