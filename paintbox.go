//+build windows

package wui

import (
	"github.com/gonutz/w32"

	"math"
)

func NewPaintbox() *Paintbox {
	return &Paintbox{}
}

type Paintbox struct {
	control
	onPaint func(*Canvas)
}

func (p *Paintbox) create(id int) {
	p.control.create(id, 0, "STATIC", w32.SS_OWNERDRAW)
}

func (p *Paintbox) SetOnPaint(f func(*Canvas)) {
	p.onPaint = f
}

func (p *Paintbox) Paint() {
	if p.handle != 0 {
		w32.InvalidateRect(p.handle, nil, true)
	}
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

func (c *Canvas) SetFont(font *Font) {
	w32.SelectObject(c.hdc, w32.HGDIOBJ(font.handle))
}
