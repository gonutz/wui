//+build windows

package wui

import "github.com/gonutz/w32"

// TODO Rename to CheckBox.

func NewCheckbox() *Checkbox {
	return &Checkbox{}
}

type Checkbox struct {
	textControl
	checked  bool
	onChange func(bool)
}

func (c *Checkbox) create(id int) {
	c.textControl.create(id, 0, "BUTTON", w32.WS_TABSTOP|w32.BS_AUTOCHECKBOX)
	w32.SendMessage(c.handle, w32.BM_SETCHECK, toCheckState(c.checked), 0)
}

func (c *Checkbox) Checked() bool {
	return c.checked
}

func (c *Checkbox) SetChecked(checked bool) {
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

func (c *Checkbox) SetOnChange(f func(checked bool)) {
	c.onChange = f
}

func (c *Checkbox) handleNotification(cmd uintptr) {
	if cmd == w32.BN_CLICKED {
		c.checked = w32.SendMessage(c.handle, w32.BM_GETCHECK, 0, 0) == w32.BST_CHECKED
		if c.onChange != nil {
			c.onChange(c.checked)
		}
	}
}
