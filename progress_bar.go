//+build windows

package wui

import "github.com/gonutz/w32"

func NewProgressBar() *ProgressBar {
	return &ProgressBar{}
}

type ProgressBar struct {
	control
	id           int
	vertical     bool
	movesForever bool
	value        float64
}

var _ Control = (*ProgressBar)(nil)

func (*ProgressBar) canFocus() bool {
	return false
}

func (*ProgressBar) eatsTabs() bool {
	return false
}

const maxProgressBarValue = 10000

func (p *ProgressBar) create(id int) {
	p.id = id
	p.recreate()
}

func (p *ProgressBar) recreate() {
	// Updating the window style for a progress bar does not work, instead we
	// have to destroy the window and create it anew. This changes the window
	// handle but is the only option we have.
	if p.handle != 0 {
		w32.DestroyWindow(p.handle)
	}
	var style uint = w32.PBS_SMOOTH
	if p.vertical {
		style |= w32.PBS_VERTICAL
	}
	if p.movesForever {
		style |= w32.PBS_MARQUEE
	}
	p.control.create(p.id, w32.WS_EX_CLIENTEDGE, w32.PROGRESS_CLASS, style)
	if p.movesForever {
		w32.SendMessage(p.handle, w32.PBM_SETMARQUEE, 1, 0)
	} else {
		w32.SendMessage(p.handle, w32.PBM_SETRANGE32, 0, maxProgressBarValue)
		p.SetValue(p.value)
	}
}

func (p *ProgressBar) Vertical() bool {
	return p.vertical
}

func (p *ProgressBar) SetVertical(v bool) {
	if v != p.vertical {
		p.vertical = v
		if p.handle != 0 {
			p.recreate()
		}
	}
}

func (p *ProgressBar) MovesForever() bool {
	return p.movesForever
}

func (p *ProgressBar) SetMovesForever(move bool) {
	if p.movesForever != move {
		p.movesForever = move
		if p.handle != 0 {
			p.recreate()
		}
	}
}

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

	if !p.movesForever && p.handle != 0 {
		pos := int(v*maxProgressBarValue + 0.5)
		w32.SendMessage(p.handle, w32.PBM_SETPOS, uintptr(pos), 0)
	}
}
