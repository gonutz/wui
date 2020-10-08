//+build windows

package wui

import (
	"syscall"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/gonutz/w32"
)

// Anchor defines how a child control is resized when its parent changes size.
// Children can be anchored to min (left or top), center and max (right or
// bottom) of their parents.
type Anchor int

const (
	// AnchorMin makes a control stick to the left/top side of its parent. The
	// control's size is not changed.
	AnchorMin Anchor = iota

	// AnchorMax makes a control stick to the right/bottom side of its parent.
	// The control's size is not changed.
	AnchorMax

	// AnchorCenter keeps a control fixed at the horizontal/vertical center of
	// its parent, e.g. if it is placed 10 pixels left of the center, it will
	// stay 10 pixels left of its parent's center. Its size is not changed.
	AnchorCenter

	// AnchorMinAndMax makes a control's borders stick to its parent's borders.
	// The size changes propertionally to its parent's size, keeping the
	// original distances to its parents borders.
	AnchorMinAndMax

	// AnchorMinAndCenter makes the left/top side of a control stick to its
	// parent's left/top side while the right/bottom side sticks to the parent's
	// center.
	AnchorMinAndCenter

	// AnchorMaxAndCenter makes the right/bottom side of a control stick to its
	// parent's right/bottom side while the left/top side sticks to the parent's
	// center.
	AnchorMaxAndCenter
)

type control struct {
	handle   w32.HWND
	x        int
	y        int
	width    int
	height   int
	hAnchor  Anchor
	vAnchor  Anchor
	parent   container
	disabled bool
	hidden   bool
	onResize func()
}

func (c *control) Parent() Container {
	return c.parent
}

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

func (c *control) SetOnResize(f func()) {
	c.onResize = f
}

func (c *control) OnResize() func() {
	return c.onResize
}

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
	resize := false
	if c.width != width || c.height != height {
		resize = true
	}
	c.x, c.y, c.width, c.height = x, y, width, height
	if c.handle != 0 {
		w32.SetWindowPos(
			c.handle, 0,
			c.x, c.y, c.width, c.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	if resize && c.onResize != nil {
		c.onResize()
	}
}

func (c *control) Anchors() (horizontal, vertical Anchor) {
	return c.hAnchor, c.vAnchor
}

func (c *control) SetHorizontalAnchor(a Anchor) {
	c.hAnchor = a
}

func (c *control) SetVerticalAnchor(a Anchor) {
	c.vAnchor = a
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

func (c *control) handleNotification(cmd uintptr) {}

type textControl struct {
	control
	text        string
	font        *Font
	cursorStart int
	cursorEnd   int
}

func (c *textControl) create(id int, exStyle uint, className string, style uint) {
	c.control.create(id, exStyle, className, style)
	w32.SetWindowText(c.handle, c.text)
	c.SetFont(c.font)
	if c.cursorStart != 0 || c.cursorEnd != 0 {
		c.setCursor(c.cursorStart, c.cursorEnd)
	}
	w32.SetWindowSubclass(c.handle, syscall.NewCallback(func(
		window w32.HWND,
		msg uint32,
		wParam, lParam uintptr,
		subclassID uintptr,
		refData uintptr,
	) uintptr {
		switch msg {
		case w32.WM_CHAR:
			if wParam == 1 && w32.GetKeyState(w32.VK_CONTROL)&0x8000 != 0 {
				// Ctrl+A was pressed - select all text.
				c.SelectAll()
				return 0
			}
			if wParam == 127 {
				// Ctrl+Backspace was pressed, if there is currently a selected
				// active, delete it. If there is just the cursor, delete the
				// last word before the cursor.
				text := []rune(c.Text())
				start, end := c.CursorPosition()
				if start != end {
					// There is a selection, delete it.
					c.SetText(string(append(text[:start], text[end:]...)))
					c.SetCursorPosition(start)
				} else {
					// No selection, delete the last word before the cursor.
					newText, newCursor := deleteWordBeforeCursor(text, start)
					c.SetText(newText)
					c.SetCursorPosition(newCursor)
				}
				return 0
			}
			return w32.DefSubclassProc(window, msg, wParam, lParam)
		default:
			return w32.DefSubclassProc(window, msg, wParam, lParam)
		}
	}), 0, 0)
}

func deleteWordBeforeCursor(text []rune, cursor int) (newText string, newCursor int) {
	prefix := text[:cursor]
	n := len(prefix)

	if n >= 2 && prefix[n-2] == '\r' && prefix[n-1] == '\n' {
		prefix = prefix[:n-2]
	} else if n <= 1 {
		prefix = nil
	} else {
		if unicode.IsSpace(prefix[n-1]) {
			for n > 0 && unicode.IsSpace(prefix[n-1]) {
				prefix = prefix[:n-1]
				n--
			}
		}
		for n > 0 && !unicode.IsSpace(prefix[n-1]) {
			prefix = prefix[:n-1]
			n--
		}
	}

	newText = string(append(prefix, text[cursor:]...))
	newCursor = len(prefix)
	return
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
	// TODO Allow this before showing a window.
	if c.handle != 0 {
		w32.SetFocus(c.handle)
	}
}

func (c *textControl) HasFocus() bool {
	return c.handle != 0 && w32.GetFocus() == c.handle
}

// CursorPosition returns the current cursor position, respectively the current
// selection.
//
// If no selection is active, the returned start and end values are the same.
// They then indicate the index of the UTF-8 character in this EditLine's Text()
// before which the caret is currently set.
//
// If a selection is active, start is the index of the first selected UTF-8
// character in Text() and end is the position one character after the end of
// the selection. The selected text is thus
//
//     c.Text()[start:end]
func (c *textControl) CursorPosition() (start, end int) {
	if c.handle != 0 {
		var start, end uint32
		w32.SendMessage(
			c.handle,
			w32.EM_GETSEL,
			uintptr(unsafe.Pointer(&start)),
			uintptr(unsafe.Pointer(&end)),
		)
		c.cursorStart, c.cursorEnd = int(start), int(end)
	}
	return c.cursorStart, c.cursorEnd
}

func (c *textControl) SetCursorPosition(pos int) {
	c.setCursor(pos, pos)
}

func (c *textControl) SetSelection(start, end int) {
	c.setCursor(start, end)
}

func (c *textControl) SelectAll() {
	c.setCursor(0, -1)
}

func (c *textControl) setCursor(start, end int) {
	c.cursorStart = start
	c.cursorEnd = end

	if c.handle != 0 {
		w32.SendMessage(
			c.handle,
			w32.EM_SETSEL,
			uintptr(uint32(c.cursorStart)),
			uintptr(uint32(c.cursorEnd)),
		)
		w32.SendMessage(c.handle, w32.EM_SCROLLCARET, 0, 0)
	} else {
		c.clampCursorToText()
	}
}

func (c *textControl) clampCursorToText() {
	// If called before we have a window, we have to handle clamping of the
	// positions ourselves.
	n := utf8.RuneCountInString(c.Text())

	if c.cursorEnd == -1 {
		c.cursorEnd = n
	}

	if c.cursorStart < 0 {
		c.cursorStart = 0
	}
	if c.cursorStart > n {
		c.cursorStart = n
	}

	if c.cursorEnd < 0 {
		c.cursorEnd = 0
	}
	if c.cursorEnd > n {
		c.cursorEnd = n
	}

	if c.cursorEnd < c.cursorStart {
		c.cursorStart, c.cursorEnd = c.cursorEnd, c.cursorStart
	}
}
