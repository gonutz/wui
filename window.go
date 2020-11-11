//+build windows

package wui

import (
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/gonutz/w32"
)

var windows windowStack

// TODO windowStack used to have a mutex but the solution was never complete.
// Try to make the system as a whole safe for concurrent access. This is not the
// highest priority. Make it work synchronously first.
type windowStack struct {
	windows []*Window
}

func (s *windowStack) top() *Window {
	if len(s.windows) == 0 {
		return nil
	}
	return s.windows[len(s.windows)-1]
}

func (s *windowStack) push(w *Window) {
	s.windows = append(s.windows, w)
}

func (s *windowStack) pop() {
	if len(s.windows) > 0 {
		s.windows = s.windows[:len(s.windows)-1]
	}
}

type WindowState int

const (
	WindowNormal WindowState = iota
	WindowMaximized
	WindowMinimized
)

func (s WindowState) String() string {
	switch s {
	case WindowNormal:
		return "wui.WindowNormal"
	case WindowMaximized:
		return "wui.WindowMaximized"
	case WindowMinimized:
		return "wui.WindowMinimized"
	default:
		return "unknown WindowState"
	}
}

func (s WindowState) toCmd() int {
	switch s {
	case WindowMaximized:
		return w32.SW_MAXIMIZE
	case WindowMinimized:
		return w32.SW_MINIMIZE
	default:
		return w32.SW_SHOWNORMAL
	}
}

func NewWindow() *Window {
	w := &Window{
		background: ColorButtonFace,
		cursor:     CursorArrow,
		alpha:      255,
	}
	w.SetInnerBounds(100, 50, 600, 400)
	return w
}

type Window struct {
	handle           w32.HWND
	parent           *Window
	hidesBorder      bool
	fixedSize        bool
	hidesMinButton   bool
	hidesMaxButton   bool
	hidesCloseButton bool
	title            string
	x                int
	y                int
	width            int
	height           int
	lastInnerWidth   int
	lastInnerHeight  int
	state            WindowState
	background       Color
	cursor           *Cursor
	menu             *Menu
	menuStrings      []*MenuString
	font             *Font
	controls         []Control
	children         []Control
	icon             *Icon
	showConsole      bool
	altF4disabled    bool
	shortcuts        []shortcut
	accelTable       w32.HACCEL
	lastFocus        w32.HWND
	alpha            uint8
	onShow           func()
	onClose          func()
	onCanClose       func() bool
	onMouseMove      func(x, y int)
	onMouseWheel     func(x, y int, delta float64)
	onMouseDown      func(button MouseButton, x, y int)
	onMouseUp        func(button MouseButton, x, y int)
	onKeyDown        func(key int)
	onKeyUp          func(key int)
	onResize         func()
}

func (w *Window) Children() []Control {
	return w.children
}

type MouseButton int

const (
	MouseButtonLeft MouseButton = iota
	MouseButtonMiddle
	MouseButtonRight
)

func (b MouseButton) String() string {
	switch b {
	case MouseButtonLeft:
		return "wui.MouseButtonLeft"
	case MouseButtonMiddle:
		return "wui.MouseButtonLeft"
	case MouseButtonRight:
		return "wui.MouseButtonRight"
	}
	return "unknown MouseButton"
}

type Control interface {
	Bounds() (x, y, width, height int)
	SetBounds(x, y, width, height int)
	Anchors() (horizontal, vertical Anchor)
	SetHorizontalAnchor(a Anchor)
	SetVerticalAnchor(a Anchor)
	Parent() Container
	Handle() uintptr
	Visible() bool
	Enabled() bool

	setParent(parent Container)
	create(id int)
	parentFontChanged()
	handleNotification(cmd uintptr)
	canFocus() bool
	eatsTabs() bool
	closing()
	destroy()
}

type Container interface {
	Add(Control)
	Children() []Control
	Parent() Container
	Bounds() (x, y, width, height int)
	SetBounds(x, y, width, height int)
	InnerBounds() (x, y, width, height int)
	Font() *Font
	Visible() bool
	Enabled() bool

	setParent(parent Container)
	getHandle() w32.HWND
	getInstance() w32.HINSTANCE
	onWM_COMMAND(w, l uintptr)
	onWM_DRAWITEM(w, l uintptr)
	onWM_NOTIFY(w, l uintptr)
	getIDFor(c Control) int
}

func (*Window) setParent(parent Container) {
	// TODO This is just to implement the container interface.
}

func (w *Window) getHandle() w32.HWND {
	return w.handle
}

func (w *Window) getInstance() w32.HINSTANCE {
	return w32.HINSTANCE(w32.GetWindowLong(w.handle, w32.GWL_HINSTANCE))
}

func (w *Window) Title() string { return w.title }

func (w *Window) SetTitle(title string) {
	w.title = title
	if w.handle != 0 {
		w32.SetWindowText(w.handle, title)
	}
}

func (w *Window) style() uint {
	var s uint

	if w.hidesBorder {
		s |= w32.WS_POPUP
	} else {
		s |= w32.WS_CAPTION | w32.WS_SYSMENU
	}

	if !w.fixedSize {
		s |= w32.WS_SIZEBOX
	}

	if !w.hidesMinButton {
		s |= w32.WS_MINIMIZEBOX
	}

	if !w.hidesMaxButton {
		s |= w32.WS_MAXIMIZEBOX
	}

	return s
}

func (w *Window) extendedStyle() uint {
	if w.alpha != 255 {
		return w32.WS_EX_LAYERED
	}
	return 0
}

func (w *Window) readBounds() {
	r := w32.GetWindowRect(w.handle)
	w.x = int(r.Left)
	w.y = int(r.Top)
	w.width = int(r.Width())
	w.height = int(r.Height())
}

func (w *Window) X() int {
	x, _, _, _ := w.Bounds()
	return x
}

func (w *Window) Y() int {
	_, y, _, _ := w.Bounds()
	return y
}

func (w *Window) Position() (x, y int) {
	x, y, _, _ = w.Bounds()
	return
}

func (w *Window) Width() int {
	_, _, width, _ := w.Bounds()
	return width
}

func (w *Window) Height() int {
	_, _, _, height := w.Bounds()
	return height
}

func (w *Window) Size() (width, height int) {
	_, _, width, height = w.Bounds()
	return
}

func (w *Window) Bounds() (x, y, width, height int) {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.x, w.y, w.width, w.height
}

func (w *Window) SetX(x int) {
	_, y, width, height := w.Bounds()
	w.SetBounds(x, y, width, height)
}

func (w *Window) SetY(y int) {
	x, _, width, height := w.Bounds()
	w.SetBounds(x, y, width, height)
}

func (w *Window) SetPosition(x, y int) {
	_, _, width, height := w.Bounds()
	w.SetBounds(x, y, width, height)
}

func (w *Window) SetWidth(width int) {
	x, y, _, height := w.Bounds()
	w.SetBounds(x, y, width, height)
}

func (w *Window) SetHeight(height int) {
	x, y, width, _ := w.Bounds()
	w.SetBounds(x, y, width, height)
}

func (w *Window) SetSize(width, height int) {
	x, y, _, _ := w.Bounds()
	w.SetBounds(x, y, width, height)
}

func (w *Window) SetBounds(x, y, width, height int) {
	if w.handle != 0 {
		// The window will receive a WM_SIZE which will handle anchoring child
		// controls.
		w32.SetWindowPos(
			w.handle, 0,
			x, y, width, height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	} else {
		oldW, oldH := w.InnerSize()
		w.x = x
		w.y = y
		w.width = width
		w.height = height
		newW, newH := w.InnerSize()
		repositionChidrenByAnchors(w, oldW, oldH, newW, newH)
	}
}

func repositionChidrenByAnchors(c Container, oldW, oldH, newW, newH int) {
	dw := newW - oldW
	dh := newH - oldH
	oldCenterX := oldW / 2
	newCenterX := newW / 2
	oldCenterY := oldH / 2
	newCenterY := newH / 2
	for _, c := range c.Children() {
		x, y, w, h := c.Bounds()
		hAnchor, vAnchor := c.Anchors()

		if hAnchor == AnchorMinAndMax {
			w += dw
		} else if hAnchor == AnchorMax {
			x += dw
		} else if hAnchor == AnchorCenter {
			dx := oldCenterX - x
			x = newCenterX - dx
		} else if hAnchor == AnchorMinAndCenter {
			w += newCenterX - oldCenterX
		} else if hAnchor == AnchorMaxAndCenter {
			x += newCenterX - oldCenterX
			w += dw - (newCenterX - oldCenterX)
		}

		if vAnchor == AnchorMinAndMax {
			h += dh
		} else if vAnchor == AnchorMax {
			y += dh
		} else if vAnchor == AnchorCenter {
			dy := oldCenterY - y
			y = newCenterY - dy
		} else if vAnchor == AnchorMinAndCenter {
			h += newCenterY - oldCenterY
		} else if vAnchor == AnchorMaxAndCenter {
			y += newCenterY - oldCenterY
			h += dh - (newCenterY - oldCenterY)
		}

		c.SetBounds(x, y, w, h)
	}
}

func (w *Window) InnerX() int {
	x, _, _, _ := w.InnerBounds()
	return x
}

func (w *Window) SetInnerX(x int) {
	_, y, width, height := w.InnerBounds()
	w.SetInnerBounds(x, y, width, height)
}

func (w *Window) InnerY() int {
	_, y, _, _ := w.InnerBounds()
	return y
}

func (w *Window) SetInnerY(y int) {
	x, _, width, height := w.InnerBounds()
	w.SetInnerBounds(x, y, width, height)
}

func (w *Window) InnerPosition() (x, y int) {
	x, y, _, _ = w.InnerBounds()
	return
}

func (w *Window) SetInnerPosition(x, y int) {
	_, _, width, height := w.InnerBounds()
	w.SetInnerBounds(x, y, width, height)
}

func (w *Window) InnerWidth() int {
	_, _, width, _ := w.InnerBounds()
	return width
}

func (w *Window) SetInnerWidth(width int) {
	x, y, _, height := w.InnerBounds()
	w.SetInnerBounds(x, y, width, height)
}

func (w *Window) InnerHeight() int {
	_, _, _, height := w.InnerBounds()
	return height
}

func (w *Window) SetInnerHeight(height int) {
	x, y, width, _ := w.InnerBounds()
	w.SetInnerBounds(x, y, width, height)
}

func (w *Window) InnerSize() (width, height int) {
	_, _, width, height = w.InnerBounds()
	return
}

func (w *Window) SetInnerSize(width, height int) {
	x, y, _, _ := w.InnerBounds()
	w.SetInnerBounds(x, y, width, height)
}

func (w *Window) InnerBounds() (x, y, width, height int) {
	if w.handle != 0 {
		x, y = w32.ClientToScreen(w.handle, 0, 0)
		r := w32.GetClientRect(w.handle)
		width = int(r.Width())
		height = int(r.Height())
	} else {
		x, y = w.Position()
		var r w32.RECT
		w32.AdjustWindowRectEx(&r, w.style(), w.menu != nil, w.extendedStyle())
		x -= int(r.Left)
		y -= int(r.Top)
		width = w.width - int(r.Width())
		height = w.height - int(r.Height())
	}
	return
}

func (w *Window) SetInnerBounds(x, y, width, height int) {
	var r w32.RECT
	w32.AdjustWindowRectEx(&r, w.style(), w.menu != nil, w.extendedStyle())
	w.x = x + int(r.Left)
	w.y = y + int(r.Top)
	w.width = width + int(r.Width())
	w.height = height + int(r.Height())
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
}

func (w *Window) SetState(s WindowState) {
	w.state = s
	if w.handle != 0 {
		w32.ShowWindow(w.handle, s.toCmd())
	}
}

func (w *Window) State() WindowState {
	if w.handle != 0 {
		var p w32.WINDOWPLACEMENT
		if w32.GetWindowPlacement(w.handle, &p) {
			switch p.ShowCmd {
			case w32.SW_MAXIMIZE:
				w.state = WindowMaximized
			case w32.SW_SHOWMINIMIZED, w32.SW_MINIMIZE,
				w32.SW_SHOWMINNOACTIVE, w32.SW_FORCEMINIMIZE:
				w.state = WindowMinimized
			default:
				w.state = WindowNormal
			}
		}
	}
	return w.state
}

func (w *Window) GetBackground() Color {
	return w.background
}

func (w *Window) SetBackground(c Color) {
	w.background = c
	if w.handle != 0 {
		brush := w32.CreateSolidBrush(uint32(c))
		w32.SetClassLongPtr(w.handle, w32.GCLP_HBRBACKGROUND, uintptr(brush))
		w32.InvalidateRect(w.handle, nil, true)
	}
}

func (w *Window) Cursor() *Cursor { return w.cursor }

func (w *Window) SetCursor(c *Cursor) {
	w.cursor = c
	if w.handle != 0 {
		w32.SetClassLongPtr(w.handle, w32.GCLP_HCURSOR, uintptr(c.handle))
	}
}

func (w *Window) Menu() *Menu {
	return w.menu
}

func (w *Window) SetMenu(m *Menu) {
	w.menu = m
	if w.handle != 0 {
		// TODO update menu
	}
}

func (w *Window) Font() *Font {
	return w.font
}

func (w *Window) SetFont(f *Font) {
	w.font = f
	// TODO This should probably go over the children and recurse down from
	// there, making sure all parents have their font set first, before their
	// children.
	for _, c := range w.controls {
		c.parentFontChanged()
	}
}

// TODO When Adding a Control to one Container while it already has a parent,
// either panic or remove it from the old parent first. Do this when working on
// Remove.

func (w *Window) Add(c Control) {
	w.children = append(w.children, c)
	c.setParent(w)
	if w.handle != 0 {
		c.create(w.getIDFor(c))
	}
}

func (w *Window) getIDFor(c Control) int {
	for i := range w.controls {
		if c == w.controls[i] {
			return i
		}
	}
	w.controls = append(w.controls, c)
	return len(w.controls) - 1
}

func (w *Window) SetOnShow(f func()) {
	w.onShow = f
}

func (w *Window) SetOnClose(f func()) {
	w.onClose = f
}

// SetOnCanClose is passed a function that is called when the window is about to
// be closed, e.g. when the user hits Alt+F4. If f returns true the window is
// closed, if f returns false, the window stays open.
func (w *Window) SetOnCanClose(f func() bool) {
	w.onCanClose = f
}

func (w *Window) SetOnMouseMove(f func(x, y int)) {
	w.onMouseMove = f
}

func (w *Window) SetOnMouseWheel(f func(x, y int, delta float64)) {
	w.onMouseWheel = f
}

func (w *Window) SetOnMouseDown(f func(button MouseButton, x, y int)) {
	w.onMouseDown = f
}

func (w *Window) SetOnMouseUp(f func(button MouseButton, x, y int)) {
	w.onMouseUp = f
}

func (w *Window) SetOnKeyDown(f func(key int)) {
	w.onKeyDown = f
}

func (w *Window) SetOnKeyUp(f func(key int)) {
	w.onKeyUp = f
}

func (w *Window) OnResize() func() {
	return w.onResize
}

func (w *Window) SetOnResize(f func()) {
	w.onResize = f
}

func (w *Window) Close() {
	if w.handle != 0 {
		w32.SendMessage(w.handle, w32.WM_CLOSE, 0, 0)
	}
}

func (w *Window) interceptMessage(msg *w32.MSG) bool {
	if msg.Message == w32.WM_KEYDOWN && msg.WParam == w32.VK_TAB {
		focus := uintptr(w32.GetFocus())
		cur := func() int {
			for i := range w.controls {
				if w.controls[i].Handle() == focus {
					return i
				}
			}
			return -1
		}()
		if cur != -1 && w.controls[cur].eatsTabs() {
			return false
		}
		shiftDown := w32.GetKeyState(w32.VK_SHIFT)&0x8000 != 0
		nth := func(i int) int {
			return (cur + 1 + i) % len(w.controls)
		}
		if shiftDown {
			nth = func(i int) int {
				return (cur + len(w.controls) - 1 - i) % len(w.controls)
			}
		}
		for i := range w.controls {
			j := nth(i)
			if w.controls[j].canFocus() &&
				Visible(w.controls[j]) &&
				Enabled(w.controls[j]) {
				w32.SetFocus(w32.HWND(w.controls[j].Handle()))
				return true
			}
		}
		return true
	}
	return false
}

// Visible return true if the given control and all of its parents are visible.
// This is different from the control's Visible() function which only returns
// the visibility of the control itself. If it is contained in an invisible
// parent, c.Visible() will return true while Visible(c) will return false.
func Visible(c VisibleControl) bool {
	return c == nil || c.Visible() && Visible(c.Parent())
}

type VisibleControl interface {
	Visible() bool
	Parent() Container
}

// Enabled return true if the given control and all of its parents are enabled.
// This is different from the control's Enabled() function which only returns
// the enabled state of the control itself. If it is contained in a disabled
// parent, c.Enabled() will return true while Enabled(c) will return false.
func Enabled(c EnabledControl) bool {
	return c == nil || c.Enabled() && Enabled(c.Parent())
}

type EnabledControl interface {
	Enabled() bool
	Parent() Container
}

func (w *Window) onMsg(window w32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	mouseX := int(lParam & 0xFFFF)
	mouseY := int(lParam&0xFFFF0000) >> 16
	switch msg {
	case w32.WM_MOUSEMOVE:
		if w.onMouseMove != nil {
			w.onMouseMove(mouseX, mouseY)
			return 0
		}
	case w32.WM_MOUSEWHEEL:
		if w.onMouseWheel != nil {
			delta := float64(int16((wParam&0xFFFF0000)>>16)) / 120
			w.onMouseWheel(mouseX, mouseY, delta)
		}
		return 0
	case w32.WM_LBUTTONDOWN, w32.WM_MBUTTONDOWN, w32.WM_RBUTTONDOWN:
		if w.onMouseDown != nil {
			b := MouseButtonLeft
			if msg == w32.WM_MBUTTONDOWN {
				b = MouseButtonMiddle
			}
			if msg == w32.WM_RBUTTONDOWN {
				b = MouseButtonRight
			}
			w.onMouseDown(b, mouseX, mouseY)
		}
		return 0
	case w32.WM_LBUTTONUP, w32.WM_MBUTTONUP, w32.WM_RBUTTONUP:
		if w.onMouseUp != nil {
			b := MouseButtonLeft
			if msg == w32.WM_MBUTTONUP {
				b = MouseButtonMiddle
			}
			if msg == w32.WM_RBUTTONUP {
				b = MouseButtonRight
			}
			w.onMouseUp(b, mouseX, mouseY)
		}
		return 0
	case w32.WM_DRAWITEM:
		w.onWM_DRAWITEM(wParam, lParam)
		return 0
	case w32.WM_KEYDOWN:
		if w.onKeyDown != nil {
			w.onKeyDown(int(wParam))
			return 0
		}
	case w32.WM_KEYUP:
		if w.onKeyUp != nil {
			w.onKeyUp(int(wParam))
			return 0
		}
	case w32.WM_COMMAND:
		w.onWM_COMMAND(wParam, lParam)
		return 0
	case w32.WM_SYSCOMMAND:
		if w.altF4disabled && wParam == w32.SC_CLOSE && (lParam>>16) <= 0 {
			return 0
		}
		return w32.DefWindowProc(window, msg, wParam, lParam)
	case w32.WM_NOTIFY:
		w.onWM_NOTIFY(wParam, lParam)
		return 0
	case w32.WM_SIZE:
		oldW, oldH := w.lastInnerWidth, w.lastInnerHeight
		newW, newH := w.InnerSize()
		repositionChidrenByAnchors(w, oldW, oldH, newW, newH)
		w.lastInnerWidth, w.lastInnerHeight = newW, newH
		if w.onResize != nil {
			w.onResize()
		}
		w32.InvalidateRect(window, nil, true)
		switch wParam {
		case w32.SIZE_MAXIMIZED:
			w.state = WindowMaximized
		case w32.SIZE_MINIMIZED:
			w.state = WindowMinimized
		case w32.SIZE_RESTORED:
			w.state = WindowNormal
		}
		return 0
	case w32.WM_ACTIVATE:
		active := wParam != 0
		if active {
			if w.lastFocus != 0 {
				w32.SetFocus(w.lastFocus)
			}
		} else {
			w.lastFocus = w32.GetFocus()
		}
		return 0
	case w32.WM_HSCROLL, w32.WM_VSCROLL:
		for _, c := range w.controls {
			if lParam == c.Handle() {
				if s, ok := c.(*Slider); ok {
					s.handleChange(wParam & 0xFFFF)
				}
			}
		}
	case w32.WM_DESTROY:
		w32.PostQuitMessage(0)
		return 0
	case w32.WM_CLOSE:
		if w.onCanClose != nil {
			if w.onCanClose() == false {
				return 0
			}
		}
		if w.parent != nil {
			w32.EnableWindow(w.parent.handle, true)
			w32.SetForegroundWindow(w.parent.handle)
		}
		if w.onClose != nil {
			w.onClose()
		}
		w.closing()
	}
	return w32.DefWindowProc(window, msg, wParam, lParam)
}

func (w *Window) onWM_DRAWITEM(wParam, lParam uintptr) {
	index := wParam
	if 0 <= index && index < uintptr(len(w.controls)) {
		if p, ok := w.controls[index].(*PaintBox); ok {
			if p.onPaint != nil {
				drawItem := ((*w32.DRAWITEMSTRUCT)(unsafe.Pointer(lParam)))
				// create a back buffer
				p.backBuffer.setMinSize(drawItem.HDC, p.width, p.height)
				bmpOld := w32.SelectObject(
					p.backBuffer.dc,
					w32.HGDIOBJ(p.backBuffer.bmp),
				)
				c := &Canvas{
					hdc:    p.backBuffer.dc,
					width:  p.width,
					height: p.height,
				}
				if p.parent != nil {
					c.SetFont(p.parent.Font())
				}
				c.ClearDrawRegions()
				p.onPaint(c)

				// blit the backbuffer to the front
				w32.BitBlt(
					drawItem.HDC, 0, 0, p.width, p.height,
					p.backBuffer.dc, 0, 0, w32.SRCCOPY,
				)
				w32.SelectObject(p.backBuffer.dc, bmpOld)
			}
		}
	}
}

const className = "wui_window_class"

func (w *Window) Show() error {
	if w.handle != 0 {
		return errors.New("wui.Window.Show: window already visible")
	}

	if windows.top() != nil {
		return errors.New("wui.Window.Show: another window is already visible")
	}
	windows.push(w)
	defer windows.pop()

	runtime.LockOSThread()
	if !w.showConsole {
		hideConsoleWindow()
	}
	setManifest()

	class := w32.WNDCLASSEX{
		Background: w32.CreateSolidBrush(uint32(w.background)),
		WndProc:    syscall.NewCallback(w.onMsg),
		Cursor:     w.cursor.handle,
		ClassName:  syscall.StringToUTF16Ptr(className),
	}
	atom := w32.RegisterClassEx(&class)
	if atom == 0 {
		return errors.New("wui.Window.Show: RegisterClassEx failed")
	}
	defer w32.UnregisterClassAtom(atom, w32.GetModuleHandle(""))

	// We remember the desired state, the window setup will make a WM_SIZE
	// message with a restored window state arrive before we call ShowWindow.
	state := w.state

	window := w32.CreateWindowEx(
		w.extendedStyle(),
		syscall.StringToUTF16Ptr(className),
		syscall.StringToUTF16Ptr(w.title),
		w.style(),
		w.x, w.y, w.width, w.height,
		0, 0, 0, nil,
	)
	if window == 0 {
		return errors.New("wui.Window.Show: CreateWindowEx failed")
	}
	w.handle = window
	if w.alpha != 255 {
		w32.SetLayeredWindowAttributes(w.handle, 0, w.alpha, w32.LWA_ALPHA)
	}
	if w.hidesCloseButton {
		w32.EnableMenuItem(
			w32.GetSystemMenu(w.handle, false),
			w32.SC_CLOSE,
			w32.MF_BYCOMMAND|w32.MF_DISABLED|w32.MF_GRAYED,
		)
	}

	w.updateAccelerators()
	w.lastInnerWidth, w.lastInnerHeight = w.InnerSize()
	w.createContents()
	w.applyIcon()
	w32.ShowWindow(window, state.toCmd())
	w.readBounds()
	if w.onShow != nil {
		w.onShow()
	}

	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) != 0 {
		if !w.interceptMessage(&msg) {
			if w.accelTable == 0 || !w32.TranslateAccelerator(w.handle, w.accelTable, &msg) {
				w32.TranslateMessage(&msg)
				w32.DispatchMessage(&msg)
			}
		}
	}

	w.destroy()

	return nil
}

func (w *Window) createContents() {
	if w.menu != nil {
		var addItems func(m w32.HMENU, items []MenuItem)
		addItems = func(m w32.HMENU, items []MenuItem) {
			for _, item := range items {
				switch menuItem := item.(type) {
				case *Menu:
					menu := w32.CreateMenu()
					w32.AppendMenu(m, w32.MF_POPUP, uintptr(menu), menuItem.name)
					addItems(menu, menuItem.items)
				case *MenuString:
					id := uintptr(len(w.menuStrings))
					w32.AppendMenu(
						m,
						w32.MF_STRING,
						id,
						menuItem.text,
					)
					menuItem.window = w.handle
					menuItem.menu = m
					menuItem.id = uint(id)
					w.menuStrings = append(w.menuStrings, menuItem)
				case menuSeparator:
					w32.AppendMenu(m, w32.MF_SEPARATOR, 0, "")
				}
			}
		}
		menuBar := w32.CreateMenu()
		addItems(menuBar, w.menu.items)
		w32.SetMenu(w.handle, menuBar)
		for _, m := range w.menuStrings {
			if m.Checked() {
				m.SetChecked(true)
			}
		}
	}

	for _, c := range w.children {
		c.create(w.getIDFor(c))
	}
}

func (w *Window) closing() {
	for _, c := range w.children {
		c.closing()
	}
}

func (w *Window) destroy() {
	if w.handle != 0 {
		for _, c := range w.children {
			c.destroy()
		}
		w32.DestroyWindow(w.handle)
		w.handle = 0
	}
}

func (w *Window) onWM_COMMAND(wParam, lParam uintptr) {
	wHi := (wParam & 0xFFFF0000) >> 16
	wLo := wParam & 0xFFFF
	if lParam == 0 && wHi == 0 {
		// low word of w contains menu ID
		id := int(wLo)
		if 0 <= id && id < len(w.menuStrings) {
			f := w.menuStrings[id].onClick
			if f != nil {
				f()
			}
		}
	} else if lParam == 0 && wHi == 1 {
		// low word of w contains accelerator ID
		index := int(wLo)
		if 0 <= index && index < len(w.shortcuts) {
			w.shortcuts[index].f()
		}
	} else if lParam != 0 {
		// control clicked
		index := wParam & 0xFFFF
		cmd := (wParam & 0xFFFF0000) >> 16
		if index < uintptr(len(w.controls)) {
			w.controls[index].handleNotification(cmd)
		}
	}
}

func (w *Window) onWM_NOTIFY(wParam, lParam uintptr) {
	header := *((*w32.NMHDR)(unsafe.Pointer(lParam)))
	if header.Code == uint32(w32.UDN_DELTAPOS) {
		i := int(wParam)
		if 0 <= i && i < len(w.controls) {
			if f, ok := w.controls[i].(*FloatUpDown); ok {
				updown := *((*w32.NMUPDOWN)(unsafe.Pointer(lParam)))
				f.SetValue(f.value - float64(updown.Delta))
			}
		}
	} else if header.Code == w32.LVN_ITEMCHANGED&0xFFFFFFFF {
		i := int(wParam)
		if 0 <= i && i < len(w.controls) {
			if t, ok := w.controls[i].(*StringTable); ok {
				change := *((*w32.NMLISTVIEW)(unsafe.Pointer(lParam)))
				if change.UChanged == w32.LVIF_STATE {
					if change.UNewState&(w32.LVIS_FOCUSED|w32.LVIS_SELECTED) != 0 {
						t.newItemSelected(int(change.IItem))
					} else {
						t.itemDeselected()
					}
				}
			}
		}
	}
}

// hideConsoleWindow hides the associated console window that gets created for
// Windows applications that are of type console instead of type GUI. When
// building you can pass the ldflag H=windowsgui to suppress this but if you
// just go build or go run, a console window will pop open along with the GUI
// window. hideConsoleWindow hides it.
func hideConsoleWindow() {
	console := w32.GetConsoleWindow()
	if console == 0 {
		return // No console attached.
	}
	// If this application is the process that created the console window, then
	// this program was not compiled with the -H=windowsgui flag and on start-up
	// it created a console along with the main application window. In this case
	// hide the console window. See
	// http://stackoverflow.com/questions/9009333/how-to-check-if-the-program-is-run-from-a-console
	_, consoleProcID := w32.GetWindowThreadProcessId(console)
	if w32.GetCurrentProcessId() == consoleProcID {
		w32.ShowWindowAsync(console, w32.SW_HIDE)
	}
}

func setManifest() {
	manifest := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
    <dependency>
        <dependentAssembly>
            <assemblyIdentity
				type="win32"
				processorArchitecture="*"
				language="*"
				name="Microsoft.Windows.Common-Controls"
				version="6.0.0.0"
				publicKeyToken="6595b64144ccf1df"
			/>
        </dependentAssembly>
    </dependency>
</assembly>`
	// create a temporary manifest file, load it, then delete it
	f, err := ioutil.TempFile("", "manifest_")
	if err != nil {
		return
	}
	manifestPath := f.Name()
	defer os.Remove(manifestPath)
	f.WriteString(manifest)
	f.Close()
	ctx := w32.CreateActCtx(&w32.ACTCTX{
		Source: syscall.StringToUTF16Ptr(manifestPath),
	})
	w32.ActivateActCtx(ctx)
}

func (w *Window) applyIcon() {
	if w.handle != 0 {
		var h uintptr
		if w.icon != nil {
			h = uintptr(w.icon.handle)
		}
		w32.SendMessage(w.handle, w32.WM_SETICON, w32.ICON_SMALL, h)
		w32.SendMessage(w.handle, w32.WM_SETICON, w32.ICON_SMALL2, h)
		w32.SendMessage(w.handle, w32.WM_SETICON, w32.ICON_BIG, h)
	}
}

func (w *Window) Icon() *Icon {
	return w.icon
}

func (w *Window) SetIcon(icon *Icon) {
	w.icon = icon
	w.applyIcon()
}

// TODO Show and ShowModal are very similar and need to stay in sync, refactor
// the two functions for easier maintenance.

func (w *Window) ShowModal() error {
	if w.handle != 0 {
		return errors.New("wui.Window.ShowModal: window already visible")
	}

	w.parent = windows.top()
	if w.parent == nil {
		return w.Show()
	}
	windows.push(w)
	defer windows.pop()

	if w.icon == nil {
		w.icon = w.parent.icon
	}
	if w.font == nil {
		w.font = w.parent.font
	}

	// We remember the desired state, the window setup will make a WM_SIZE
	// message with a restored window state arrive before we call ShowWindow.
	state := w.state

	window := w32.CreateWindowEx(
		w.extendedStyle(),
		syscall.StringToUTF16Ptr(className),
		syscall.StringToUTF16Ptr(w.title),
		w.style(),
		w.x, w.y, w.width, w.height,
		w.parent.handle,
		0, 0, nil,
	)
	if window == 0 {
		return errors.New("wui.Window.ShowModal: CreateWindowEx failed")
	}
	w.handle = window
	if w.alpha != 255 {
		w32.SetLayeredWindowAttributes(w.handle, 0, w.alpha, w32.LWA_ALPHA)
	}
	if w.hidesCloseButton {
		w32.EnableMenuItem(
			w32.GetSystemMenu(w.handle, false),
			w32.SC_CLOSE,
			w32.MF_BYCOMMAND|w32.MF_DISABLED|w32.MF_GRAYED,
		)
	}

	w32.SetWindowSubclass(w.handle, syscall.NewCallback(func(
		window w32.HWND,
		msg uint32,
		wParam, lParam uintptr,
		subclassID uintptr,
		refData uintptr,
	) uintptr {
		return w.onMsg(window, msg, wParam, lParam)
	}), 0, 0)

	w.updateAccelerators()
	w.lastInnerWidth, w.lastInnerHeight = w.InnerSize()
	w.createContents()
	w.applyIcon()
	w32.ShowWindow(window, state.toCmd())
	w32.EnableWindow(w.parent.handle, false)
	w.readBounds()
	if w.onShow != nil {
		w.onShow()
	}

	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) != 0 {
		if !w.interceptMessage(&msg) {
			if w.accelTable == 0 || !w32.TranslateAccelerator(w.handle, w.accelTable, &msg) {
				w32.TranslateMessage(&msg)
				w32.DispatchMessage(&msg)
			}
		}
	}
	return nil
}

func (w *Window) ShowConsoleOnStart() {
	w.showConsole = true
}

func (w *Window) HideConsoleOnStart() {
	w.showConsole = false
}

// TODO Remove DisableAltF4. Do we really need it? If we do, expose this
// behavior in a more general way, e.g. as a close reason in OnCanClose().

func (w *Window) DisableAltF4() {
	w.altF4disabled = true
}

func (w *Window) EnableAltF4() {
	w.altF4disabled = false
}

// TODO Destroy functions for all the things, also a Remove from containers.

func (w *Window) Destroy() {
	if w.handle != 0 {
		w32.DestroyWindow(w.handle)
	}
}

func (w *Window) Alpha() uint8 {
	return w.alpha
}

func (w *Window) SetAlpha(a uint8) {
	w.alpha = a
	if w.handle != 0 {
		style := w32.GetWindowLong(w.handle, w32.GWL_EXSTYLE)
		if w.alpha != 255 {
			if style&w32.WS_EX_LAYERED == 0 {
				w32.SetWindowLong(
					w.handle,
					w32.GWL_EXSTYLE,
					style|w32.WS_EX_LAYERED,
				)
			}
			w32.SetLayeredWindowAttributes(
				w.handle,
				0,
				w.alpha,
				w32.LWA_ALPHA,
			)
		} else {
			w32.SetWindowLong(
				w.handle,
				w32.GWL_EXSTYLE,
				style & ^w32.WS_EX_LAYERED,
			)
			w32.RedrawWindow(
				w.handle,
				nil,
				0,
				w32.RDW_ERASE|w32.RDW_INVALIDATE|w32.RDW_FRAME|w32.RDW_ALLCHILDREN,
			)
		}
	}
}

type shortcut struct {
	// accel has its Cmd set to 0 for comparibility. It will be copied and Cmd
	// assigned on the copies when creating the accelerator table.
	accel w32.ACCEL
	f     func()
}

func toACCEL(keys []Key) w32.ACCEL {
	var a w32.ACCEL
	a.Virt = w32.FVIRTKEY
	for _, key := range keys {
		switch key {
		case KeyControl, KeyLeftControl, KeyRightControl:
			a.Virt |= w32.FCONTROL
		case KeyShift, KeyLeftShift, KeyRightShift:
			a.Virt |= w32.FSHIFT
		case KeyAlt, KeyLeftAlt, KeyRightAlt:
			a.Virt |= w32.FALT
		default:
			a.Key = uint16(key)
		}
	}
	return a
}

func (w *Window) SetShortcut(f func(), keys ...Key) {
	if len(keys) == 0 {
		return
	}
	if w.accelTable != 0 {
		defer w.updateAccelerators()
	}
	s := shortcut{accel: toACCEL(keys), f: f}
	// Look for an existing shortcut for this key combination and replace it if
	// we find it.
	for i := range w.shortcuts {
		if w.shortcuts[i].accel == s.accel {
			w.shortcuts[i].f = f // Replace the handler function.
			if f == nil {
				// Setting nil deletes the shortcut.
				w.shortcuts = append(w.shortcuts[:i], w.shortcuts[i+1:]...)
			}
			return
		}
	}
	// If we have not returned until this point, the shortcut is new, so create
	// it.
	if s.f != nil {
		w.shortcuts = append(w.shortcuts, s)
	}
}

func (w *Window) updateAccelerators() {
	if w.accelTable != 0 {
		w32.DestroyAcceleratorTable(w.accelTable)
		w.accelTable = 0
	}
	if len(w.shortcuts) > 0 {
		// NOTE There is a bug somewhere in our system: when there is only one
		// accelerator in the array, CreateAcceleratorTable will return an
		// invalid 0 handle. Thus a single shortcut will never be triggered. As
		// soon as there is another accelerator, it works. This is why we add +1
		// to the length here. The last accelerator is never used, its index is
		// outside w.shortcuts and thus ignored.
		accels := make([]w32.ACCEL, len(w.shortcuts)+1)
		for i := range w.shortcuts {
			accels[i] = w.shortcuts[i].accel
			accels[i].Cmd = uint16(i)
		}
		w.accelTable = w32.CreateAcceleratorTable(accels)
	}
}

// TODO Have good scrollbars. Do we still want to have Scroll? For other
// containers as well?

func (w *Window) Scroll(dx, dy int) {
	if w.handle != 0 {
		w32.ScrollWindow(w.handle, dx, dy, nil, nil)
	}
}

// TODO When is Repaint needed?

func (w *Window) Repaint() {
	if w.handle != 0 {
		w32.InvalidateRect(w.handle, nil, true)
	}
}

// Monitor returns the handle to the monitor (HMONITOR) that the window is over.
// Before the window is shown, Monitor returns 0.
func (w *Window) Monitor() uintptr {
	if w.handle == 0 {
		return 0
	}
	return uintptr(w32.MonitorFromWindow(w.handle, w32.MONITOR_DEFAULTTONULL))
}

func (w *Window) Parent() Container {
	if w.parent == nil {
		// This special case is necessary to make == nil comparisons work.
		return nil
	}
	return w.parent
}

func (w *Window) Handle() uintptr {
	return uintptr(w.handle)
}

func (w *Window) HasMinButton() bool {
	return !w.hidesMinButton
}

func (w *Window) SetHasMinButton(hasMin bool) {
	w.hidesMinButton = !hasMin
	if w.handle != 0 {
		style := w32.GetWindowLong(w.handle, w32.GWL_STYLE)

		if hasMin {
			style = style | w32.WS_MINIMIZEBOX
		} else {
			style = style & ^w32.WS_MINIMIZEBOX
		}

		w32.SetWindowLong(w.handle, w32.GWL_STYLE, style)
		w32.SetWindowPos(
			w.handle,
			0, 0, 0, 0, 0,
			w32.SWP_FRAMECHANGED|w32.SWP_NOMOVE|w32.SWP_NOZORDER|
				w32.SWP_NOSIZE|w32.SWP_NOACTIVATE,
		)
	}
}

func (w *Window) HasMaxButton() bool {
	return !w.hidesMaxButton
}

func (w *Window) SetHasMaxButton(hasMax bool) {
	w.hidesMaxButton = !hasMax
	if w.handle != 0 {
		style := w32.GetWindowLong(w.handle, w32.GWL_STYLE)

		if hasMax {
			style = style | w32.WS_MAXIMIZEBOX
		} else {
			style = style & ^w32.WS_MAXIMIZEBOX
		}

		w32.SetWindowLong(w.handle, w32.GWL_STYLE, style)
		w32.SetWindowPos(
			w.handle,
			0, 0, 0, 0, 0,
			w32.SWP_FRAMECHANGED|w32.SWP_NOMOVE|w32.SWP_NOZORDER|
				w32.SWP_NOSIZE|w32.SWP_NOACTIVATE,
		)
	}
}

func (w *Window) HasCloseButton() bool {
	return !w.hidesCloseButton
}

func (w *Window) SetHasCloseButton(hasClose bool) {
	w.hidesCloseButton = !hasClose
	if w.handle != 0 {
		var state uint = w32.MF_BYCOMMAND | w32.MF_DISABLED | w32.MF_GRAYED
		if hasClose {
			state = w32.MF_BYCOMMAND | w32.MF_ENABLED
		}
		w32.EnableMenuItem(w32.GetSystemMenu(w.handle, false), w32.SC_CLOSE, state)
	}
}

func (w *Window) Resizable() bool {
	return !w.fixedSize
}

func (w *Window) SetResizable(canResize bool) {
	if canResize != w.Resizable() {
		w.changeStyles(func() { w.fixedSize = !canResize })
	}
}

func (w *Window) HasBorder() bool {
	return !w.hidesBorder
}

func (w *Window) SetHasBorder(border bool) {
	if border != w.HasBorder() {
		w.changeStyles(func() { w.hidesBorder = !border })
	}
}

func (w *Window) changeStyles(change func()) {
	if w.handle == 0 {
		x, y, width, height := w.InnerBounds()
		change()
		w.SetInnerBounds(x, y, width, height)
	} else {
		// We want to keep the inner bounds the same as they were, this way no
		// controls change their position. First we restore the window, if it is
		// minimized or maximized the inner size would not be right. We restore
		// the original window state at the end. A problematic example: a
		// maximized window without border is set to have a border now. The
		// inner size fills the whole screen while maximized. Adding a border
		// while keeping the inner size constant would add borders which lie
		// outside the screen. This is why we need to restore the window first.
		state := w.state
		w.SetState(WindowNormal)
		// Remember inner bounds.
		x, y, width, height := w.InnerBounds()
		// Update to the new window styles.
		change()
		w32.SetWindowLong(w.handle, w32.GWL_STYLE, int32(w.style()))
		w32.SetWindowLong(w.handle, w32.GWL_EXSTYLE, int32(w.extendedStyle()))
		// Restore the original inner bounds.
		w.SetInnerBounds(x, y, width, height)
		// SetWindowPos and ShowWindow are necessary to make Windows realize the
		// changes in styles.
		w32.SetWindowPos(
			w.handle, 0, 0, 0, 0, 0,
			w32.SWP_FRAMECHANGED|w32.SWP_NOMOVE|w32.SWP_NOZORDER|
				w32.SWP_NOSIZE|w32.SWP_NOACTIVATE,
		)
		w32.ShowWindow(w.handle, w32.SW_SHOWNORMAL)
		// Restore the original window state.
		w.SetState(state)
	}
}

func (w *Window) Visible() bool {
	return true
}

func (w *Window) Enabled() bool {
	return true
}
