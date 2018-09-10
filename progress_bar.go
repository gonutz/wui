//+build windows

package wui

import "github.com/gonutz/w32"

const maxProgressBarValue = 10000

type ProgressBar struct {
	handle   w32.HWND
	parent   *Window
	x        int
	y        int
	width    int
	height   int
	hidden   bool
	disabled bool
	value    float64
}

func NewProgressBar() *ProgressBar {
	return &ProgressBar{}
}

func (p *ProgressBar) Value() float64 {
	return p.value
}

func (p *ProgressBar) SetValue(v float64) *ProgressBar {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	p.value = v
	if p.handle != 0 {
		pos := int(v*maxProgressBarValue + 0.5)
		w32.SendMessage(p.handle, w32.PBM_SETPOS, uintptr(pos), 0)
	}
	return p
}
