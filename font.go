//+build windows

package wui

import (
	"errors"
	"unicode/utf16"

	"github.com/gonutz/w32"
)

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
	copy(logfont.FaceName[:], utf16.Encode([]rune(desc.Name)))
	handle := w32.CreateFontIndirect(&logfont)
	if handle == 0 {
		return nil, errors.New("wui.NewFont: unable to create font, please check your description")
	}
	return &Font{Desc: desc, handle: handle}, nil
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
