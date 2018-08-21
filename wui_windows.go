package wui

import (
	"errors"
	"io/ioutil"
	"os"
	"runtime"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/gonutz/w32"
)

func NewWindow() *Window {
	return &Window{
		X:           w32.CW_USEDEFAULT,
		Y:           w32.CW_USEDEFAULT,
		Width:       w32.CW_USEDEFAULT,
		Height:      w32.CW_USEDEFAULT,
		Cursor:      w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_ARROW)),
		Style:       w32.WS_OVERLAPPEDWINDOW,
		ShowCommand: w32.SW_SHOWNORMAL,
	}
}

type Window struct {
	Handle     w32.HWND
	Title      string
	Background w32.HBRUSH
	Cursor     w32.HCURSOR
	ClassName  string
	ClassStyle uint32
	Style      uint
	X          int
	Y          int
	// Width is the outer window width if it is >= 0.
	// If Width is negative, it is the negative client area width.
	Width int
	// Height is the outer window height if it is >= 0.
	// If Height is negative, it is the negative client area height.
	Height      int
	ShowCommand int
	Font        *Font
	Menu        *Menu
	Controls    []*Control
	OnKeyDown   func(key int)
	OnKeyUp     func(key int)
	OnMouseMove func(x, y int)
	OnPaint     func(p Painter)
	font        w32.HFONT
	menuItems   []*MenuItem
}

type Painter w32.HDC

func (p Painter) TextOut(x, y int, text string) {
	w32.TextOut(w32.HDC(p), x, y, text)
}

type Control struct {
	Handle   w32.HWND
	Text     string
	X        int
	Y        int
	Width    int
	Height   int
	Enabled  bool
	OnClick  func()
	class    string
	style    uint
	concrete interface{}
}

func (c *Control) SetEnabled(e bool) {
	c.Enabled = e
	if c.Handle != 0 {
		w32.EnableWindow(c.Handle, e)
	}
}

func (c *Control) SetPos(x, y, width, height int) {
	c.X, c.Y, c.Width, c.Height = x, y, width, height
	if c.Handle != 0 {
		w32.SetWindowPos(c.Handle, 0, x, y, width, height, w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER)
	}
}

const controlIDOffset = 2

func (w *Window) Show() (int, error) {
	runtime.LockOSThread()

	setManifest()
	w32.InitCommonControlsEx(&w32.INITCOMMONCONTROLSEX{ICC: w32.ICC_UPDOWN_CLASS})

	if w.ClassName == "" {
		w.ClassName = "wui_window_class"
	}
	var clientRect w32.RECT
	w32.AdjustWindowRect(&clientRect, w.Style, w.Menu != nil)
	if w.Width < 0 && w.Width != w32.CW_USEDEFAULT {
		w.Width = -w.Width + int(clientRect.Width())
	}
	if w.Height < 0 && w.Height != w32.CW_USEDEFAULT {
		w.Height = -w.Height + int(clientRect.Height())
	}
	if w.Background == 0 {
		w.Background = w32.GetSysColorBrush(w32.COLOR_BTNFACE)
	}

	class := w32.WNDCLASSEX{
		Background: w.Background,
		WndProc: syscall.NewCallback(func(
			window w32.HWND,
			msg uint32,
			wParam, lParam uintptr,
		) uintptr {
			switch msg {
			case w32.WM_MOUSEMOVE:
				if w.OnMouseMove != nil {
					w.OnMouseMove(
						int(lParam&0xFFFF),
						int(lParam&0xFFFF0000)>>16,
					)
					return 0
				}
			case w32.WM_KEYDOWN:
				if w.OnKeyDown != nil {
					w.OnKeyDown(int(wParam))
					return 0
				}
			case w32.WM_KEYUP:
				if w.OnKeyUp != nil {
					w.OnKeyUp(int(wParam))
					return 0
				}
			case w32.WM_COMMAND:
				if lParam == 0 && wParam&0xFFFF0000 == 0 {
					// menu item clicked
					id := int(wParam & 0xFFFF)
					if 0 <= id && id < len(w.menuItems) {
						f := w.menuItems[id].OnClick
						if f != nil {
							f()
						}
					}
				} else if lParam != 0 {
					// control clicked
					id := wParam & 0xFFFF
					index := id - controlIDOffset
					if 0 <= index && index < uintptr(len(w.Controls)) {
						c := w.Controls[index]
						if c.OnClick != nil {
							c.OnClick()
						}
						if check, ok := c.concrete.(*Checkbox); ok {
							checked := w32.IsDlgButtonChecked(w.Handle, id) == w32.BST_CHECKED
							if checked != check.Checked {
								check.Checked = checked
								if check.OnCheckChange != nil {
									check.OnCheckChange(check.Checked)
								}
							}
						}
						return 0
					}
				}
			case w32.WM_PAINT:
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
				}
			case w32.WM_DESTROY:
				w32.PostQuitMessage(0)
				return 0
			}
			return w32.DefWindowProc(window, msg, wParam, lParam)
		}),
		Cursor:    w.Cursor,
		ClassName: syscall.StringToUTF16Ptr(w.ClassName),
		Style:     w.ClassStyle,
	}
	atom := w32.RegisterClassEx(&class)
	if atom == 0 {
		return 0, errors.New("win.NewWindow: RegisterClassEx failed")
	}
	window := w32.CreateWindowEx(
		0,
		syscall.StringToUTF16Ptr(w.ClassName),
		syscall.StringToUTF16Ptr(w.Title),
		w.Style,
		w.X, w.Y, w.Width, w.Height,
		0, 0, 0, nil,
	)
	if window == 0 {
		return 0, errors.New("win.NewWindow: CreateWindowEx failed")
	}
	w.Handle = window

	if w.Font != nil {
		height := int32(w.Font.Height)
		weight := int32(w32.FW_NORMAL)
		if w.Font.Bold {
			weight = w32.FW_BOLD
		}
		var italic byte
		if w.Font.Italic {
			italic = 1
		}
		var underlined byte
		if w.Font.Underlined {
			underlined = 1
		}
		var strikedOut byte
		if w.Font.StrikedOut {
			strikedOut = 1
		}
		f := w32.LOGFONT{
			Height:         height,
			Width:          0,
			Escapement:     0,
			Orientation:    0,
			Weight:         weight,
			Italic:         italic,
			Underline:      underlined,
			StrikeOut:      strikedOut,
			CharSet:        w32.DEFAULT_CHARSET,
			OutPrecision:   w32.OUT_CHARACTER_PRECIS,
			ClipPrecision:  w32.CLIP_CHARACTER_PRECIS,
			Quality:        w32.DEFAULT_QUALITY,
			PitchAndFamily: w32.DEFAULT_PITCH | w32.FF_DONTCARE,
		}
		copy(f.FaceName[:], utf16.Encode([]rune(w.Font.Name)))
		w.font = w32.CreateFontIndirect(&f)
		defer w32.DeleteObject(w32.HGDIOBJ(w.font))
	}

	if w.Menu != nil {
		var addItems func(m w32.HMENU, items []menuItem)
		addItems = func(m w32.HMENU, items []menuItem) {
			for _, item := range items {
				switch menuItem := item.(type) {
				case *Menu:
					menu := w32.CreateMenu()
					w32.AppendMenu(m, w32.MF_POPUP, uintptr(menu), menuItem.name)
					addItems(menu, menuItem.items)
				case *MenuItem:
					w32.AppendMenu(
						m,
						w32.MF_STRING,
						uintptr(len(w.menuItems)),
						menuItem.name,
					)
					w.menuItems = append(w.menuItems, menuItem)
				case *menuSeparator:
					w32.AppendMenu(m, w32.MF_SEPARATOR, 0, "")
				}
			}
		}
		menuBar := w32.CreateMenu()
		addItems(menuBar, w.Menu.items)
		w32.SetMenu(window, menuBar)
	}

	instance := w32.HINSTANCE(w32.GetWindowLong(window, w32.GWL_HINSTANCE))
	focussed := false
	for i, c := range w.Controls {
		c.Handle = w32.CreateWindowExStr(
			0,
			c.class,
			c.Text,
			w32.WS_VISIBLE|w32.WS_CHILD|c.style,
			c.X, c.Y, c.Width, c.Height,
			w.Handle,
			w32.HMENU(i+controlIDOffset),
			instance,
			nil,
		)
		if !focussed && c.Enabled && c.style&w32.WS_TABSTOP != 0 {
			focussed = true
			w32.SetFocus(c.Handle)
		}
		if w.font != 0 {
			w32.SendMessage(c.Handle, w32.WM_SETFONT, uintptr(w.font), 1)
		}
		w32.EnableWindow(c.Handle, c.Enabled)
		if check, ok := c.concrete.(*Checkbox); ok {
			w32.SendMessage(
				c.Handle,
				w32.BM_SETCHECK,
				toCheckState(check.Checked),
				0,
			)
		}
	}

	w32.ShowWindow(window, w.ShowCommand)

	var msg w32.MSG
	for w32.GetMessage(&msg, 0, 0, 0) != 0 {
		// TODO this eats VK_ESCAPE and VK_RETURN and makes escape press a focused button?!
		if !w32.IsDialogMessage(w.Handle, &msg) {
			w32.TranslateMessage(&msg)
			w32.DispatchMessage(&msg)
		}
	}
	return int(msg.WParam), nil // exit code passed to PostQuitMessage
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

func (w *Window) SetTitle(t string) {
	w.Title = t
	if w.Handle != 0 {
		w32.SetWindowText(w.Handle, w.Title)
	}
}

func (w *Window) Close() {
	w32.SendMessage(w.Handle, w32.WM_CLOSE, 0, 0)
}

type Font struct {
	Name       string
	Height     int
	Bold       bool
	Italic     bool
	Underlined bool
	StrikedOut bool
}

func NewFont() *Font {
	return &Font{}
}

type Menu struct {
	name  string
	items []menuItem
}

type menuItem interface {
	isMenuItem()
}

func NewMenu() *Menu {
	return &Menu{}
}

func (m *Menu) AddMenu(name string) *Menu {
	sub := &Menu{name: name}
	m.items = append(m.items, sub)
	return sub
}

func (m *Menu) AddItem(name string) *MenuItem {
	item := &MenuItem{name: name}
	m.items = append(m.items, item)
	return item
}

func (m *Menu) AddSeparator() {
	m.items = append(m.items, &menuSeparator{})
}

type MenuItem struct {
	OnClick func()
	name    string
}

type menuSeparator struct{}

func (*Menu) isMenuItem()          {}
func (*MenuItem) isMenuItem()      {}
func (*menuSeparator) isMenuItem() {}

func NewButton(parent *Window) *Control {
	b := &Control{
		Enabled: true,
		class:   "BUTTON",
		style:   w32.BS_DEFPUSHBUTTON | w32.WS_TABSTOP,
	}
	b.concrete = b
	parent.Controls = append(parent.Controls, b)
	return b
}

type Checkbox struct {
	Control
	Checked       bool
	OnCheckChange func(checked bool)
}

func NewCheckbox(parent *Window) *Checkbox {
	c := &Checkbox{
		Control: Control{
			Enabled: true,
			class:   "BUTTON",
			style:   w32.BS_AUTOCHECKBOX | w32.WS_TABSTOP,
		},
	}
	c.concrete = c
	parent.Controls = append(parent.Controls, &c.Control)
	return c
}

func toCheckState(checked bool) uintptr {
	if checked {
		return w32.BST_CHECKED
	}
	return w32.BST_UNCHECKED
}

func (c *Checkbox) SetChecked(checked bool) {
	if checked != c.Checked {
		c.Checked = checked
		w32.SendMessage(c.Handle, w32.BM_SETCHECK, toCheckState(c.Checked), 0)
		if c.OnCheckChange != nil {
			c.OnCheckChange(c.Checked)
		}
	}
}
