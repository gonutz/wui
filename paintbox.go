//go:build windows
// +build windows

package wui

import (
	"image"
	"image/draw"
	"math"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

func NewPaintBox() *PaintBox {
	return &PaintBox{}
}

type PaintBox struct {
	control
	backBuffer  backBuffer
	onPaint     func(*Canvas)
	onMouseMove func(x, y int)
}

var _ Control = (*PaintBox)(nil)

func (*PaintBox) canFocus() bool {
	return false
}

func (*PaintBox) eatsTabs() bool {
	return false
}

type backBuffer struct {
	w, h int
	dc   w32.HDC
	bmp  w32.HBITMAP
}

func (b *backBuffer) setMinSize(hdc w32.HDC, w, h int) {
	if w > b.w || h > b.h {
		if b.dc != 0 {
			w32.DeleteObject(w32.HGDIOBJ(b.bmp))
			w32.DeleteDC(b.dc)
		}

		b.dc = w32.CreateCompatibleDC(hdc)
		b.bmp = w32.CreateCompatibleBitmap(hdc, w, h)
		b.w = w
		b.h = h
	}
}

func (p *PaintBox) create(id int) {
	p.control.create(id, 0, "STATIC", w32.SS_OWNERDRAW)
	w32.SetWindowSubclass(p.handle, syscall.NewCallback(func(
		window w32.HWND,
		msg uint32,
		wParam, lParam uintptr,
		subclassID uintptr,
		refData uintptr,
	) uintptr {
		switch msg {
		case w32.WM_NCHITTEST:
			if p.onMouseMove != nil {
				x := int(int16(lParam & 0xFFFF))
				y := int(int16((lParam & 0xFFFF0000) >> 16))
				x, y, _ = w32.ScreenToClient(p.handle, x, y)
				p.onMouseMove(x, y)
			}
		}
		return w32.DefSubclassProc(window, msg, wParam, lParam)
	}), 0, 0)
}

func (p *PaintBox) OnMouseMove() func(x, y int) {
	return p.onMouseMove
}

func (p *PaintBox) SetOnMouseMove(f func(x, y int)) {
	p.onMouseMove = f
}

func (p *PaintBox) SetOnPaint(f func(*Canvas)) {
	p.onPaint = f
}

func (p *PaintBox) OnPaint() func(*Canvas) {
	return p.onPaint
}

func (p *PaintBox) Paint() {
	if p.handle != 0 {
		w32.InvalidateRect(p.handle, nil, true)
	}
}

type Canvas struct {
	hdc     w32.HDC
	width   int
	height  int
	regions []w32.HRGN
}

// Handle returns the handle to the canvas' device context (HDC).
func (c *Canvas) Handle() uintptr {
	return uintptr(c.hdc)
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

func (c *Canvas) PushDrawRegion(x, y, width, height int) {
	r := w32.CreateRectRgn(x, y, x+width, y+height)
	if len(c.regions) > 0 {
		w32.CombineRgn(r, r, c.regions[len(c.regions)-1], w32.RGN_AND)
	}
	c.regions = append(c.regions, r)
	w32.SelectClipRgn(c.hdc, r)
}

func (c *Canvas) PopDrawRegion() {
	n := len(c.regions)
	if n == 0 {
		return
	}
	if n == 1 {
		w32.SelectClipRgn(c.hdc, 0)
	} else {
		w32.SelectClipRgn(c.hdc, c.regions[n-2])
	}
	w32.DeleteObject(w32.HGDIOBJ(c.regions[n-1]))
	c.regions = c.regions[:n-1]
}

func (c *Canvas) ClearDrawRegions() {
	w32.SelectClipRgn(c.hdc, 0)
	for _, r := range c.regions {
		w32.DeleteObject(w32.HGDIOBJ(r))
	}
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

type Point struct {
	X, Y int32
}

func (c *Canvas) Polyline(p []Point, color Color) {
	if len(p) < 2 {
		return
	}
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.NULL_BRUSH))
	w32.PolylineMem(c.hdc, unsafe.Pointer(&p[0]), len(p))
}

func (c *Canvas) Polygon(p []Point, color Color) {
	if len(p) < 2 {
		return
	}
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_PEN))
	w32.SetDCPenColor(c.hdc, w32.COLORREF(color))
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.DC_BRUSH))
	w32.SetDCBrushColor(c.hdc, w32.COLORREF(color))
	w32.PolygonMem(c.hdc, unsafe.Pointer(&p[0]), len(p))
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

// TODO What about line breaks in TextRects (\n vs \r\n)?

func (c *Canvas) TextRect(x, y, w, h int, s string, color Color) {
	c.TextRectFormat(x, y, w, h, s, FormatTopLeft, color)
}

// TextRectExtent returns the size of the text when drawn in a rectangle of the
// given width. The given width is necessary because text rects use word breaks
// and thus given a smaller width might produce a higher text height.
func (c *Canvas) TextRectExtent(s string, givenWidth int) (width, height int) {
	// TODO What is the max of givenWidth, int can be larger than 0x7FFFFFFF (or
	// whatever it is) so do we clamp it or have a const NoWidth=0x7FFFFFFF?
	var flags uint = w32.DT_WORDBREAK | w32.DT_NOFULLWIDTHCHARBREAK | w32.DT_EXPANDTABS
	var r w32.RECT
	const maxInt32 = 0x7FFFFFFF
	if givenWidth > maxInt32 {
		givenWidth = maxInt32
	}
	r.Right = int32(givenWidth)
	w32.DrawText(c.hdc, s, &r, flags|w32.DT_CALCRECT)
	return int(r.Width()), int(r.Height())
}

type Format int

const (
	FormatTopLeft Format = iota
	FormatCenterLeft
	FormatBottomLeft
	FormatTopCenter
	FormatCenter
	FormatBottomCenter
	FormatTopRight
	FormatCenterRight
	FormatBottomRight
)

func (c *Canvas) TextRectFormat(x, y, w, h int, s string, format Format, color Color) {
	w32.SetBkMode(c.hdc, w32.TRANSPARENT)
	w32.SelectObject(c.hdc, w32.GetStockObject(w32.NULL_BRUSH))
	w32.SetTextColor(c.hdc, w32.COLORREF(color))
	r := w32.RECT{
		Left:   int32(x),
		Top:    int32(y),
		Right:  int32(x + w),
		Bottom: int32(y + h),
	}
	var flags uint = w32.DT_WORDBREAK | w32.DT_NOFULLWIDTHCHARBREAK | w32.DT_EXPANDTABS
	// add the appropriate horizontal positioning flag
	switch format {
	default:
		flags |= w32.DT_LEFT
	case FormatTopCenter, FormatCenter, FormatBottomCenter:
		flags |= w32.DT_CENTER
	case FormatTopRight, FormatCenterRight, FormatBottomRight:
		flags |= w32.DT_RIGHT
	}
	// w32.DrawText will only respect w32.DT_VCENTER and w32.DT_BOTTOM if the
	// single-line option is also set, this means that we actually have to do
	// the work of positioning the text vertically ourselves
	switch format {
	default:
		w32.DrawText(c.hdc, s, &r, flags)
	case FormatCenterLeft, FormatCenter, FormatCenterRight:
		calc := r
		w32.DrawText(c.hdc, s, &calc, flags|w32.DT_CALCRECT)
		if calc.Height() < r.Height() {
			r.Top += (r.Height() - calc.Height()) / 2
		}
		w32.DrawText(c.hdc, s, &r, flags)
	case FormatBottomLeft, FormatBottomCenter, FormatBottomRight:
		calc := r
		w32.DrawText(c.hdc, s, &calc, flags|w32.DT_CALCRECT)
		if calc.Height() < r.Height() {
			r.Top += r.Height() - calc.Height()
		}
		w32.DrawText(c.hdc, s, &r, flags)
	}
	w32.SetBkMode(c.hdc, w32.OPAQUE)
}

func (c *Canvas) SetFont(font *Font) {
	if font != nil {
		w32.SelectObject(c.hdc, w32.HGDIOBJ(font.handle))
	}
}

func (c *Canvas) DrawImage(img *Image, src Rectangle, destX, destY int) {
	if src.Width == 0 {
		src.Width = img.width
	}
	if src.Height == 0 {
		src.Height = img.height
	}

	hdcMem := w32.CreateCompatibleDC(c.hdc)
	old := w32.SelectObject(hdcMem, w32.HGDIOBJ(img.bitmap))

	w32.AlphaBlend(
		c.hdc,
		destX, destY, src.Width, src.Height,
		hdcMem,
		src.X, src.Y, src.Width, src.Height,
		w32.BLENDFUNC{
			BlendOp:             w32.AC_SRC_OVER,
			BlendFlags:          0,
			SourceConstantAlpha: 255,
			AlphaFormat:         w32.AC_SRC_ALPHA,
		},
	)

	w32.SelectObject(hdcMem, old)
	w32.DeleteDC(hdcMem)
}

// NewImageFromHBITMAP takes a handle to a bitmap (HBITMAP) and makes it an
// Image that you can use in Canvas.DrawImage.
func NewImageFromHBITMAP(bitmap uintptr, width, height int) *Image {
	return &Image{
		bitmap: w32.HBITMAP(bitmap),
		width:  width,
		height: height,
	}
}

func NewImage(img image.Image) *Image {
	var bmp w32.BITMAPINFO
	bmp.BmiHeader.BiSize = uint32(unsafe.Sizeof(bmp.BmiHeader))
	bmp.BmiHeader.BiWidth = int32(img.Bounds().Dx())
	bmp.BmiHeader.BiHeight = -int32(img.Bounds().Dy())
	bmp.BmiHeader.BiPlanes = 1
	bmp.BmiHeader.BiBitCount = 32
	bmp.BmiHeader.BiCompression = w32.BI_RGB

	var bits unsafe.Pointer
	bitmap := w32.CreateDIBSection(0, &bmp, 0, &bits, 0, 0)
	rgba := toRGBA(img)
	pixels := rgba.Pix
	var dest []byte
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&dest))
	hdrp.Data = uintptr(bits)
	hdrp.Len = len(pixels)
	hdrp.Cap = hdrp.Len
	// swap red and blue because we need BGR and not RGB on Windows
	for i := 0; i < len(pixels); i += 4 {
		dest[i+0] = pixels[i+2]
		dest[i+1] = pixels[i+1]
		dest[i+2] = pixels[i+0]
		dest[i+3] = pixels[i+3]
	}
	return &Image{
		bitmap: bitmap,
		width:  img.Bounds().Dx(),
		height: img.Bounds().Dy(),
	}
}

func toRGBA(img image.Image) *image.RGBA {
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, rgba.Bounds().Min, draw.Src)
	return rgba
}

type Image struct {
	bitmap w32.HBITMAP
	width  int
	height int
}

func (img *Image) Width() int {
	return img.width
}

func (img *Image) Height() int {
	return img.height
}

func (img *Image) Size() (w, h int) {
	return img.width, img.height
}

func (img *Image) Bounds() Rectangle {
	return Rect(0, 0, img.width, img.height)
}

func Rect(x, y, width, height int) Rectangle {
	return Rectangle{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
	}
}

type Rectangle struct {
	X, Y, Width, Height int
}
