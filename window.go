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

type Control interface {
	isControl()
	create(parent *Window, id int, instance w32.HINSTANCE)
	parentFontChanged()
}

const controlIDOffset = 2

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

func (*Window) isContainer() {}

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
		for _, c := range w.controls {
			c.parentFontChanged()
		}
	}
	return w
}

func (w *Window) Add(c Control) *Window {
	w.controls = append(w.controls, c)
	if w.handle != 0 {
		id := len(w.controls) - 1 + controlIDOffset
		instance := w32.HINSTANCE(w32.GetWindowLong(w.handle, w32.GWL_HINSTANCE))
		c.create(w, id, instance)
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

	runtime.LockOSThread()

	hideConsoleWindow()

	setManifest()

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
						case *Checkbox:
							state := w32.IsDlgButtonChecked(window, id)
							checked := state == w32.BST_CHECKED
							if c.checked != checked {
								c.checked = checked
								if c.onChange != nil {
									c.onChange(checked)
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
