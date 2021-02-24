//+build windows

package wui

import "github.com/gonutz/w32/v2"

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

var _ Control = (*Slider)(nil)

func (s *Slider) closing() {
	s.CursorPosition()
}

func (*Slider) canFocus() bool {
	return true
}

func (s *Slider) OnTabFocus() func() {
	return s.onTabFocus
}

func (s *Slider) SetOnTabFocus(f func()) {
	s.onTabFocus = f
}

func (*Slider) eatsTabs() bool {
	return false
}

type SliderOrientation int

const (
	HorizontalSlider SliderOrientation = iota
	VerticalSlider
)

func (o SliderOrientation) String() string {
	// NOTE that these strings are used in the designer to get their
	// representations as Go code so they must always correspond to their
	// constant names and be prefixed with the package name.
	switch o {
	case HorizontalSlider:
		return "wui.HorizontalSlider"
	case VerticalSlider:
		return "wui.VerticalSlider"
	default:
		return "unknown SliderOrientation"
	}
}

type TickPosition int

const (
	TicksBottomOrRight TickPosition = iota
	TicksTopOrLeft
	TicksOnBothSides
)

func (p TickPosition) String() string {
	// NOTE that these strings are used in the designer to get their
	// representations as Go code so they must always correspond to their
	// constant names and be prefixed with the package name.
	switch p {
	case TicksBottomOrRight:
		return "wui.TicksBottomOrRight"
	case TicksTopOrLeft:
		return "wui.TicksTopOrLeft"
	case TicksOnBothSides:
		return "wui.TicksOnBothSides"
	default:
		return "unknown TickPosition"
	}
}

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
		s.SetCursorPosition(s.cursor)
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

// SetMinMax sets the minimum and maximum values of the Slider. min must be
// smaller or equal to max. If the current cursor is outside these ranges, it is
// clamped.
func (s *Slider) SetMinMax(min, max int) {
	s.min, s.max = min, max
	if s.handle != 0 {
		const redraw = 1
		w32.SendMessage(s.handle, w32.TBM_SETRANGEMIN, 0, uintptr(min))
		w32.SendMessage(s.handle, w32.TBM_SETRANGEMAX, redraw, uintptr(max))
	} else {
		if s.cursor < min {
			s.cursor = min
		}
		if s.cursor > max {
			s.cursor = max
		}
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

// SetCursorPosition sets the position of the cursor. It clamps it to the
// Slider's min/max values.
func (s *Slider) SetCursorPosition(cursor int) {
	if cursor < s.min {
		cursor = s.min
	}
	if cursor > s.max {
		cursor = s.max
	}
	s.cursor = cursor
	if s.handle != 0 {
		const redraw = 1
		w32.SendMessage(s.handle, w32.TBM_SETPOS, redraw, uintptr(s.cursor))
	}
}

// CursorPosition returns the current position of the Slider.
func (s *Slider) CursorPosition() int {
	if s.handle != 0 {
		s.cursor = int(w32.SendMessage(s.handle, w32.TBM_GETPOS, 0, 0))
	}
	return s.cursor
}

func (s *Slider) TickFrequency() int {
	return s.tickFrequency
}

// SetTickFrequency sets the distance between two ticks to n (only every n'th
// tick is drawn). The first and last ticks are always drawn (except if you hide
// ticks altogether).
func (s *Slider) SetTickFrequency(n int) {
	if n <= 0 {
		return
	}
	s.tickFrequency = n
	if s.handle != 0 {
		w32.SendMessage(s.handle, w32.TBM_SETTICFREQ, uintptr(n), 0)
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

// SetTicksVisible only takes effect before the window is shown on screen. At
// run-time this cannot be changed.
func (s *Slider) SetTicksVisible(show bool) {
	if s.handle == 0 {
		s.hideTicks = !show
	}
}

func (s *Slider) TicksVisible() bool {
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
		s.onChange(s.CursorPosition())
	}
}
