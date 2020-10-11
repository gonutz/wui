//+build windows

package wui

import "github.com/gonutz/w32"

func NewSlider() *Slider {
	return &Slider{
		max:           10,
		tickFrequency: 1,
		arrowInc:      1,
		mouseInc:      2,
	}
}

type Slider struct {
	control
	min           int
	max           int
	cursor        int
	tickFrequency int
	arrowInc      int
	mouseInc      int
	hideTicks     bool
	vertical      bool
	tickPosition  TickPosition
	onChange      func(cursor int)
}

type SliderOrientation int

const (
	HorizontalSlider SliderOrientation = iota
	VerticalSlider
)

type TickPosition int

const (
	TicksBottomOrRight TickPosition = iota
	TicksTopOrLeft
	TicksOnBothSides
)

func (s *Slider) create(id int) {
	var style uint = w32.WS_TABSTOP

	if s.hideTicks {
		style |= w32.TBS_NOTICKS
	} else {
		style |= w32.TBS_AUTOTICKS
	}

	if s.vertical {
		style |= w32.TBS_VERT
	} else {
		style |= w32.TBS_HORZ
	}

	switch s.tickPosition {
	case TicksBottomOrRight:
		style |= w32.TBS_BOTTOM
	case TicksTopOrLeft:
		style |= w32.TBS_TOP
	case TicksOnBothSides:
		style |= w32.TBS_BOTH
	}

	s.control.create(id, 0, "msctls_trackbar32", style)

	if s.cursor != 0 {
		s.SetCursor(s.cursor)
	}
	if s.tickFrequency != 1 {
		s.SetTickFrequency(s.tickFrequency)
	}
	s.SetMinMax(s.min, s.max)
	s.SetArrowIncrement(s.arrowInc)
	s.SetMouseIncrement(s.mouseInc)
}

func (s *Slider) SetMin(min int) {
	s.SetMinMax(min, s.max)
}

func (s *Slider) SetMax(max int) {
	s.SetMinMax(s.min, max)
}

func (s *Slider) SetMinMax(min, max int) {
	s.min, s.max = min, max
	if s.handle != 0 {
		const redraw = 1
		w32.SendMessage(s.handle, w32.TBM_SETRANGEMIN, 0, uintptr(min))
		w32.SendMessage(s.handle, w32.TBM_SETRANGEMAX, redraw, uintptr(max))
	}
}

func (s *Slider) MinMax() (min, max int) {
	return s.min, s.max
}

func (s *Slider) Min() int {
	min, _ := s.MinMax()
	return min
}

func (s *Slider) Max() int {
	_, max := s.MinMax()
	return max
}

func (s *Slider) SetCursor(cursor int) {
	s.cursor = cursor
	if s.handle != 0 {
		const redraw = 1
		w32.SendMessage(s.handle, w32.TBM_SETPOS, redraw, uintptr(s.cursor))
	}
}

func (s *Slider) Cursor() int {
	if s.handle != 0 {
		s.cursor = int(w32.SendMessage(s.handle, w32.TBM_GETPOS, 0, 0))
	}
	return s.cursor
}

func (s *Slider) TickFrequency() int {
	return s.tickFrequency
}

func (s *Slider) SetTickFrequency(f int) {
	s.tickFrequency = f
	if s.handle != 0 {
		w32.SendMessage(s.handle, w32.TBM_SETTICFREQ, uintptr(f), 0)
	}
}

// SetArrowIncrement is for the arrow keys, left/down/right/up.
func (s *Slider) SetArrowIncrement(inc int) {
	s.arrowInc = inc
	if s.handle != 0 {
		w32.SendMessage(s.handle, w32.TBM_SETLINESIZE, 0, uintptr(inc))
	}
}

func (s *Slider) ArrowIncrement() int {
	return s.arrowInc
}

// SetMouseIncrement is for mouse clicks and the page up/down keys.
func (s *Slider) SetMouseIncrement(inc int) {
	s.mouseInc = inc
	if s.handle != 0 {
		w32.SendMessage(s.handle, w32.TBM_SETPAGESIZE, 0, uintptr(inc))
	}
}

func (s *Slider) MouseIncrement() int {
	return s.mouseInc
}

// SetHasTicks only takes effect before the window is shown on screen. At
// run-time this cannot be changed.
func (s *Slider) SetHasTicks(show bool) {
	if s.handle == 0 {
		s.hideTicks = !show
	}
}

func (s *Slider) HasTicks() bool {
	return !s.hideTicks
}

// SetOrientation only takes effect before the window is shown on screen. At
// run-time this cannot be changed.
func (s *Slider) SetOrientation(o SliderOrientation) {
	if s.handle == 0 {
		s.vertical = o == VerticalSlider
	}
}

func (s *Slider) Orientation() SliderOrientation {
	if s.vertical {
		return VerticalSlider
	}
	return HorizontalSlider
}

// SetTickPosition only takes effect before the window is shown on screen. At
// run-time this cannot be changed.
func (s *Slider) SetTickPosition(p TickPosition) {
	if s.handle == 0 {
		s.tickPosition = p
	}
}

func (s *Slider) TickPosition() TickPosition {
	return s.tickPosition
}

func (s *Slider) SetOnChange(f func(cursor int)) {
	s.onChange = f
}

func (s *Slider) OnChange() func(cursor int) {
	return s.onChange
}

func (s *Slider) handleChange(reason uintptr) {
	if s.onChange != nil &&
		reason != w32.TB_ENDTRACK && reason != w32.TB_THUMBPOSITION {
		s.onChange(s.Cursor())
	}
}
