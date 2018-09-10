//+build windows

package wui

import "github.com/gonutz/w32"

func NewRadioButton() *RadioButton {
	return &RadioButton{}
}

type RadioButton struct {
	textControl
	checked  bool
	onChange func(bool)
}

func (r *RadioButton) create(id int) {
	r.textControl.create(
		id,
		0,
		"BUTTON",
		w32.WS_TABSTOP|w32.BS_AUTORADIOBUTTON|w32.BS_NOTIFY,
	)
	w32.SendMessage(r.handle, w32.BM_SETCHECK, toCheckState(r.checked), 0)
}

func (r *RadioButton) Checked() bool {
	return r.checked
}

func (r *RadioButton) SetChecked(checked bool) {
	if checked == r.checked {
		return
	}
	r.checked = checked
	if r.handle != 0 {
		w32.SendMessage(r.handle, w32.BM_SETCHECK, toCheckState(r.checked), 0)
	}
	if r.onChange != nil {
		r.onChange(r.checked)
	}
	return
}

func (r *RadioButton) SetOnChange(f func(checked bool)) {
	r.onChange = f
}
