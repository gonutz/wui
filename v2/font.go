//+build windows

package wui

import (
	"errors"

	"github.com/gonutz/w32/v2"
)

var NoExactFontMatch = errors.New("wui.NewFont: the desired font was not found in the system, a replacement is used")

// NewFont returns a font according to the given description and an error. The
// error might be NoExactFontMatch in which case the returned Font is valid, but
// the system did not find an exact match. In case the creation fails, the
// returned Font is nil and the error gives the reason.
func NewFont(desc FontDesc) (*Font, error) {
	var weight int32 = w32.FW_NORMAL
	if desc.Bold {
		weight = w32.FW_BOLD
	}
	byteBool := func(b bool) byte {
		if b {
			return 1
		}
		return 0
	}
	logfont := w32.LOGFONT{
		Height:         int32(desc.Height),
		Width:          0,
		Escapement:     0,
		Orientation:    0,
		Weight:         weight,
		Italic:         byteBool(desc.Italic),
		Underline:      byteBool(desc.Underlined),
		StrikeOut:      byteBool(desc.StrikedOut),
		CharSet:        w32.DEFAULT_CHARSET,
		OutPrecision:   w32.OUT_CHARACTER_PRECIS,
		ClipPrecision:  w32.CLIP_CHARACTER_PRECIS,
		Quality:        w32.DEFAULT_QUALITY,
		PitchAndFamily: w32.DEFAULT_PITCH | w32.FF_DONTCARE,
	}
	logfont.SetFaceName(desc.Name)

	found := false
	w32.EnumFontFamiliesEx(w32.GetDC(0), logfont, func(*w32.ENUMLOGFONTEX, *w32.ENUMTEXTMETRIC, w32.FontType) bool {
		found = true
		return false
	})
	var err error
	if !found {
		err = NoExactFontMatch
	}

	handle := w32.CreateFontIndirect(&logfont)
	if handle == 0 {
		return nil, errors.New("wui.NewFont: unable to create font, please check your description")
	}

	return &Font{Desc: desc, handle: handle}, err
}

type FontDesc struct {
	Name       string
	Height     int
	Bold       bool
	Italic     bool
	Underlined bool
	StrikedOut bool
}

type Font struct {
	Desc   FontDesc
	handle w32.HFONT
}
