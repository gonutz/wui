//+build windows

package wui

import "github.com/gonutz/w32/v2"

// TODO: We can create two radio buttons both set to checked and only later add
// them both to their parents. This will keep them both checked, SetChecked only
// sets siblings to false when the parent is assigned. parent.Add could make
// sure that only one radio button is checked. Maybe do this in RadioButton.create
// since it is guaranteed to have its parents set.

// NewRadioButton returns a new, unchecked RadioButton.
func NewRadioButton() *RadioButton {
	return &RadioButton{}
}

// RadioButton is a control with a circle that is either filled to indicate
// checked state or unfilled to indicate unchecked state. It has a label next to
// the circle. Of multiple radio buttons in a container, only one can ever be
// checked at a time. Also all might be unchecked. Checking one radio button
// will automatically uncheck all others in the container.
type RadioButton struct {
	textControl
	checked bool
	onCheck func(bool)
}

var _ Control = (*RadioButton)(nil)

func (*RadioButton) canFocus() bool {
	return true
}

func (r *RadioButton) OnTabFocus() func() {
	return r.onTabFocus
}

func (r *RadioButton) SetOnTabFocus(f func()) {
	r.onTabFocus = f
}

func (*RadioButton) eatsTabs() bool {
	return false
}

func (r *RadioButton) create(id int) {
	var style uint = w32.WS_TABSTOP | w32.BS_AUTORADIOBUTTON | w32.BS_NOTIFY
	r.textControl.create(id, 0, "BUTTON", style)
	if r.checked {
		w32.SendMessage(r.handle, w32.BM_SETCHECK, toCheckState(r.checked), 0)
	}
}

// Checked returns true if the radio button is checked and false if not.
func (r *RadioButton) Checked() bool {
	r.updateCachedCheckState()
	return r.checked
}

func (r *RadioButton) updateCachedCheckState() {
	if r.handle != 0 {
		r.checked = w32.BST_CHECKED == w32.SendMessage(r.handle, w32.BM_GETCHECK, 0, 0)
	}
}

// SetChecked updates the state of the RadioButton. If checked is true and the
// button was not checked before, the OnCheck notification is called. Checking a
// radio button will uncheck all its siblings, i.e. all radio buttons that are
// in the same parent container.
func (r *RadioButton) SetChecked(checked bool) {
	r.updateCachedCheckState()
	if checked == r.checked {
		return
	}
	r.checked = checked
	if r.handle != 0 {
		// Windows will uncheck all siblings for us.
		w32.SendMessage(r.handle, w32.BM_SETCHECK, toCheckState(r.checked), 0)
	} else if checked {
		// If a radio button gets checked before we have a window handle,
		// Windows will not uncheck its siblings for us, we have to do it
		// ourselves.
		if r.parent != nil {
			for _, sibling := range r.parent.Children() {
				if radio, ok := sibling.(*RadioButton); ok {
					if radio != r {
						radio.SetChecked(false)
					}
				}
			}
		}
		if r.onCheck != nil {
			r.onCheck(r.checked)
		}
	}
}

// OnCheck returns the callback set in SetOnCheck.
func (r *RadioButton) OnCheck() func(checked bool) {
	return r.onCheck
}

// SetOnCheck sets a callback for when the RadioButton is checked. It is not
// called when the RadioButton is being unchecked.
func (r *RadioButton) SetOnCheck(f func(checked bool)) {
	r.onCheck = f
}

func (r *RadioButton) handleNotification(cmd uintptr) {
	if cmd == w32.BN_CLICKED {
		r.updateCachedCheckState()
		if r.checked && r.onCheck != nil {
			r.onCheck(r.checked)
		}
	}
}
