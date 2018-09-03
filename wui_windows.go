package wui

import (
	"errors"
	"io/ioutil"
	"math"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/gonutz/w32"
)

func MessageBox(caption, text string) {
	w32.MessageBox(0, text, caption, w32.MB_OK|w32.MB_TOPMOST)
}

func MessageBoxError(caption, text string) {
	w32.MessageBox(0, text, caption, w32.MB_OK|w32.MB_ICONERROR|w32.MB_TOPMOST)
}

func MessageBoxOKCancel(caption, text string) bool {
	return w32.MessageBox(0, text, caption, w32.MB_OKCANCEL|w32.MB_TOPMOST) == w32.IDOK
}

func MessageBoxYesNo(caption, text string) bool {
	return w32.MessageBox(0, text, caption, w32.MB_YESNO|w32.MB_TOPMOST) == w32.IDYES
}

type Control interface {
	isControl()
}

type Window struct {
	handle      w32.HWND
	className   string
	classStyle  uint32
	title       string
	style       uint
	x           int
	y           int
	width       int
	height      int
	state       int
	background  w32.HBRUSH
	cursor      w32.HCURSOR
	menu        *Menu
	menuStrings []*MenuString
	font        *Font
	controls    []Control
	onShow      func(*Window)
	onClose     func(*Window)
	onMouseMove func(x, y int)
	onKeyDown   func(key int)
	onKeyUp     func(key int)
	onResize    func()
}

func NewWindow() *Window {
	return &Window{
		className:  "wui_window_class",
		x:          w32.CW_USEDEFAULT,
		y:          w32.CW_USEDEFAULT,
		width:      w32.CW_USEDEFAULT,
		height:     w32.CW_USEDEFAULT,
		style:      w32.WS_OVERLAPPEDWINDOW,
		state:      w32.SW_NORMAL,
		background: w32.GetSysColorBrush(w32.COLOR_BTNFACE),
		cursor:     w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_ARROW)),
	}
}

func (w *Window) ClassName() string { return w.className }

func (w *Window) SetClassName(name string) *Window {
	if w.handle != 0 {
		w.className = name
	}
	return w
}

func (w *Window) ClassStyle() uint32 { return w.classStyle }

func (w *Window) SetClassStyle(style uint32) *Window {
	w.classStyle = style
	if w.handle != 0 {
		w32.SetClassLongPtr(w.handle, w32.GCL_STYLE, uintptr(w.classStyle))
		w.classStyle = uint32(w32.GetClassLongPtr(w.handle, w32.GCL_STYLE))
	}
	return w
}

func (w *Window) Title() string { return w.title }

func (w *Window) SetTitle(title string) *Window {
	w.title = title
	if w.handle != 0 {
		w32.SetWindowText(w.handle, title)
	}
	return w
}

func (w *Window) Style() uint { return w.style }

func (w *Window) SetStyle(ws uint) *Window {
	w.style = ws
	if w.handle != 0 {
		w32.SetWindowLongPtr(w.handle, w32.GWL_STYLE, uintptr(w.style))
		w32.ShowWindow(w.handle, w.state) // for the new style to take effect
		w.style = uint(w32.GetWindowLongPtr(w.handle, w32.GWL_STYLE))
		w.readBounds()
	}
	return w
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

func (w *Window) SetX(x int) *Window {
	w.x = x
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return w
}

func (w *Window) Y() int {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.y
}

func (w *Window) SetY(y int) *Window {
	w.y = y
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return w
}

func (w *Window) Pos() (x, y int) {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.x, w.y
}

func (w *Window) SetPos(x, y int) *Window {
	w.x = x
	w.y = y
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return w
}

func (w *Window) Width() int {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.width
}

func (w *Window) SetWidth(width int) *Window {
	if width <= 0 {
		return w
	}
	w.width = width
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return w
}

func (w *Window) Height() int {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.height
}

func (w *Window) SetHeight(height int) *Window {
	if height <= 0 {
		return w
	}
	w.height = height
	if w.handle != 0 {
		w32.SetWindowPos(
			w.handle, 0,
			w.x, w.y, w.width, w.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return w
}

func (w *Window) Size() (width, height int) {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.width, w.height
}

func (w *Window) SetSize(width, height int) *Window {
	if width <= 0 || height <= 0 {
		return w
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
	return w
}

func (w *Window) Bounds() (x, y, width, height int) {
	if w.handle != 0 {
		w.readBounds()
	}
	return w.x, w.y, w.width, w.height
}

func (w *Window) SetBounds(x, y, width, height int) *Window {
	if width <= 0 || height <= 0 {
		return w
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
	return w
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
			height = -height
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

func (w *Window) SetClientWidth(width int) *Window {
	if width <= 0 {
		return w
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
	return w
}

func (w *Window) SetClientHeight(height int) *Window {
	if height <= 0 {
		return w
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
	return w
}

func (w *Window) SetClientSize(width, height int) *Window {
	if width <= 0 || height <= 0 {
		return w
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
	return w
}

func (w *Window) setState(s uint) *Window {
	w.state = w32.SW_MAXIMIZE
	if w.handle != 0 {
		w32.ShowWindow(w.handle, w.state)
	}
	return w
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

func (w *Window) Maximize() *Window {
	return w.setState(w32.SW_MAXIMIZE)
}

func (w *Window) Minimized() bool {
	if w.handle != 0 {
		w.readState()
	}
	return w.state == w32.SW_MINIMIZE
}

func (w *Window) Minimize() *Window {
	return w.setState(w32.SW_MINIMIZE)
}

func (w *Window) Restore() *Window {
	return w.setState(w32.SW_SHOWNORMAL)
}

func (w *Window) GetBackground() w32.HBRUSH { return w.background }

func (w *Window) SetBackground(b w32.HBRUSH) *Window {
	w.background = b
	if w.handle != 0 {
		w32.SetClassLongPtr(w.handle, w32.GCLP_HBRBACKGROUND, uintptr(b))
		w32.InvalidateRect(w.handle, nil, true)
	}
	return w
}

func (w *Window) Cursor() w32.HCURSOR { return w.cursor }

func (w *Window) SetCursor(c w32.HCURSOR) *Window {
	w.cursor = c
	if w.handle != 0 {
		w32.SetClassLongPtr(w.handle, w32.GCLP_HCURSOR, uintptr(c))
	}
	return w
}

func (w *Window) Menu() *Menu {
	return w.menu
}

func (w *Window) SetMenu(m *Menu) *Window {
	w.menu = m
	if w.handle != 0 {
		// TODO update menu
	}
	return w
}

type MenuItem interface {
	isMenuItem()
}

type Menu struct {
	name  string
	items []MenuItem
}

func NewMenu(name string) *Menu {
	return &Menu{name: name}
}

func (*Menu) isMenuItem() {}

func (m *Menu) Add(item MenuItem) *Menu {
	m.items = append(m.items, item)
	return m
}

func NewMenuString(name string) *MenuString {
	return &MenuString{name: name}
}

type MenuString struct {
	name    string
	onClick func()
}

func (*MenuString) isMenuItem() {}

func (m *MenuString) SetOnClick(f func()) *MenuString {
	m.onClick = f
	return m
}

func NewMenuSeparator() MenuItem {
	return separator
}

type menuSeparator int

func (menuSeparator) isMenuItem() {}

var separator menuSeparator

func (w *Window) Font() *Font {
	return w.font
}

func (w *Window) SetFont(f *Font) *Window {
	w.font = f
	if f == nil {
		return w
	}
	if w.handle != 0 {
		f.create()
		for _, control := range w.controls {
			var handle w32.HWND
			switch c := control.(type) {
			case *Button:
				handle = c.handle
			case *NumberUpDown:
				handle = c.editHandle
			case *Label:
				handle = c.handle
			case *Paintbox:
				handle = c.handle
			default:
				panic("unhandled control type")
			}
			w32.SendMessage(handle, w32.WM_SETFONT, uintptr(w.font.handle), 1)
		}
	}
	return w
}

const controlIDOffset = 2

func (w *Window) Add(c Control) *Window {
	w.controls = append(w.controls, c)
	if w.handle != 0 {
		createControl(
			c,
			w,
			len(w.controls)-1+controlIDOffset,
			w32.HINSTANCE(w32.GetWindowLong(w.handle, w32.GWL_HINSTANCE)),
		)
	}
	return w
}

func (w *Window) SetOnShow(f func(*Window)) *Window {
	w.onShow = f
	return w
}

func (w *Window) SetOnClose(f func(*Window)) *Window {
	w.onClose = f
	return w
}

func (w *Window) SetOnMouseMove(f func(x, y int)) *Window {
	w.onMouseMove = f
	return w
}

func (w *Window) SetOnKeyDown(f func(key int)) *Window {
	w.onKeyDown = f
	return w
}

func (w *Window) SetOnKeyUp(f func(key int)) *Window {
	w.onKeyUp = f
	return w
}

func (w *Window) SetOnResize(f func()) *Window {
	w.onResize = f
	return w
}

func (w *Window) Close() {
	if w.handle != 0 {
		w32.SendMessage(w.handle, w32.WM_CLOSE, 0, 0)
	}
}

func (w *Window) Show() error {
	if w.handle != 0 {
		return errors.New("wui.Window.Show: window already visible")
	}

	hideConsoleWindow()

	runtime.LockOSThread()

	setManifest()

	{
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

	class := w32.WNDCLASSEX{
		Background: w.background,
		WndProc: syscall.NewCallback(func(
			window w32.HWND,
			msg uint32,
			wParam, lParam uintptr,
		) uintptr {
			switch msg {
			case w32.WM_MOUSEMOVE:
				if w.onMouseMove != nil {
					w.onMouseMove(
						int(lParam&0xFFFF),
						int(lParam&0xFFFF0000)>>16,
					)
					return 0
				}
			case w32.WM_DRAWITEM:
				id := wParam
				index := id - controlIDOffset
				if 0 <= index && index < uintptr(len(w.controls)) {
					if p, ok := w.controls[index].(*Paintbox); ok {
						if p.onPaint != nil {
							p.onPaint(&Canvas{
								hdc:    ((*w32.DRAWITEMSTRUCT)(unsafe.Pointer(lParam))).HDC,
								width:  p.width,
								height: p.height,
							})
						}
					}
				}
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
				if lParam == 0 && wParam&0xFFFF0000 == 0 {
					id := int(wParam & 0xFFFF)
					if 0 <= id && id < len(w.menuStrings) {
						f := w.menuStrings[id].onClick
						if f != nil {
							f()
						}
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
						case *NumberUpDown:
							if cmd == w32.EN_CHANGE {
								if c.onValueChange != nil {
									c.onValueChange(int(c.Value()))
								}
							}
						}
						return 0
					}
				}

			case w32.WM_SIZE:
				if w.onResize != nil {
					w.onResize()
				}
				w32.InvalidateRect(window, nil, true)
				return 0
			case w32.WM_DESTROY:
				if w.onClose != nil {
					w.onClose(w)
				}
				w32.PostQuitMessage(0)
				return 0
			}
			return w32.DefWindowProc(window, msg, wParam, lParam)
		}),
		Cursor:    w.cursor,
		ClassName: syscall.StringToUTF16Ptr(w.className),
		Style:     w.classStyle,
	}
	atom := w32.RegisterClassEx(&class)
	if atom == 0 {
		return errors.New("win.NewWindow: RegisterClassEx failed")
	}
	window := w32.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr(w.className),
		syscall.StringToUTF16Ptr(w.title),
		w.style,
		w.x, w.y, w.width, w.height,
		0, 0, 0, nil,
	)
	if window == 0 {
		return errors.New("win.NewWindow: CreateWindowEx failed")
	}
	w.handle = window

	if w.font != nil {
		w.font.create()
	}

	if w.menu != nil {
		var addItems func(m w32.HMENU, items []MenuItem)
		addItems = func(m w32.HMENU, items []MenuItem) {
			for _, item := range items {
				switch menuItem := item.(type) {
				case *Menu:
					menu := w32.CreateMenu()
					w32.AppendMenu(m, w32.MF_POPUP, uintptr(menu),
						menuItem.name)
					addItems(menu, menuItem.items)
				case *MenuString:
					w32.AppendMenu(
						m,
						w32.MF_STRING,
						uintptr(len(w.menuStrings)),
						menuItem.name,
					)
					w.menuStrings = append(w.menuStrings, menuItem)
				case menuSeparator:
					w32.AppendMenu(m, w32.MF_SEPARATOR, 0, "")
				}
			}
		}
		menuBar := w32.CreateMenu()
		addItems(menuBar, w.menu.items)
		w32.SetMenu(window, menuBar)
	}

	instance := w32.HINSTANCE(w32.GetWindowLong(window, w32.GWL_HINSTANCE))
	for i, c := range w.controls {
		createControl(c, w, i+controlIDOffset, instance)
	}

	w32.ShowWindow(window, w.state)
	w.readBounds()
	if w.onShow != nil {
		w.onShow(w)
	}

	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) != 0 {
		// TODO this eats VK_ESCAPE and VK_RETURN and makes escape press a
		// focused button?!
		if !w32.IsDialogMessage(w.handle, &msg) {
			w32.TranslateMessage(&msg)
			w32.DispatchMessage(&msg)
		}
	}
	return nil
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

func createControl(
	control Control,
	parent *Window,
	id int,
	instance w32.HINSTANCE,
) {
	switch c := control.(type) {
	case *Button:
		c.handle = w32.CreateWindowExStr(
			0,
			"BUTTON",
			c.text,
			w32.WS_VISIBLE|w32.WS_CHILD|w32.WS_TABSTOP|w32.BS_DEFPUSHBUTTON,
			c.x, c.y, c.width, c.height,
			parent.handle, w32.HMENU(id), instance, nil,
		)
		if c.disabled {
			w32.EnableWindow(c.handle, false)
		}
		if parent.font != nil {
			w32.SendMessage(
				c.handle,
				w32.WM_SETFONT,
				uintptr(parent.font.handle),
				1,
			)
		}
	case *NumberUpDown:
		upDown := w32.CreateWindowStr(
			w32.UPDOWN_CLASS,
			"",
			w32.WS_VISIBLE|w32.WS_CHILD|
				w32.UDS_SETBUDDYINT|w32.UDS_ALIGNRIGHT|w32.UDS_ARROWKEYS,
			c.x, c.y, c.width, c.height,
			parent.handle, 0, instance, nil,
		)
		edit := w32.CreateWindowExStr(
			w32.WS_EX_CLIENTEDGE,
			"EDIT",
			strconv.Itoa(int(c.value)),
			w32.WS_TABSTOP|w32.WS_VISIBLE|w32.WS_CHILD|w32.ES_NUMBER,
			c.x, c.y, c.width, c.height,
			parent.handle, w32.HMENU(id), instance, nil,
		)
		if c.disabled {
			w32.EnableWindow(edit, false)
		}
		w32.SendMessage(upDown, w32.UDM_SETBUDDY, uintptr(edit), 0)
		w32.SendMessage(
			upDown,
			w32.UDM_SETRANGE32,
			uintptr(c.minValue),
			uintptr(c.maxValue),
		)
		c.upDownHandle = upDown
		c.editHandle = edit
		if parent.font != nil {
			w32.SendMessage(
				edit,
				w32.WM_SETFONT,
				uintptr(parent.font.handle),
				1,
			)
		}
	case *Label:
		c.handle = w32.CreateWindowExStr(
			0,
			"STATIC",
			c.text,
			w32.WS_VISIBLE|w32.WS_CHILD|w32.SS_CENTERIMAGE|c.align,
			c.x, c.y, c.width, c.height,
			parent.handle, w32.HMENU(id), instance, nil,
		)
		if parent.font != nil {
			w32.SendMessage(
				c.handle,
				w32.WM_SETFONT,
				uintptr(parent.font.handle),
				1,
			)
		}
	case *Paintbox:
		c.handle = w32.CreateWindowExStr(
			0,
			"STATIC",
			"",
			w32.WS_VISIBLE|w32.WS_CHILD|w32.SS_OWNERDRAW,
			c.x, c.y, c.width, c.height,
			parent.handle, w32.HMENU(id), instance, nil,
		)
		if parent.font != nil {
			w32.SendMessage(
				c.handle,
				w32.WM_SETFONT,
				uintptr(parent.font.handle),
				1,
			)
		}
	default:
		panic("unhandled control type")
	}
}

type Font struct {
	handle     w32.HFONT
	name       string
	height     int
	bold       bool
	italic     bool
	underlined bool
	strikedOut bool
}

func NewFont() *Font {
	return &Font{}
}

func (f *Font) Name() string     { return f.name }
func (f *Font) Height() int      { return f.height }
func (f *Font) Bold() bool       { return f.bold }
func (f *Font) Italic() bool     { return f.italic }
func (f *Font) Underlined() bool { return f.underlined }
func (f *Font) StrikedOut() bool { return f.strikedOut }

func (f *Font) create() {
	if f.handle != 0 {
		w32.DeleteObject(w32.HGDIOBJ(f.handle))
	}
	weight := int32(w32.FW_NORMAL)
	if f.bold {
		weight = w32.FW_BOLD
	}
	byteBool := func(b bool) byte {
		if b {
			return 1
		}
		return 0
	}
	desc := w32.LOGFONT{
		Height:         int32(f.height),
		Width:          0,
		Escapement:     0,
		Orientation:    0,
		Weight:         weight,
		Italic:         byteBool(f.italic),
		Underline:      byteBool(f.underlined),
		StrikeOut:      byteBool(f.strikedOut),
		CharSet:        w32.DEFAULT_CHARSET,
		OutPrecision:   w32.OUT_CHARACTER_PRECIS,
		ClipPrecision:  w32.CLIP_CHARACTER_PRECIS,
		Quality:        w32.DEFAULT_QUALITY,
		PitchAndFamily: w32.DEFAULT_PITCH | w32.FF_DONTCARE,
	}
	copy(desc.FaceName[:], utf16.Encode([]rune(f.name)))
	f.handle = w32.CreateFontIndirect(&desc)
}

func (f *Font) SetName(name string) *Font {
	f.name = name
	return f
}

func (f *Font) SetHeight(height int) *Font {
	f.height = height
	return f
}

func (f *Font) SetBold(bold bool) *Font {
	f.bold = bold
	return f
}

func (f *Font) SetItalic(italic bool) *Font {
	f.italic = italic
	return f
}

func (f *Font) SetUnderlined(underlined bool) *Font {
	f.underlined = underlined
	return f
}

func (f *Font) SetStrikedOut(strikedOut bool) *Font {
	f.strikedOut = strikedOut
	return f
}

type Button struct {
	handle   w32.HWND
	text     string
	x        int
	y        int
	width    int
	height   int
	disabled bool
	onClick  func()
}

func (*Button) isControl() {}

func NewButton() *Button {
	return &Button{}
}

func (b *Button) X() int      { return b.x }
func (b *Button) Y() int      { return b.y }
func (b *Button) Width() int  { return b.width }
func (b *Button) Height() int { return b.height }

func (b *Button) SetText(text string) *Button {
	b.text = text
	if b.handle != 0 {
		w32.SetWindowText(b.handle, b.text)
	}
	return b
}

func (b *Button) SetBounds(x, y, width, height int) *Button {
	b.x = x
	b.y = y
	b.width = width
	b.height = height
	if b.handle != 0 {
		w32.SetWindowPos(
			b.handle, 0,
			b.x, b.y, b.width, b.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return b
}

func (b *Button) Enabled() bool {
	return !b.disabled
}

func (b *Button) SetEnabled(e bool) *Button {
	b.disabled = !e
	if b.handle != 0 {
		w32.EnableWindow(b.handle, e)
	}
	return b
}

func (b *Button) SetOnClick(f func()) *Button {
	b.onClick = f
	return b
}

type NumberUpDown struct {
	upDownHandle  w32.HWND
	editHandle    w32.HWND
	x             int
	y             int
	width         int
	height        int
	value         int32
	minValue      int32
	maxValue      int32
	disabled      bool
	onValueChange func(value int)
}

func (*NumberUpDown) isControl() {}

func NewNumberUpDown() *NumberUpDown {
	return &NumberUpDown{
		minValue: math.MinInt32,
		maxValue: math.MaxInt32,
	}
}

func (n *NumberUpDown) Enabled() bool {
	return !n.disabled
}

func (n *NumberUpDown) SetEnabled(e bool) *NumberUpDown {
	n.disabled = !e
	if n.editHandle != 0 {
		w32.EnableWindow(n.editHandle, e)
	}
	return n
}

func (n *NumberUpDown) Value() int {
	if n.upDownHandle != 0 {
		n.value = int32(w32.SendMessage(n.upDownHandle, w32.UDM_GETPOS32, 0, 0))
	}
	return int(n.value)
}

func (n *NumberUpDown) SetValue(v int) *NumberUpDown {
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
	return n
}

func (n *NumberUpDown) SetMinValue(min int) *NumberUpDown {
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
	return n
}

func (n *NumberUpDown) SetMaxValue(max int) *NumberUpDown {
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
	return n
}

func (n *NumberUpDown) SetMinMaxValues(min, max int) *NumberUpDown {
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
	return n
}

func (n *NumberUpDown) SetBounds(x, y, width, height int) *NumberUpDown {
	n.x = x
	n.y = y
	n.width = width
	n.height = height
	if n.editHandle != 0 {
		w32.SetWindowPos(
			n.editHandle, 0,
			n.x, n.y, n.width, n.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
		w32.SetWindowPos(
			n.upDownHandle, 0,
			n.x, n.y, n.width, n.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
		w32.SendMessage(
			n.upDownHandle,
			w32.UDM_SETBUDDY,
			uintptr(n.editHandle),
			0,
		)
	}
	return n
}

func (n *NumberUpDown) SetOnValueChange(f func(value int)) *NumberUpDown {
	n.onValueChange = f
	return n
}

type Label struct {
	handle w32.HWND
	x      int
	y      int
	width  int
	height int
	text   string
	align  uint
}

func NewLabel() *Label {
	return &Label{
		align: w32.SS_LEFT,
	}
}

func (*Label) isControl() {}

func (l *Label) X() int      { return l.x }
func (l *Label) Y() int      { return l.y }
func (l *Label) Width() int  { return l.width }
func (l *Label) Height() int { return l.height }

func (l *Label) SetText(text string) *Label {
	l.text = text
	if l.handle != 0 {
		w32.SetWindowText(l.handle, text)
	}
	return l
}

func (l *Label) SetBounds(x, y, width, height int) *Label {
	l.x = x
	l.y = y
	l.width = width
	l.height = height
	if l.handle != 0 {
		w32.SetWindowPos(
			l.handle, 0,
			l.x, l.y, l.width, l.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return l
}

func (l *Label) setAlign(align uint) *Label {
	l.align = align
	if l.handle != 0 {
		style := uint(w32.GetWindowLongPtr(l.handle, w32.GWL_STYLE))
		style = style &^ w32.SS_LEFT &^ w32.SS_CENTER &^ w32.SS_RIGHT
		w32.SetWindowLongPtr(l.handle, w32.GWL_STYLE, uintptr(style|l.align))
		w32.InvalidateRect(l.handle, nil, true)
	}
	return l
}

func (l *Label) SetLeftAlign() *Label {
	return l.setAlign(w32.SS_LEFT)
}

func (l *Label) SetCenterAlign() *Label {
	return l.setAlign(w32.SS_CENTER)
}

func (l *Label) SetRightAlign() *Label {
	return l.setAlign(w32.SS_RIGHT)
}

type Paintbox struct {
	handle  w32.HWND
	x       int
	y       int
	width   int
	height  int
	onPaint func(*Canvas)
}

func NewPaintbox() *Paintbox {
	return &Paintbox{}
}

func (*Paintbox) isControl() {}

func (p *Paintbox) Paint() {
	if p.handle != 0 {
		w32.InvalidateRect(p.handle, nil, true)
	}
}

func (p *Paintbox) SetBounds(x, y, width, height int) *Paintbox {
	p.x = x
	p.y = y
	p.width = width
	p.height = height
	if p.handle != 0 {
		w32.SetWindowPos(
			p.handle, 0,
			p.x, p.y, p.width, p.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return p
}

func (p *Paintbox) SetOnPaint(f func(*Canvas)) *Paintbox {
	p.onPaint = f
	return p
}

type Color w32.COLORREF

func (c Color) R() uint8 { return uint8(c & 0xFF) }
func (c Color) G() uint8 { return uint8((c & 0xFF00) >> 8) }
func (c Color) B() uint8 { return uint8((c & 0xFF0000) >> 16) }

func RGB(r, g, b uint8) Color {
	return Color(r) + Color(g)<<8 + Color(b)<<16
}

type Canvas struct {
	hdc    w32.HDC
	width  int
	height int
}

func (c *Canvas) Size() (width, height int) {
	width, height = c.width, c.height
	return
}

func (c *Canvas) Width() int {
	return c.width
}

func (c *Canvas) Height() int {
	return c.height
}

func (c *Canvas) DrawRect(x, y, width, height int, color Color) {
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.NULL_BRUSH))
	w32.Rectangle(c.hdc, x, y, x+width, y+height)
}

func (c *Canvas) FillRect(x, y, width, height int, color Color) {
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_BRUSH))
	w32.SetDCBrushColor(c.hdc, w32.COLORREF(color))
	w32.Rectangle(c.hdc, x, y, x+width, y+height)
}

func (c *Canvas) Line(x1, y1, x2, y2 int, color Color) {
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.MoveToEx(c.hdc, x1, y1, nil)
	w32.LineTo(c.hdc, x2, y2)
}

func (c *Canvas) DrawEllipse(x, y, width, height int, color Color) {
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.NULL_BRUSH))
	w32.Ellipse(c.hdc, x, y, x+width, y+height)
}

func (c *Canvas) FillEllipse(x, y, width, height int, color Color) {
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_BRUSH))
	w32.SetDCBrushColor(c.hdc, w32.COLORREF(color))
	w32.Ellipse(c.hdc, x, y, x+width, y+height)
}

func (c *Canvas) Polyline(p []w32.POINT, color Color) {
	if len(p) < 2 {
		return
	}
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.NULL_BRUSH))
	w32.Polyline(c.hdc, p)
}

func (c *Canvas) Polygon(p []w32.POINT, color Color) {
	if len(p) < 2 {
		return
	}
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_BRUSH))
	w32.SetDCBrushColor(c.hdc, w32.COLORREF(color))
	w32.Polygon(c.hdc, p)
}

func (c *Canvas) Arc(x, y, width, height int, fromClockAngle, dAngle float64, color Color) {
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	c.arcLike(x, y, width, height, fromClockAngle, dAngle, w32.Arc)
}

func (c *Canvas) FillPie(x, y, width, height int, fromClockAngle, dAngle float64, color Color) {
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_BRUSH))
	w32.SetDCBrushColor(c.hdc, w32.COLORREF(color))
	c.arcLike(x, y, width, height, fromClockAngle, dAngle, w32.Pie)
}

func (c *Canvas) DrawPie(x, y, width, height int, fromClockAngle, dAngle float64, color Color) {
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.NULL_BRUSH))
	c.arcLike(x, y, width, height, fromClockAngle, dAngle, w32.Pie)
}

func (c *Canvas) arcLike(
	x, y, width, height int,
	fromClockAngle, dAngle float64,
	draw func(w32.HDC, int, int, int, int, int, int, int, int) bool) {
	toRad := func(clock float64) float64 {
		return (90 - clock) * math.Pi / 180
	}
	a, b := fromClockAngle+dAngle, fromClockAngle
	if dAngle < 0 {
		a, b = b, a
	}
	y1, x1 := math.Sincos(toRad(a))
	y2, x2 := math.Sincos(toRad(b))
	x1, x2, y1, y2 = 100*x1, 100*x2, -100*y1, -100*y2
	round := func(f float64) int {
		if f < 0 {
			return int(f - 0.5)
		}
		return int(f + 0.5)
	}
	cx := float64(x) + float64(width)/2.0
	cy := float64(y) + float64(height)/2.0
	draw(
		c.hdc,
		x, y, x+width, y+height,
		round(cx+100*x1), round(cy+100*y1), round(cx+100*x2), round(cy+100*y2),
	)
}

func (c *Canvas) TextExtent(s string) (width, height int) {
	size, ok := w32.GetTextExtentPoint32(c.hdc, s)
	if ok {
		width = int(size.CX)
		height = int(size.CY)
	}
	return
}

func (c *Canvas) TextOut(x, y int, s string, color Color) {
	w32.SetBkMode(c.hdc, w32.TRANSPARENT)
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.NULL_BRUSH))
	w32.SetTextColor(c.hdc, w32.COLORREF(color))
	w32.TextOut(c.hdc, x, y, s)
	w32.SetBkMode(c.hdc, w32.OPAQUE)
}

func (c *Canvas) SetFont(font *Font) *Canvas {
	// TODO this creates a new font every time, either cache them or create one
	// per canvas
	font.create()
	w32.SelectObject(c.hdc, w32.HGDIOBJ(font.handle))
	return c
}

// NOTE this was the first attempt, restructuring the API
//func NewWindow() *Window {
//	return &Window{
//		X:           w32.CW_USEDEFAULT,
//		Y:           w32.CW_USEDEFAULT,
//		Width:       w32.CW_USEDEFAULT,
//		Height:      w32.CW_USEDEFAULT,
//		Cursor:      w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_ARROW)),
//		Style:       w32.WS_OVERLAPPEDWINDOW,
//		ShowCommand: w32.SW_SHOWNORMAL,
//	}
//}
//
//type Window struct {
//	Handle     w32.HWND
//	Title      string
//	Background w32.HBRUSH
//	Cursor     w32.HCURSOR
//	ClassName  string
//	ClassStyle uint32
//	Style      uint
//	X          int
//	Y          int
//	// Width is the outer window width if it is >= 0.
//	// If Width is negative, it is the negative client area width.
//	Width int
//	// Height is the outer window height if it is >= 0.
//	// If Height is negative, it is the negative client area height.
//	Height      int
//	ShowCommand int
//	Font        *Font
//	Menu        *Menu
//	Controls    []*Control
//	OnKeyDown   func(key int)
//	OnKeyUp     func(key int)
//	OnMouseMove func(x, y int)
//	OnPaint     func(p Painter)
//	font        w32.HFONT
//	menuItems   []*MenuItem
//}
//
//type Painter w32.HDC
//
//func (p Painter) TextOut(x, y int, text string) {
//	w32.TextOut(w32.HDC(p), x, y, text)
//}
//
//type Control struct {
//	Handle   w32.HWND
//	Text     string
//	X        int
//	Y        int
//	Width    int
//	Height   int
//	Enabled  bool
//	OnClick  func()
//	class    string
//	style    uint
//	concrete interface{}
//}
//
//func (c *Control) SetEnabled(e bool) {
//	c.Enabled = e
//	if c.Handle != 0 {
//		w32.EnableWindow(c.Handle, e)
//	}
//}
//
//func (c *Control) SetPos(x, y, width, height int) {
//	c.X, c.Y, c.Width, c.Height = x, y, width, height
//	if c.Handle != 0 {
//		w32.SetWindowPos(c.Handle, 0,
//			x, y, width, height,
//			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
//		)
//	}
//}
//
//func (w *Window) Show() (int, error) {
//	runtime.LockOSThread()
//
//	setManifest()
//	w32.InitCommonControlsEx(
//		&w32.INITCOMMONCONTROLSEX{ICC: w32.ICC_UPDOWN_CLASS},
//	)
//
//	if w.ClassName == "" {
//		w.ClassName = "wui_window_class"
//	}
//	var clientRect w32.RECT
//	w32.AdjustWindowRect(&clientRect, w.Style, w.Menu != nil)
//	if w.Width < 0 && w.Width != w32.CW_USEDEFAULT {
//		w.Width = -w.Width + int(clientRect.Width())
//	}
//	if w.Height < 0 && w.Height != w32.CW_USEDEFAULT {
//		w.Height = -w.Height + int(clientRect.Height())
//	}
//	if w.Background == 0 {
//		w.Background = w32.GetSysColorBrush(w32.COLOR_BTNFACE)
//	}
//
//	class := w32.WNDCLASSEX{
//		Background: w.Background,
//		WndProc: syscall.NewCallback(func(
//			window w32.HWND,
//			msg uint32,
//			wParam, lParam uintptr,
//		) uintptr {
//			switch msg {
//			case w32.WM_MOUSEMOVE:
//				if w.OnMouseMove != nil {
//					w.OnMouseMove(
//						int(lParam&0xFFFF),
//						int(lParam&0xFFFF0000)>>16,
//					)
//					return 0
//				}
//			case w32.WM_KEYDOWN:
//				if w.OnKeyDown != nil {
//					w.OnKeyDown(int(wParam))
//					return 0
//				}
//			case w32.WM_KEYUP:
//				if w.OnKeyUp != nil {
//					w.OnKeyUp(int(wParam))
//					return 0
//				}
//			case w32.WM_COMMAND:
//				if lParam == 0 && wParam&0xFFFF0000 == 0 {
//					// menu item clicked
//					id := int(wParam & 0xFFFF)
//					if 0 <= id && id < len(w.menuItems) {
//						f := w.menuItems[id].OnClick
//						if f != nil {
//							f()
//						}
//					}
//				} else if lParam != 0 {
//					// control clicked
//					id := wParam & 0xFFFF
//					index := id - controlIDOffset
//					if 0 <= index && index < uintptr(len(w.Controls)) {
//						c := w.Controls[index]
//						if c.OnClick != nil {
//							c.OnClick()
//						}
//						if check, ok := c.concrete.(*Checkbox); ok {
//							state := w32.IsDlgButtonChecked(w.Handle, id)
//							checked := state == w32.BST_CHECKED
//							if checked != check.Checked {
//								check.Checked = checked
//								if check.OnCheckChange != nil {
//									check.OnCheckChange(check.Checked)
//								}
//							}
//						}
//						return 0
//					}
//				}
//			case w32.WM_PAINT:
//				if w.OnPaint != nil {
//					func() {
//						var ps w32.PAINTSTRUCT
//						hdc := w32.BeginPaint(window, &ps)
//						defer w32.EndPaint(window, &ps)
//						var b w32.LOGBRUSH
//						w32.GetObject(
//							w32.HGDIOBJ(w.Background),
//							unsafe.Sizeof(b),
//							unsafe.Pointer(&b),
//						)
//						w32.SetBkColor(hdc, b.LbColor)
//						if w.font != 0 {
//							w32.SelectObject(hdc, w32.HGDIOBJ(w.font))
//						}
//						w.OnPaint(Painter(hdc))
//					}()
//					return 0
//				}
//			case w32.WM_DESTROY:
//				w32.PostQuitMessage(0)
//				return 0
//			}
//			return w32.DefWindowProc(window, msg, wParam, lParam)
//		}),
//		Cursor:    w.Cursor,
//		ClassName: syscall.StringToUTF16Ptr(w.ClassName),
//		Style:     w.ClassStyle,
//	}
//	atom := w32.RegisterClassEx(&class)
//	if atom == 0 {
//		return 0, errors.New("win.NewWindow: RegisterClassEx failed")
//	}
//	window := w32.CreateWindowEx(
//		0,
//		syscall.StringToUTF16Ptr(w.ClassName),
//		syscall.StringToUTF16Ptr(w.Title),
//		w.Style,
//		w.X, w.Y, w.Width, w.Height,
//		0, 0, 0, nil,
//	)
//	if window == 0 {
//		return 0, errors.New("win.NewWindow: CreateWindowEx failed")
//	}
//	w.Handle = window
//
//	if w.Font != nil {
//		height := int32(w.Font.Height)
//		weight := int32(w32.FW_NORMAL)
//		if w.Font.Bold {
//			weight = w32.FW_BOLD
//		}
//		var italic byte
//		if w.Font.Italic {
//			italic = 1
//		}
//		var underlined byte
//		if w.Font.Underlined {
//			underlined = 1
//		}
//		var strikedOut byte
//		if w.Font.StrikedOut {
//			strikedOut = 1
//		}
//		f := w32.LOGFONT{
//			Height:         height,
//			Width:          0,
//			Escapement:     0,
//			Orientation:    0,
//			Weight:         weight,
//			Italic:         italic,
//			Underline:      underlined,
//			StrikeOut:      strikedOut,
//			CharSet:        w32.DEFAULT_CHARSET,
//			OutPrecision:   w32.OUT_CHARACTER_PRECIS,
//			ClipPrecision:  w32.CLIP_CHARACTER_PRECIS,
//			Quality:        w32.DEFAULT_QUALITY,
//			PitchAndFamily: w32.DEFAULT_PITCH | w32.FF_DONTCARE,
//		}
//		copy(f.FaceName[:], utf16.Encode([]rune(w.Font.Name)))
//		w.font = w32.CreateFontIndirect(&f)
//		defer w32.DeleteObject(w32.HGDIOBJ(w.font))
//	}
//
//	if w.Menu != nil {
//		var addItems func(m w32.HMENU, items []menuItem)
//		addItems = func(m w32.HMENU, items []menuItem) {
//			for _, item := range items {
//				switch menuItem := item.(type) {
//				case *Menu:
//					menu := w32.CreateMenu()
//					w32.AppendMenu(m, w32.MF_POPUP, uintptr(menu),
// menuItem.name)
//					addItems(menu, menuItem.items)
//				case *MenuItem:
//					w32.AppendMenu(
//						m,
//						w32.MF_STRING,
//						uintptr(len(w.menuItems)),
//						menuItem.name,
//					)
//					w.menuItems = append(w.menuItems, menuItem)
//				case *menuSeparator:
//					w32.AppendMenu(m, w32.MF_SEPARATOR, 0, "")
//				}
//			}
//		}
//		menuBar := w32.CreateMenu()
//		addItems(menuBar, w.Menu.items)
//		w32.SetMenu(window, menuBar)
//	}
//
//	instance := w32.HINSTANCE(w32.GetWindowLong(window, w32.GWL_HINSTANCE))
//	focussed := false
//	for i, c := range w.Controls {
//		c.Handle = w32.CreateWindowExStr(
//			0,
//			c.class,
//			c.Text,
//			w32.WS_VISIBLE|w32.WS_CHILD|c.style,
//			c.X, c.Y, c.Width, c.Height,
//			w.Handle,
//			w32.HMENU(i+controlIDOffset),
//			instance,
//			nil,
//		)
//		if !focussed && c.Enabled && c.style&w32.WS_TABSTOP != 0 {
//			focussed = true
//			w32.SetFocus(c.Handle)
//		}
//		if w.font != 0 {
//			w32.SendMessage(c.Handle, w32.WM_SETFONT, uintptr(w.font), 1)
//		}
//		w32.EnableWindow(c.Handle, c.Enabled)
//		if check, ok := c.concrete.(*Checkbox); ok {
//			w32.SendMessage(
//				c.Handle,
//				w32.BM_SETCHECK,
//				toCheckState(check.Checked),
//				0,
//			)
//		}
//	}
//
//	w32.ShowWindow(window, w.ShowCommand)
//
//	var msg w32.MSG
//	for w32.GetMessage(&msg, 0, 0, 0) != 0 {
//		// TODO this eats VK_ESCAPE and VK_RETURN and makes escape press a
//		// focused button?!
//		if !w32.IsDialogMessage(w.Handle, &msg) {
//			w32.TranslateMessage(&msg)
//			w32.DispatchMessage(&msg)
//		}
//	}
//	return int(msg.WParam), nil // exit code passed to PostQuitMessage
//}
//
//func (w *Window) SetTitle(t string) {
//	w.Title = t
//	if w.Handle != 0 {
//		w32.SetWindowText(w.Handle, w.Title)
//	}
//}
//
//func (w *Window) Close() {
//	w32.SendMessage(w.Handle, w32.WM_CLOSE, 0, 0)
//}
//
//type Font struct {
//	Name       string
//	Height     int
//	Bold       bool
//	Italic     bool
//	Underlined bool
//	StrikedOut bool
//}
//
//func NewFont() *Font {
//	return &Font{}
//}
//
//type Menu struct {
//	name  string
//	items []menuItem
//}
//
//type menuItem interface {
//	isMenuItem()
//}
//
//func NewMenu() *Menu {
//	return &Menu{}
//}
//
//func (m *Menu) AddMenu(name string) *Menu {
//	sub := &Menu{name: name}
//	m.items = append(m.items, sub)
//	return sub
//}
//
//func (m *Menu) AddItem(name string) *MenuItem {
//	item := &MenuItem{name: name}
//	m.items = append(m.items, item)
//	return item
//}
//
//func (m *Menu) AddSeparator() {
//	m.items = append(m.items, &menuSeparator{})
//}
//
//type MenuItem struct {
//	OnClick func()
//	name    string
//}
//
//type menuSeparator struct{}
//
//func (*Menu) isMenuItem()          {}
//func (*MenuItem) isMenuItem()      {}
//func (*menuSeparator) isMenuItem() {}
//
//func NewButton(parent *Window) *Control {
//	b := &Control{
//		Enabled: true,
//		class:   "BUTTON",
//		style:   w32.BS_DEFPUSHBUTTON | w32.WS_TABSTOP,
//	}
//	b.concrete = b
//	parent.Controls = append(parent.Controls, b)
//	return b
//}
//
//type Checkbox struct {
//	Control
//	Checked       bool
//	OnCheckChange func(checked bool)
//}
//
//func NewCheckbox(parent *Window) *Checkbox {
//	c := &Checkbox{
//		Control: Control{
//			Enabled: true,
//			class:   "BUTTON",
//			style:   w32.BS_AUTOCHECKBOX | w32.WS_TABSTOP,
//		},
//	}
//	c.concrete = c
//	parent.Controls = append(parent.Controls, &c.Control)
//	return c
//}
//
//func toCheckState(checked bool) uintptr {
//	if checked {
//		return w32.BST_CHECKED
//	}
//	return w32.BST_UNCHECKED
//}
//
//func (c *Checkbox) SetChecked(checked bool) {
//	if checked != c.Checked {
//		c.Checked = checked
//		w32.SendMessage(c.Handle, w32.BM_SETCHECK, toCheckState(c.Checked), 0)
//		if c.OnCheckChange != nil {
//			c.OnCheckChange(c.Checked)
//		}
//	}
//}
//
