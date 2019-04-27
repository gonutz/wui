//+build windows

package wui

import (
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"syscall"
	"unicode"
	"unicode/utf16"
	"unsafe"

	"github.com/gonutz/w32"
)

var windows windowStack

type windowStack struct {
	windows []*Window
	mu      sync.Mutex
}

func (s *windowStack) top() *Window {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.windows) == 0 {
		return nil
	}
	return s.windows[len(s.windows)-1]
}

func (s *windowStack) push(w *Window) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows = append(s.windows, w)
}

func (s *windowStack) pop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.windows) > 0 {
		s.windows = s.windows[:len(s.windows)-1]
	}
}

func NewWindow() *Window {
	return &Window{
		className:  "wui_window_class",
		x:          w32.CW_USEDEFAULT,
		y:          w32.CW_USEDEFAULT,
		width:      w32.CW_USEDEFAULT,
		height:     w32.CW_USEDEFAULT,
		style:      w32.WS_OVERLAPPEDWINDOW,
		state:      w32.SW_SHOWNORMAL,
		background: w32.GetSysColorBrush(w32.COLOR_BTNFACE),
		cursor:     w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_ARROW)),
	}
}

func NewDialogWindow() *Window {
	return &Window{
		className:  "wui_window_class",
		x:          w32.CW_USEDEFAULT,
		y:          w32.CW_USEDEFAULT,
		width:      w32.CW_USEDEFAULT,
		height:     w32.CW_USEDEFAULT,
		style:      w32.WS_OVERLAPPED | w32.WS_CAPTION | w32.WS_SYSMENU,
		state:      w32.SW_SHOWNORMAL,
		background: w32.GetSysColorBrush(w32.COLOR_BTNFACE),
		cursor:     w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_ARROW)),
	}
}

type Window struct {
	handle        w32.HWND
	parent        *Window
	className     string
	classStyle    uint32
	title         string
	style         uint
	exStyle       uint
	x             int
	y             int
	width         int
	height        int
	state         int
	background    w32.HBRUSH
	cursor        w32.HCURSOR
	menu          *Menu
	menuStrings   []*MenuString
	font          *Font
	controls      []Control
	icon          uintptr
	showConsole   bool
	altF4disabled bool
	shortcuts     []shortcut
	accelTable    w32.HACCEL
	onShow        func()
	onClose       func()
	onCanClose    func() bool
	onMouseMove   func(x, y int)
	onMouseWheel  func(x, y int, delta float64)
	onKeyDown     func(key int)
	onKeyUp       func(key int)
	onResize      func()
}

type Control interface {
	isControl()
	setParent(parent container)
	create(id int)
	parentFontChanged()
}

type Container interface {
	Add(Control)
}

type container interface {
	isContainer()
	setParent(parent container)
	getHandle() w32.HWND
	getInstance() w32.HINSTANCE
	Font() *Font
	registerControl(c Control)
	onWM_COMMAND(w, l uintptr)
	onWM_DRAWITEM(w, l uintptr)
	onWM_NOTIFY(w, l uintptr)
	controlCount() int
}

func (*Window) isContainer() {}

func (*Window) setParent(parent container) {}

func (w *Window) getHandle() w32.HWND {
	return w.handle
}

func (w *Window) getInstance() w32.HINSTANCE {
	return w32.HINSTANCE(w32.GetWindowLong(w.handle, w32.GWL_HINSTANCE))
}

func (w *Window) ClassName() string { return w.className }

func (w *Window) SetClassName(name string) {
	if w.handle != 0 {
		w.className = name
	}
}

func (w *Window) ClassStyle() uint32 { return w.classStyle }

func (w *Window) SetClassStyle(style uint32) {
	w.classStyle = style
	if w.handle != 0 {
		w32.SetClassLongPtr(w.handle, w32.GCL_STYLE, uintptr(w.classStyle))
		w.classStyle = uint32(w32.GetClassLongPtr(w.handle, w32.GCL_STYLE))
	}
}

func (w *Window) Title() string { return w.title }

func (w *Window) SetTitle(title string) {
	w.title = title
	if w.handle != 0 {
		w32.SetWindowText(w.handle, title)
	}
}

func (w *Window) Style() uint { return w.style }

func (w *Window) SetStyle(ws uint) {
	w.style = ws
	if w.handle != 0 {
		w32.SetWindowLongPtr(w.handle, w32.GWL_STYLE, uintptr(w.style))
		w32.ShowWindow(w.handle, w.state) // for the new style to take effect
		w.style = uint(w32.GetWindowLongPtr(w.handle, w32.GWL_STYLE))
		w.readBounds()
	}
}

func (w *Window) ExtendedStyle() uint { return w.exStyle }

func (w *Window) SetExtendedStyle(x uint) {
	w.exStyle = x
	if w.handle != 0 {
		w32.SetWindowLongPtr(w.handle, w32.GWL_EXSTYLE, uintptr(w.exStyle))
		w32.ShowWindow(w.handle, w.state) // for the new style to take effect
		w.exStyle = uint(w32.GetWindowLongPtr(w.handle, w32.GWL_EXSTYLE))
		w.readBounds()
	}
}

func (w *Window) readBounds() {
	r := w32.GetWindowRect(w.handle)
	w.x = int(r.Left)
	w.y = int(r.Top)
	w.width = int(r.Width())
	w.height = int(r.Height())
}

func (w *Window) X() int {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.x
}

func (w *Window) SetX(x int) {
	w.x = x
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
}

func (w *Window) Y() int {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.y
}

func (w *Window) SetY(y int) {
	w.y = y
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
}

func (w *Window) Pos() (x, y int) {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.x, w.y
}

func (w *Window) SetPos(x, y int) {
	w.x = x
	w.y = y
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
}

func (w *Window) Width() int {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.width
}

func (w *Window) SetWidth(width int) {
	if width <= 0 {
		return
	}
	w.width = width
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
}

func (w *Window) Height() int {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.height
}

func (w *Window) SetHeight(height int) {
	if height <= 0 {
		return
	}
	w.height = height
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
}

func (w *Window) Size() (width, height int) {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.width, w.height
}

func (w *Window) SetSize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	w.width = width
	w.height = height
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return
}

func (w *Window) Bounds() (x, y, width, height int) {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.x, w.y, w.width, w.height
}

func (w *Window) SetBounds(x, y, width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	w.x = x
	w.y = y
	w.width = width
	w.height = height
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
}

func (w *Window) ClientX() int {
	x, _ := w.ClientPos()
	return x
}

func (w *Window) ClientY() int {
	_, y := w.ClientPos()
	return y
}

func (w *Window) ClientPos() (x, y int) {
	if w.handle != 0 {
		x, y = w32.ClientToScreen(w.handle, 0, 0)
	}
	return
}

func (w *Window) ClientWidth() int {
	width, _ := w.ClientSize()
	return width
}

func (w *Window) ClientHeight() int {
	_, height := w.ClientSize()
	return height
}

func (w *Window) ClientSize() (width, height int) {
	if w.handle == 0 {
		if w.width < 0 {
			width = -w.width
		}
		if w.height < 0 {
			height = -w.height
		}
	} else {
		r := w32.GetClientRect(w.handle)
		width = int(r.Width())
		height = int(r.Height())
	}
	return
}

func (w *Window) ClientBounds() (x, y, width, height int) {
	x, y = w.ClientPos()
	width, height = w.ClientSize()
	return
}

func (w *Window) SetClientWidth(width int) {
	if width <= 0 {
		return
	}
	if w.handle != 0 {
		var r w32.RECT
		w32.AdjustWindowRect(&r, w.style, w.menu != nil)
		w.width = width + int(r.Width())
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	} else {
		// save negative size for Show to indicate client size
		w.width = -width
	}
}

func (w *Window) SetClientHeight(height int) {
	if height <= 0 {
		return
	}
	if w.handle != 0 {
		var r w32.RECT
		w32.AdjustWindowRect(&r, w.style, w.menu != nil)
		w.height = height + int(r.Height())
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	} else {
		// save negative size for Show to indicate client size
		w.height = -height
	}
}

func (w *Window) SetClientSize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}
	if w.handle != 0 {
		var r w32.RECT
		w32.AdjustWindowRect(&r, w.style, w.menu != nil)
		w.width = width + int(r.Width())
		w.height = height + int(r.Height())
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	} else {
		// save negative size for Show to indicate client size
		w.width = -width
		w.height = -height
	}
}

func (w *Window) setState(s uint) {
	w.state = w32.SW_MAXIMIZE
	if w.handle != 0 {
		w32.ShowWindow(w.handle, w.state)
	}
}

func (w *Window) readState() {
	var p w32.WINDOWPLACEMENT
	if w32.GetWindowPlacement(w.handle, &p) {
		w.state = int(p.ShowCmd)
	}
}

func (w *Window) Maximized() bool {
	if w.handle != 0 {
		w.readState()
	}
	return w.state == w32.SW_MAXIMIZE
}

func (w *Window) Maximize() {
	w.setState(w32.SW_MAXIMIZE)
}

func (w *Window) Minimized() bool {
	if w.handle != 0 {
		w.readState()
	}
	return w.state == w32.SW_MINIMIZE
}

func (w *Window) Minimize() {
	w.setState(w32.SW_MINIMIZE)
}

func (w *Window) Restore() {
	w.setState(w32.SW_SHOWNORMAL)
}

func (w *Window) GetBackground() w32.HBRUSH { return w.background }

func (w *Window) SetBackground(b w32.HBRUSH) {
	w.background = b
	if w.handle != 0 {
		w32.SetClassLongPtr(w.handle, w32.GCLP_HBRBACKGROUND, uintptr(b))
		w32.InvalidateRect(w.handle, nil, true)
	}
}

func (w *Window) Cursor() w32.HCURSOR { return w.cursor }

func (w *Window) SetCursor(c w32.HCURSOR) {
	w.cursor = c
	if w.handle != 0 {
		w32.SetClassLongPtr(w.handle, w32.GCLP_HCURSOR, uintptr(c))
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
	for _, c := range w.controls {
		c.parentFontChanged()
	}
}

const controlIDOffset = 2

func (w *Window) Add(c Control) {
	c.setParent(w)
	if w.handle != 0 {
		c.create(len(w.controls) + controlIDOffset)
	}
	w.registerControl(c)
}

func (w *Window) registerControl(c Control) {
	w.controls = append(w.controls, c)
}

func (w *Window) controlCount() int {
	return len(w.controls)
}

func (w *Window) SetOnShow(f func()) {
	w.onShow = f
}

func (w *Window) SetOnClose(f func()) {
	w.onClose = f
}

func (w *Window) SetOnCanClose(f func() bool) {
	w.onCanClose = f
}

func (w *Window) SetOnMouseMove(f func(x, y int)) {
	w.onMouseMove = f
}

func (w *Window) SetOnMouseWheel(f func(x, y int, delta float64)) {
	w.onMouseWheel = f
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

func (w *Window) onMsg(window w32.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case w32.WM_MOUSEMOVE:
		if w.onMouseMove != nil {
			w.onMouseMove(
				int(lParam&0xFFFF),
				int(lParam&0xFFFF0000)>>16,
			)
			return 0
		}
	case w32.WM_MOUSEWHEEL:
		if w.onMouseWheel != nil {
			w.onMouseWheel(
				int(lParam&0xFFFF),
				int(lParam&0xFFFF0000)>>16,
				float64(int16(int32(wParam&0xFFFF0000)>>16))/120,
			)
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
		if w.onResize != nil {
			w.onResize()
		}
		w32.InvalidateRect(window, nil, true)
		return 0
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
		return w32.DefWindowProc(window, msg, wParam, lParam)
	}
	return w32.DefWindowProc(window, msg, wParam, lParam)
}

func (w *Window) onWM_DRAWITEM(wParam, lParam uintptr) {
	id := wParam
	index := id - controlIDOffset
	if 0 <= index && index < uintptr(len(w.controls)) {
		if p, ok := w.controls[index].(*Paintbox); ok {
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
		Background: w.background,
		WndProc: syscall.NewCallback(func(
			window w32.HWND,
			msg uint32,
			wParam, lParam uintptr,
		) uintptr {
			return w.onMsg(window, msg, wParam, lParam)
		}),
		Cursor:    w.cursor,
		ClassName: syscall.StringToUTF16Ptr(w.className),
		Style:     w.classStyle,
	}
	atom := w32.RegisterClassEx(&class)
	if atom == 0 {
		return errors.New("win.NewWindow: RegisterClassEx failed")
	}

	w.adjustClientRect()
	window := w32.CreateWindowEx(
		w.exStyle,
		syscall.StringToUTF16Ptr(w.className),
		syscall.StringToUTF16Ptr(w.title),
		w.style,
		w.x, w.y, w.width, w.height,
		0, 0, 0, nil,
	)
	if window == 0 {
		return errors.New("win.Window.Show: CreateWindowEx failed")
	}
	w.handle = window

	w.updateAccelerators()
	w.createContents()
	w.applyIcon()
	w32.ShowWindow(window, w.state)
	w.readBounds()
	if w.onShow != nil {
		w.onShow()
	}

	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) != 0 {
		// TODO this eats VK_ESCAPE and VK_RETURN and makes escape press a
		// focused button?!
		if w.accelTable == 0 ||
			!w32.TranslateAccelerator(w.handle, w.accelTable, &msg) {
			if !w32.IsDialogMessage(w.handle, &msg) {
				w32.TranslateMessage(&msg)
				w32.DispatchMessage(&msg)
			}
		}
	}
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

	for i, c := range w.controls {
		c.create(i + controlIDOffset)
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
		if f := w.shortcuts[index].f; f != nil {
			f()
		}
	} else if lParam != 0 {
		// control clicked
		id := wParam & 0xFFFF
		cmd := (wParam & 0xFFFF0000) >> 16
		index := id - controlIDOffset
		if 0 <= index && index < uintptr(len(w.controls)) {
			control := w.controls[index]
			switch c := control.(type) {
			case *Button:
				if c.onClick != nil {
					c.onClick()
				}
			case *FloatUpDown:
				if cmd == w32.EN_CHANGE {
					// TODO
					//if c.onValueChange != nil {
					//	c.onValueChange(int(c.Value()))
					//}
				}
			case *IntUpDown:
				if cmd == w32.EN_CHANGE {
					if c.onValueChange != nil {
						c.onValueChange(int(c.Value()))
					}
				}
			case *Checkbox:
				state := w32.IsDlgButtonChecked(c.parent.getHandle(), id)
				checked := state == w32.BST_CHECKED
				if c.checked != checked {
					c.checked = checked
					if c.onChange != nil {
						c.onChange(checked)
					}
				}
			case *RadioButton:
				// look through all RadioButtons to see which have
				// changed, first change to false, at the end change
				// the newly selected one to true, always in this
				// order
				var changedToTrue *RadioButton
				for i, c := range w.controls {
					if rb, ok := c.(*RadioButton); ok {
						id := uintptr(i) + controlIDOffset
						state := w32.IsDlgButtonChecked(
							rb.parent.getHandle(),
							id,
						)
						checked := state == w32.BST_CHECKED
						if rb.checked != checked {
							if checked {
								changedToTrue = rb
							} else {
								rb.checked = checked
								if rb.onChange != nil {
									rb.onChange(checked)
								}
							}
						}
					}
				}
				if changedToTrue != nil {
					changedToTrue.checked = true
					if changedToTrue.onChange != nil {
						changedToTrue.onChange(true)
					}
				}
			case *Combobox:
				if cmd == w32.CBN_SELCHANGE {
					if c.onChange != nil {
						c.onChange(c.SelectedIndex())
					}
				}
			case *EditLine:
				if cmd == w32.EN_CHANGE {
					if c.onTextChange != nil {
						c.onTextChange()
					}
				}
			}
			return
		}
	}
}

func (w *Window) onWM_NOTIFY(wParam, lParam uintptr) {
	header := *((*w32.NMHDR)(unsafe.Pointer(lParam)))
	if header.Code == uint32(w32.UDN_DELTAPOS) {
		i := int(wParam) - controlIDOffset
		if 0 <= i && i < len(w.controls) {
			if f, ok := w.controls[i].(*FloatUpDown); ok {
				updown := *((*w32.NMUPDOWN)(unsafe.Pointer(lParam)))
				f.SetValue(f.value - float64(updown.Delta))
			}
		}
	} else if header.Code == w32.LVN_ITEMCHANGED&0xFFFFFFFF {
		i := int(wParam) - controlIDOffset
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

// hideConsoleWindow hides the associated console window if it was created
// because the ldflag H=windowsgui was not provided when building.
func hideConsoleWindow() {
	console := w32.GetConsoleWindow()
	if console == 0 {
		return // no console attached
	}
	// If this application is the process that created the console window, then
	// this program was not compiled with the -H=windowsgui flag and on start-up
	// it created a console along with the main application window. In this case
	// hide the console window.
	// See
	// http://stackoverflow.com/questions/9009333/how-to-check-if-the-program-is-run-from-a-console
	// and thanks to
	// https://github.com/hajimehoshi
	// for the tip.
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
		w32.SendMessage(w.handle, w32.WM_SETICON, w32.ICON_SMALL, w.icon)
		w32.SendMessage(w.handle, w32.WM_SETICON, w32.ICON_SMALL2, w.icon)
		w32.SendMessage(w.handle, w32.WM_SETICON, w32.ICON_BIG, w.icon)
	}
}

func (w *Window) SetIconFromExeResource(resourceID uint16) {
	w.icon = uintptr(w32.LoadImage(
		w32.GetModuleHandle(""),
		w32.MakeIntResource(resourceID),
		w32.IMAGE_ICON,
		0, 0,
		w32.LR_DEFAULTSIZE|w32.LR_SHARED,
	))
	w.applyIcon()
}

func (w *Window) SetIconFromMem(mem []byte) {
	offset := w32.LookupIconIdFromDirectoryEx(
		unsafe.Pointer(&mem[0]),
		true,
		0, 0,
		w32.LR_DEFAULTCOLOR,
	)
	if offset <= 0 {
		return
	}
	w.icon = uintptr(w32.CreateIconFromResourceEx(
		unsafe.Pointer(&mem[offset]),
		uint32(len(mem[offset:])),
		true, // create icon, not cursor
		0x30000,
		0, 0,
		w32.LR_DEFAULTCOLOR|w32.LR_DEFAULTSIZE,
	))
	w.applyIcon()
}

func (w *Window) SetIconFromFile(path string) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return
	}
	w.icon = uintptr(w32.LoadImage(
		0,
		p,
		w32.IMAGE_ICON,
		0, 0,
		w32.LR_LOADFROMFILE,
	))
	w.applyIcon()
}

func (w *Window) adjustClientRect() {
	// if the width or height are negative, this indicates it is the client
	// rect's size
	var r w32.RECT
	w32.AdjustWindowRect(&r, w.style, w.menu != nil)
	if w.width < 0 && w.width != w32.CW_USEDEFAULT {
		w.width = -w.width + int(r.Width())
	}
	if w.height < 0 && w.height != w32.CW_USEDEFAULT {
		w.height = -w.height + int(r.Height())
	}
}

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

	if w.icon == 0 {
		w.icon = w.parent.icon
	}
	if w.font == nil {
		w.font = w.parent.font
	}

	w.adjustClientRect()
	window := w32.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr(w.parent.className),
		syscall.StringToUTF16Ptr(w.title),
		w.style,
		w.x, w.y, w.width, w.height,
		w.parent.handle,
		0, 0, nil,
	)
	if window == 0 {
		return errors.New("win.Window.ShowModal: CreateWindowEx failed")
	}
	w.handle = window

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
	w.createContents()
	w.applyIcon()
	w32.ShowWindow(w.handle, w32.SW_SHOWNORMAL)
	w32.EnableWindow(w.parent.handle, false)
	w.readBounds()
	if w.onShow != nil {
		w.onShow()
	}

	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) != 0 {
		// TODO this eats VK_ESCAPE and VK_RETURN and makes escape press a
		// focused button?!
		if w.accelTable == 0 ||
			!w32.TranslateAccelerator(w.handle, w.accelTable, &msg) {
			if !w32.IsDialogMessage(w.handle, &msg) {
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

func (w *Window) DisableAltF4() {
	w.altF4disabled = true
}

func (w *Window) EnableAltF4() {
	w.altF4disabled = false
}

func (w *Window) Destroy() {
	if w.handle != 0 {
		w32.DestroyWindow(w.handle)
	}
}

type ShortcutKeys struct {
	// Mod is a bit field combining any of ModControl, ModShift, ModAlt
	Mod KeyMod
	// Rune is the characters to be pressed for the accelerator. Either Rune or
	// Key must be set, if both are set, Rune takes preference.
	Rune rune
	// Key is the virtual key to be pressed, it must be a w32.VK_... constant.
	Key uint16
}

type KeyMod int

const (
	ModControl KeyMod = 1 << iota
	ModShift
	ModAlt
)

type shortcut struct {
	keys ShortcutKeys
	f    func()
}

func (s shortcut) toACCEL() w32.ACCEL {
	var a w32.ACCEL
	a.Virt |= w32.FVIRTKEY // NOTE need to set this in any case for some reason
	if s.keys.Mod&ModControl != 0 {
		a.Virt |= w32.FCONTROL
	}
	if s.keys.Mod&ModShift != 0 {
		a.Virt |= w32.FSHIFT
	}
	if s.keys.Mod&ModAlt != 0 {
		a.Virt |= w32.FALT
	}
	if s.keys.Rune != 0 {
		// use rune
		a.Key = utf16.Encode([]rune{unicode.ToUpper(s.keys.Rune)})[0]
	} else {
		// use virtual key
		a.Key = s.keys.Key
	}
	return a
}

func (w *Window) SetShortcut(keys ShortcutKeys, f func()) {
	func() {
		// check if this shortcut was set before
		for i := range w.shortcuts {
			if keys == w.shortcuts[i].keys {
				if f != nil {
					// replace shortcut
					w.shortcuts[i].f = f
				} else {
					// remove shortcut entirely
					copy(w.shortcuts[i:], w.shortcuts[i+1:])
					w.shortcuts = w.shortcuts[:len(w.shortcuts)-1]
				}
				return // shortcut was there before
			}
		}
		// if we land here, the shortcut is new
		w.shortcuts = append(w.shortcuts, shortcut{
			keys: keys,
			f:    f,
		})
	}()
	if w.accelTable != 0 {
		w.updateAccelerators()
	}
}

func (w *Window) updateAccelerators() {
	if w.accelTable != 0 {
		w32.DestroyAcceleratorTable(w.accelTable)
		w.accelTable = 0
	}
	if len(w.shortcuts) > 0 {
		accels := make([]w32.ACCEL, len(w.shortcuts))
		for i := range w.shortcuts {
			accels[i] = w.shortcuts[i].toACCEL()
			accels[i].Cmd = uint16(i)
		}
		w.accelTable = w32.CreateAcceleratorTable(accels)
	}
}

func (w *Window) Scroll(dx, dy int) {
	if w.handle != 0 {
		w32.ScrollWindow(w.handle, dx, dy, nil, nil)
	}
}

func (w *Window) Repaint() {
	if w.handle != 0 {
		w32.InvalidateRect(w.handle, nil, true)
	}
}

func (w *Window) Monitor() w32.HMONITOR {
	if w.handle == 0 {
		return 0
	}
	return w32.MonitorFromWindow(w.handle, w32.MONITOR_DEFAULTTONULL)
}

func (w *Window) Parent() *Window {
	return w.parent
}
