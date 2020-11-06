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
	border   PanelBorderStyle
	font     *Font
}

var _ Control = (*Panel)(nil)
var _ container = (*Panel)(nil)

func (*Panel) canFocus() bool {
	return false
}

func (*Panel) eatsTabs() bool {
	return false
}

type PanelBorderStyle int

const (
	PanelBorderNone PanelBorderStyle = iota
	PanelBorderSingleLine
	PanelBorderSunken
	PanelBorderSunkenThick
	PanelBorderRaised
)

func (s PanelBorderStyle) String() string {
	// NOTE that these strings are used in the designer to get their
	// representations as Go code so they must always correspond to their
	// constant names and be prefixed with the package name.
	switch s {
	case PanelBorderNone:
		return "wui.PanelBorderNone"
	case PanelBorderSingleLine:
		return "wui.PanelBorderSingleLine"
	case PanelBorderSunken:
		return "wui.PanelBorderSunken"
	case PanelBorderSunkenThick:
		return "wui.PanelBorderSunkenThick"
	case PanelBorderRaised:
		return "wui.PanelBorderRaised"
	default:
		return "unknown PanelBorderStyle"
	}
}

func borderStyleEx(b PanelBorderStyle) uint {
	if b == PanelBorderSunken {
		return w32.WS_EX_STATICEDGE
	}
	if b == PanelBorderSunkenThick {
		return w32.WS_EX_CLIENTEDGE
	}
	return 0
}

func borderStyle(b PanelBorderStyle) uint {
	if b == PanelBorderSingleLine {
		return w32.WS_BORDER
	}
	if b == PanelBorderRaised {
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
}

func (p *Panel) getHandle() w32.HWND {
	return p.handle
}

func (p *Panel) getInstance() w32.HINSTANCE {
	return p.parent.getInstance()
}

func (p *Panel) SetBorderStyle(s PanelBorderStyle) {
	p.border = s
	if p.handle != 0 {
		style := uint(w32.GetWindowLongPtr(p.handle, w32.GWL_STYLE))
		style = style &^ w32.WS_BORDER &^ w32.WS_DLGFRAME
		style |= borderStyle(s)
		w32.SetWindowLongPtr(p.handle, w32.GWL_STYLE, uintptr(style))

		exStyle := uint(w32.GetWindowLongPtr(p.handle, w32.GWL_EXSTYLE))
		exStyle = exStyle &^ w32.WS_EX_STATICEDGE &^ w32.WS_EX_CLIENTEDGE
		exStyle |= borderStyleEx(s)
		w32.SetWindowLongPtr(p.handle, w32.GWL_EXSTYLE, uintptr(exStyle))

		w32.InvalidateRect(p.parent.getHandle(), nil, true)
	}
}

func (p *Panel) BorderStyle() PanelBorderStyle {
	return p.border
}

// TODO there is a bug in this when Adding control while parent is not set
func (p *Panel) Add(c Control) {
	c.setParent(p)
	if p.handle != 0 {
		c.create(p.parent.controlCount())
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

func (p *Panel) InnerX() int {
	x, _, _, _ := p.InnerBounds()
	return x
}

func (p *Panel) InnerY() int {
	_, y, _, _ := p.InnerBounds()
	return y
}

func (p *Panel) InnerPosition() (x, y int) {
	x, y, _, _ = p.InnerBounds()
	return
}

func (p *Panel) InnerWidth() int {
	_, _, width, _ := p.InnerBounds()
	return width
}

func (p *Panel) InnerHeight() int {
	_, _, _, height := p.InnerBounds()
	return height
}

func (p *Panel) InnerSize() (width, height int) {
	_, _, width, height = p.InnerBounds()
	return
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

func (p *Panel) SetInnerX(x int) {
	_, y, width, height := p.InnerBounds()
	p.SetInnerBounds(x, y, width, height)
}

func (p *Panel) SetInnerY(y int) {
	x, _, width, height := p.InnerBounds()
	p.SetInnerBounds(x, y, width, height)
}

func (p *Panel) SetInnerPosition(x, y int) {
	_, _, width, height := p.InnerBounds()
	p.SetInnerBounds(x, y, width, height)
}

func (p *Panel) SetInnerWidth(width int) {
	x, y, _, height := p.InnerBounds()
	p.SetInnerBounds(x, y, width, height)
}

func (p *Panel) SetInnerHeight(height int) {
	x, y, width, _ := p.InnerBounds()
	p.SetInnerBounds(x, y, width, height)
}

func (p *Panel) SetInnerSize(width, height int) {
	x, y, _, _ := p.InnerBounds()
	p.SetInnerBounds(x, y, width, height)
}

func (p *Panel) SetInnerBounds(x, y, width, height int) {
	var r w32.RECT
	w32.AdjustWindowRectEx(&r, borderStyle(p.border), false, borderStyleEx(p.border))
	x += int(r.Left)
	y += int(r.Top)
	width += int(r.Width())
	height += int(r.Height())
	p.SetBounds(x, y, width, height)
}

func (p *Panel) SetBounds(x, y, width, height int) {
	_, _, oldW, oldH := p.InnerBounds()
	p.control.SetBounds(x, y, width, height)
	_, _, newW, newH := p.InnerBounds()
	repositionChidrenByAnchors(p, oldW, oldH, newW, newH)
}

// NOTE that we need to re-write all the Set... functions here to make them go
// throught Panel's SetBounds. control's Set... functions go through control's
// SetBounds which does not do what we want.

func (p *Panel) SetX(x int) {
	_, y, width, height := p.Bounds()
	p.SetBounds(x, y, width, height)
}

func (p *Panel) SetY(y int) {
	x, _, width, height := p.Bounds()
	p.SetBounds(x, y, width, height)
}

func (p *Panel) SetPosition(x, y int) {
	_, _, width, height := p.Bounds()
	p.SetBounds(x, y, width, height)
}

func (p *Panel) SetWidth(width int) {
	x, y, _, height := p.Bounds()
	p.SetBounds(x, y, width, height)
}

func (p *Panel) SetHeight(height int) {
	x, y, width, _ := p.Bounds()
	p.SetBounds(x, y, width, height)
}

func (p *Panel) SetSize(width, height int) {
	x, y, _, _ := p.Bounds()
	p.SetBounds(x, y, width, height)
}
