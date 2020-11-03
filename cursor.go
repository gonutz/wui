package wui

import (
	"errors"
	"image"

	"github.com/gonutz/w32"
)

// Cursor describes the mouse cursor image. You can use a pre-defined Cursor...
// variable (see below) or create a custom cursor with NewCursorFromImage.
type Cursor struct {
	handle w32.HCURSOR
}

var (
	// CursorArrow is the standard, default arrow cursor.
	CursorArrow = loadCursor(w32.IDC_ARROW)
	// CursorIBeam is the text cursor, it looks like the letter I.
	CursorIBeam = loadCursor(w32.IDC_IBEAM)
	// CursorWait is the hour glass or rotating circle cursor that indicates
	// that an action will take some more time.
	CursorWait = loadCursor(w32.IDC_WAIT)
	// CursorCross looks like a black + (plus) symbol.
	CursorCross = loadCursor(w32.IDC_CROSS)
	// CursorUpArrow is a vertical arrow pointing upwards.
	CursorUpArrow = loadCursor(w32.IDC_UPARROW)
	// CursorSizeNWSE is a diagonal line from top-left to bottom-right with
	// arrows at both ends.
	CursorSizeNWSE = loadCursor(w32.IDC_SIZENWSE)
	// CursorSizeNESW is a diagonal line from top-right to bottom-left with
	// arrows at both ends.
	CursorSizeNESW = loadCursor(w32.IDC_SIZENESW)
	// CursorSizeWE is a horizontal line from left to right with arrows at both
	// ends.
	CursorSizeWE = loadCursor(w32.IDC_SIZEWE)
	// CursorSizeNS is a vertical line from top to bottom with arrows at both
	// ends.
	CursorSizeNS = loadCursor(w32.IDC_SIZENS)
	// CursorSizeALL is a white + (plus) symbol with arrows at all four ends.
	CursorSizeALL = loadCursor(w32.IDC_SIZEALL)
	// CursorNo indicates that an action is not possible. It is a crossed-out
	// red circle, like a stop sign.
	CursorNo = loadCursor(w32.IDC_NO)
	// CursorHand is a hand pointing its index finger upwards.
	CursorHand = loadCursor(w32.IDC_HAND)
	// CursorAppStarting is a combination of CursorArrow and CursorWait, it has
	// the arrow cursor but with an hour glass or rotating circle next to it.
	CursorAppStarting = loadCursor(w32.IDC_APPSTARTING)
	// CursorHelp is the default arrow cursor with a little question mark icon
	// next to it.
	CursorHelp = loadCursor(w32.IDC_HELP)
)

func loadCursor(id uint16) *Cursor {
	return &Cursor{handle: w32.LoadCursor(0, w32.MakeIntResource(id))}
}

// NewCursorFromImage creates a cursor with 4 possible colors: black, white,
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
func NewCursorFromImage(img image.Image, x, y int) (*Cursor, error) {
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
