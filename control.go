//+build windows

package wui

import "github.com/gonutz/w32"

type control struct {
	handle   w32.HWND
	x        int
	y        int
	width    int
	height   int
	parent   container
	disabled bool
	hidden   bool
}

func (*control) isControl() {}

func (c *control) setParent(parent container) {
	c.parent = parent
}

func (c *control) create(id int, exStyle uint, className string, style uint) {
	var visible uint
	if !c.hidden {
		visible = w32.WS_VISIBLE
	}
	c.handle = w32.CreateWindowExStr(
		exStyle,
		className,
		"",
		visible|w32.WS_CHILD|style,
		c.x, c.y, c.width, c.height,
		c.parent.getHandle(), w32.HMENU(id), c.parent.getInstance(), nil,
	)
	if c.disabled {
		w32.EnableWindow(c.handle, false)
	}
}

func (c *control) parentFontChanged() {}

func (c *control) X() int {
	return c.x
}

func (c *control) SetX(x int) {
	c.SetBounds(x, c.y, c.width, c.height)
}

func (c *control) Y() int {
	return c.y
}

func (c *control) SetY(y int) {
	c.SetBounds(c.x, y, c.width, c.height)
}

func (c *control) Pos() (x, y int) {
	return c.x, c.y
}

func (c *control) SetPos(x, y int) {
	c.SetBounds(x, y, c.width, c.height)
}

func (c *control) Width() int {
	return c.width
}

func (c *control) SetWidth(width int) {
	c.SetBounds(c.x, c.y, width, c.height)
}

func (c *control) Height() int {
	return c.height
}

func (c *control) SetHeight(height int) {
	c.SetBounds(c.x, c.y, c.width, height)
}

func (c *control) Size() (width, height int) {
	return c.width, c.height
}

func (c *control) SetSize(width, height int) {
	c.SetBounds(c.x, c.y, width, height)
}

func (c *control) Bounds() (x, y, width, height int) {
	return c.x, c.y, c.width, c.height
}

func (c *control) SetBounds(x, y, width, height int) {
	c.x, c.y, c.width, c.height = x, y, width, height
	if c.handle != 0 {
		w32.SetWindowPos(
			c.handle, 0,
			c.x, c.y, c.width, c.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
}

func (c *control) Enabled() bool {
	return !c.disabled
}

func (c *control) SetEnabled(e bool) {
	c.disabled = !e
	if c.handle != 0 {
		w32.EnableWindow(c.handle, e)
	}
}

func (c *control) Visible() bool {
	return !c.hidden
}

func (c *control) SetVisible(v bool) {
	c.hidden = !v
	if c.handle != 0 {
		if v {
			w32.ShowWindow(c.handle, w32.SW_SHOW)
		} else {
			w32.ShowWindow(c.handle, w32.SW_HIDE)
		}

	}
}

type textControl struct {
	control
	text string
	font *Font
}

func (c *textControl) create(id int, exStyle uint, className string, style uint) {
	c.control.create(id, exStyle, className, style)
	w32.SetWindowText(c.handle, c.text)
	c.SetFont(c.font)
}

func (c *textControl) Text() string {
	if c.handle != 0 {
		c.text = w32.GetWindowText(c.handle)
	}
	return c.text
}

func (c *textControl) SetText(text string) {
	c.text = text
	if c.handle != 0 {
		// TODO this does not work after closing a dialog window with a Label
		w32.SetWindowText(c.handle, text)
	}
}

func (c *textControl) Font() *Font {
	return c.font
}

func (c *textControl) parentFontChanged() {
	c.SetFont(c.font)
}

func (c *textControl) SetFont(font *Font) {
	c.font = font
	if c.handle != 0 {
		w32.SendMessage(c.handle, w32.WM_SETFONT, uintptr(c.fontHandle()), 1)
	}
}

func (c *textControl) fontHandle() w32.HFONT {
	if c.font != nil {
		return c.font.handle
	}
	if c.parent != nil {
		font := c.parent.Font()
		if font != nil {
			return font.handle
		}
	}
	return 0
}

func (c *textControl) Focus() {
	if c.handle != 0 {
		w32.SetFocus(c.handle)
	}
}
