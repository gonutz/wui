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

	"github.com/gonutz/w32"
)

type Control interface {
	isControl()
}

type Window struct {
	handle      w32.HWND
	className   string
	classStyle  uint32
	title       string
	x           int
	y           int
	width       int
	height      int
	style       uint
	state       int
	background  w32.HBRUSH
	cursor      w32.HCURSOR
	menu        *Menu
	font        *Font
	controls    []Control
	onMouseMove func(x, y int)
	onKeyDown   func(key int)
	onKeyUp     func(key int)
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

func (w *Window) X() int      { return w.x }
func (w *Window) Y() int      { return w.y }
func (w *Window) Width() int  { return w.width }
func (w *Window) Height() int { return w.height }

func (w *Window) SetClassName(name string) *Window {
	if w.handle != 0 {
		w.className = name
	}
	return w
}

func (w *Window) SetTitle(title string) *Window {
	w.title = title
	if w.handle != 0 {
		w32.SetWindowText(w.handle, title)
	}
	return w
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

func (w *Window) SetFont(f *Font) *Window {
	w.font = f
	if f == nil {
		return w
	}
	if w.handle != 0 {
		f.create()
		for _, control := range w.controls {
			switch c := control.(type) {
			case *Button:
				w32.SendMessage(
					c.handle,
					w32.WM_SETFONT,
					uintptr(w.font.handle),
					1,
				)
			case *NumberUpDown:
				w32.SendMessage(
					c.editHandle,
					w32.WM_SETFONT,
					uintptr(w.font.handle),
					1,
				)
			default:
				panic("unhandled control type")
			}
		}
	}
	return w
}

func (w *Window) Maximize() *Window {
	w.state = w32.SW_MAXIMIZE
	if w.handle != 0 {
		w32.ShowWindow(w.handle, w.state)
	}
	return w
}

func (w *Window) Minimize() *Window {
	w.state = w32.SW_MINIMIZE
	if w.handle != 0 {
		w32.ShowWindow(w.handle, w.state)
	}
	return w
}

func (w *Window) Restore() *Window {
	w.state = w32.SW_SHOWNORMAL
	if w.handle != 0 {
		w32.ShowWindow(w.handle, w.state)
	}
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

const controlIDOffset = 2

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
					// TODO menu item clicked
					/*id := int(wParam & 0xFFFF)
					if 0 <= id && id < len(w.menuItems) {
						f := w.menuItems[id].OnClick
						if f != nil {
							f()
						}
					}*/
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
			/*case w32.WM_PAINT:
			if w.OnPaint != nil {
				func() {
					var ps w32.PAINTSTRUCT
					hdc := w32.BeginPaint(window, &ps)
					defer w32.EndPaint(window, &ps)
					var b w32.LOGBRUSH
					w32.GetObject(
						w32.HGDIOBJ(w.Background),
						unsafe.Sizeof(b),
						unsafe.Pointer(&b),
					)
					w32.SetBkColor(hdc, b.LbColor)
					if w.font != 0 {
						w32.SelectObject(hdc, w32.HGDIOBJ(w.font))
					}
					w.OnPaint(Painter(hdc))
				}()
				return 0
			}*/
			case w32.WM_DESTROY:
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
	r := w32.GetWindowRect(w.handle)
	w.x = int(r.Left)
	w.y = int(r.Top)

	if w.font != nil {
		w.font.create()
	}

	//if w.Menu != nil {
	//	var addItems func(m w32.HMENU, items []menuItem)
	//	addItems = func(m w32.HMENU, items []menuItem) {
	//		for _, item := range items {
	//			switch menuItem := item.(type) {
	//			case *Menu:
	//				menu := w32.CreateMenu()
	//				w32.AppendMenu(m, w32.MF_POPUP, uintptr(menu),
	// menuItem.name)
	//				addItems(menu, menuItem.items)
	//			case *MenuItem:
	//				w32.AppendMenu(
	//					m,
	//					w32.MF_STRING,
	//					uintptr(len(w.menuItems)),
	//					menuItem.name,
	//				)
	//				w.menuItems = append(w.menuItems, menuItem)
	//			case *menuSeparator:
	//				w32.AppendMenu(m, w32.MF_SEPARATOR, 0, "")
	//			}
	//		}
	//	}
	//	menuBar := w32.CreateMenu()
	//	addItems(menuBar, w.Menu.items)
	//	w32.SetMenu(window, menuBar)
	//}

	instance := w32.HINSTANCE(w32.GetWindowLong(window, w32.GWL_HINSTANCE))
	for i, c := range w.controls {
		createControl(c, w, i+controlIDOffset, instance)
	}

	w32.ShowWindow(window, w.state)

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
	default:
		panic("unhandled control type")
	}
}

type Menu struct {
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
	handle  w32.HWND
	text    string
	x       int
	y       int
	width   int
	height  int
	onClick func()
}

func (Button) isControl() {}

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
	onValueChange func(value int)
}

func (NumberUpDown) isControl() {}

func NewNumberUpDown() *NumberUpDown {
	return &NumberUpDown{
		minValue: math.MinInt32,
		maxValue: math.MaxInt32,
	}
}

func (n *NumberUpDown) Value() int32 {
	if n.upDownHandle != 0 {
		n.value = int32(w32.SendMessage(n.upDownHandle, w32.UDM_GETPOS32, 0, 0))
	}
	return n.value
}

func (n *NumberUpDown) SetValue(v int32) *NumberUpDown {
	n.value = v
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

func (n *NumberUpDown) SetMinValue(min int32) *NumberUpDown {
	if n.Value() < min {
		n.SetValue(min)
	}
	n.minValue = min
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

func (n *NumberUpDown) SetMaxValue(max int32) *NumberUpDown {
	if n.Value() > max {
		n.SetValue(max)
	}
	n.maxValue = max
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

func (n *NumberUpDown) SetMinMaxValues(min, max int32) *NumberUpDown {
	if n.Value() < min {
		n.SetValue(min)
	} else if n.Value() > max {
		n.SetValue(max)
	}
	n.minValue = min
	n.maxValue = max
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
