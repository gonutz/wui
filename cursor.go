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

// CursorFromImage creates a cursor with 4 possible colors: black, white,
// transparent and inverse screen color. Inverse screen color depends on what is
// under the cursor at any time. The cursor will be white minus the screen
// color to give maximum contrast for such pixels.
//
// Fully opaque black and white pixels in the image are interpreted as black and
// white.
//
// Fully transparent pixels in the image are interpreted as screen color.
//
// All other pixels are interpreted as inverse screen color. This means even
// "almost" black/white/transparent, say with intensity 0xFFFE instead of
// 0xFFFF, will be interpreted as inverse screen color.
func CursorFromImage(img image.Image, x, y int) (*Cursor, error) {
	// Cursor images have two bit masks: AND and XOR. Combining the bits in both
	// will yield these results:
	//
	// AND  XOR  Result
	//  0    0   Black
	//  0    1   White
	//  1    0   Screen color (transparent)
	//  1    1   Inverse screen color
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
			if a == 0 {
				// Alpha is 0 -> use screen color.
				and[i] |= mask
			} else if r == 0 && g == 0 && b == 0 && a == 0xFFFF {
				// Black pixel. Nothing to change here, 0 0 is the default.
			} else if r == 0xFFFF && g == 0xFFFF && b == 0xFFFF && a == 0xFFFF {
				// White pixel.
				xor[i] |= mask
			} else {
				// Everything else is inverse screen color.
				and[i] |= mask
				xor[i] |= mask
			}
		}
	}
	handle := w32.CreateCursor(w32.GetModuleHandle(""), x, y, b.Dx(), b.Dy(), and, xor)
	if handle == 0 {
		return nil, errors.New("wui.CursorFromImage: CreateCursor returned 0 handle")
	}
	return &Cursor{handle: handle}, nil
}
