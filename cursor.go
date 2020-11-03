package wui

import (
	"errors"
	"image"

	"github.com/gonutz/w32"
)

type Cursor struct {
	handle w32.HCURSOR
}

var (
	CursorArrow       = loadCursor(w32.IDC_ARROW)
	CursorIBeam       = loadCursor(w32.IDC_IBEAM)
	CursorWait        = loadCursor(w32.IDC_WAIT)
	CursorCross       = loadCursor(w32.IDC_CROSS)
	CursorUpArrow     = loadCursor(w32.IDC_UPARROW)
	CursorSize        = loadCursor(w32.IDC_SIZE)
	CursorSizeNWSE    = loadCursor(w32.IDC_SIZENWSE)
	CursorSizeNESW    = loadCursor(w32.IDC_SIZENESW)
	CursorSizeWE      = loadCursor(w32.IDC_SIZEWE)
	CursorSizeNS      = loadCursor(w32.IDC_SIZENS)
	CursorSizeALL     = loadCursor(w32.IDC_SIZEALL)
	CursorNo          = loadCursor(w32.IDC_NO)
	CursorHand        = loadCursor(w32.IDC_HAND)
	CursorAppStarting = loadCursor(w32.IDC_APPSTARTING)
	CursorHelp        = loadCursor(w32.IDC_HELP)
	CursorIcon        = loadCursor(w32.IDC_ICON)
)

func loadCursor(id uint16) *Cursor {
	return &Cursor{handle: w32.LoadCursor(0, w32.MakeIntResource(id))}
}

func CursorFromImage(img image.Image, x, y int) (*Cursor, error) {
	b := img.Bounds()
	bits := make([]byte, b.Dx()*b.Dy()*2)
	and := bits[:len(bits)/2]
	xor := bits[len(bits)/2:]
	// color.Color.RGBA() returns 32 bit numbers that really use only 16 bits,
	// color intensities range from 0 to 0xFFFF. We consider a color on if it is
	// at least half of that.
	const halfIntensity = 0x7FFF
	count := 0
	var mask byte = 1
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			i := count / 8
			count++
			mask = mask>>1 | mask<<7
			if a < halfIntensity {
				// Transparent -> AND=1 XOR=0.
				and[i] |= mask
			} else if r+g+b >= 3*halfIntensity {
				// White -> AND=0 XOR=1.
				xor[i] |= mask
			}
			// Otherwise we assume black which has AND=0 XOR=0 so we do nothing.
		}
	}
	handle := w32.CreateCursor(w32.GetModuleHandle(""), x, y, b.Dx(), b.Dy(), and, xor)
	if handle == 0 {
		return nil, errors.New("wui.CursorFromImage: CreateCursor returned 0 handle")
	}
	return &Cursor{handle: handle}, nil
}
