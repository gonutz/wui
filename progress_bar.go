//+build windows

package wui

import "github.com/gonutz/w32"

func NewProgressBar() *ProgressBar {
	return &ProgressBar{}
}

type ProgressBar struct {
	control
	x      int
	y      int
	width  int
	height int
	hidden bool
	value  float64
}

const maxProgressBarValue = 10000

func (p *ProgressBar) Value() float64 {
	return p.value
}

func (p *ProgressBar) SetValue(v float64) {
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
}
