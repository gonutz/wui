//+build windows

package wui

import (
	"syscall"

	"github.com/gonutz/w32"
)

func NewPanel() *Panel {
	return &Panel{}
}

type Panel struct {
	control
	children []Control
	border   panelBorder
	font     *Font
}

type panelBorder int

const (
	borderNone panelBorder = iota
	borderSingleLine
	borderSunken
	borderSunkenThick
	borderRaised
)

func borderStyleEx(b panelBorder) uint {
	if b == borderSunken {
		return w32.WS_EX_STATICEDGE
	}
	if b == borderSunkenThick {
		return w32.WS_EX_CLIENTEDGE
	}
	return 0
}

func borderStyle(b panelBorder) uint {
	if b == borderSingleLine {
		return w32.WS_BORDER
	}
	if b == borderRaised {
		return w32.WS_DLGFRAME
	}
	return 0
}

func (p *Panel) create(id int) {
	p.control.create(id, borderStyleEx(p.border), "STATIC", borderStyle(p.border))
	w32.SetWindowSubclass(p.handle, syscall.NewCallback(func(
		window w32.HWND,
		msg uint32,
		wParam, lParam uintptr,
		subclassID uintptr,
		refData uintptr,
	) uintptr {
		switch msg {
		case w32.WM_COMMAND:
			p.onWM_COMMAND(wParam, lParam)
			return 0
		case w32.WM_DRAWITEM:
			p.onWM_DRAWITEM(wParam, lParam)
			return 0
		case w32.WM_NOTIFY:
			p.onWM_NOTIFY(wParam, lParam)
			return 0
		default:
			return w32.DefSubclassProc(window, msg, wParam, lParam)
		}
	}), 0, 0)
	// TODO Why is this here:
	for i, c := range p.children {
		c.create(id + i + 1)
		p.parent.registerControl(c)
	}
}

func (p *Panel) getHandle() w32.HWND {
	return p.handle
}

func (p *Panel) getInstance() w32.HINSTANCE {
	return p.parent.getInstance()
}

func (p *Panel) setBorder(b panelBorder) {
	p.border = b
	if p.handle != 0 {
		style := uint(w32.GetWindowLongPtr(p.handle, w32.GWL_STYLE))
		style = style &^ w32.WS_BORDER &^ w32.WS_DLGFRAME
		style |= borderStyle(b)
		w32.SetWindowLongPtr(p.handle, w32.GWL_STYLE, uintptr(style))

		exStyle := uint(w32.GetWindowLongPtr(p.handle, w32.GWL_EXSTYLE))
		exStyle = exStyle &^ w32.WS_EX_STATICEDGE &^ w32.WS_EX_CLIENTEDGE
		exStyle |= borderStyleEx(b)
		w32.SetWindowLongPtr(p.handle, w32.GWL_EXSTYLE, uintptr(exStyle))

		w32.InvalidateRect(p.parent.getHandle(), nil, true)
	}
}

func (p *Panel) SetNoBorder() {
	p.setBorder(borderNone)
}

func (p *Panel) SetSingleLineBorder() {
	p.setBorder(borderSingleLine)
}

func (p *Panel) SetSunkenBorder() {
	p.setBorder(borderSunken)
}

func (p *Panel) SetSunkenThickBorder() {
	p.setBorder(borderSunkenThick)
}

func (p *Panel) SetRaisedBorder() {
	p.setBorder(borderRaised)
}

// TODO there is a bug in this when Adding control while parent is not set
func (p *Panel) Add(c Control) {
	c.setParent(p)
	if p.handle != 0 {
		c.create(p.parent.controlCount() + controlIDOffset)
	}
	p.registerControl(c)
	p.children = append(p.children, c)
}

func (p *Panel) Children() []Control {
	return p.children
}

func (p *Panel) onWM_COMMAND(w, l uintptr) {
	p.parent.onWM_COMMAND(w, l)
}

func (p *Panel) onWM_DRAWITEM(w, l uintptr) {
	p.parent.onWM_DRAWITEM(w, l)
}

func (p *Panel) onWM_NOTIFY(w, l uintptr) {
	p.parent.onWM_NOTIFY(w, l)
}

func (p *Panel) controlCount() int {
	return p.parent.controlCount()
}

func (p *Panel) registerControl(c Control) {
	if p.parent != nil {
		p.parent.registerControl(c)
	}
}

func (p *Panel) Font() *Font {
	if p.font == nil && p.parent != nil {
		return p.parent.Font()
	}
	return p.font
}

func (p *Panel) SetFont(f *Font) {
	p.font = f
	for _, c := range p.children {
		c.parentFontChanged()
	}
}

func (p *Panel) InnerBounds() (x, y, width, height int) {
	x, y, width, height = p.Bounds()
	var r w32.RECT
	w32.AdjustWindowRectEx(&r, borderStyle(p.border), false, borderStyleEx(p.border))
	x -= int(r.Left)
	y -= int(r.Top)
	width -= int(r.Width())
	height -= int(r.Height())
	return
}
