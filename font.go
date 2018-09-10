//+build windows

package wui

import (
	"unicode/utf16"

	"github.com/gonutz/w32"
)

type Font struct {
	handle     w32.HFONT
	name       string
	height     int
	bold       bool
	italic     bool
	underlined bool
	strikedOut bool
}

func NewFont() *Font {
	return &Font{}
}

func (f *Font) Name() string     { return f.name }
func (f *Font) Height() int      { return f.height }
func (f *Font) Bold() bool       { return f.bold }
func (f *Font) Italic() bool     { return f.italic }
func (f *Font) Underlined() bool { return f.underlined }
func (f *Font) StrikedOut() bool { return f.strikedOut }

func (f *Font) create() {
	if f.handle != 0 {
		w32.DeleteObject(w32.HGDIOBJ(f.handle))
	}
	weight := int32(w32.FW_NORMAL)
	if f.bold {
		weight = w32.FW_BOLD
	}
	byteBool := func(b bool) byte {
		if b {
			return 1
		}
		return 0
	}
	desc := w32.LOGFONT{
		Height:         int32(f.height),
		Width:          0,
		Escapement:     0,
		Orientation:    0,
		Weight:         weight,
		Italic:         byteBool(f.italic),
		Underline:      byteBool(f.underlined),
		StrikeOut:      byteBool(f.strikedOut),
		CharSet:        w32.DEFAULT_CHARSET,
		OutPrecision:   w32.OUT_CHARACTER_PRECIS,
		ClipPrecision:  w32.CLIP_CHARACTER_PRECIS,
		Quality:        w32.DEFAULT_QUALITY,
		PitchAndFamily: w32.DEFAULT_PITCH | w32.FF_DONTCARE,
	}
	copy(desc.FaceName[:], utf16.Encode([]rune(f.name)))
	f.handle = w32.CreateFontIndirect(&desc)
}

func (f *Font) SetName(name string) *Font {
	f.name = name
	return f
}

func (f *Font) SetHeight(height int) *Font {
	f.height = height
	return f
}

func (f *Font) SetBold(bold bool) *Font {
	f.bold = bold
	return f
}

func (f *Font) SetItalic(italic bool) *Font {
	f.italic = italic
	return f
}

func (f *Font) SetUnderlined(underlined bool) *Font {
	f.underlined = underlined
	return f
}

func (f *Font) SetStrikedOut(strikedOut bool) *Font {
	f.strikedOut = strikedOut
	return f
}
