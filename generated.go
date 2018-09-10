//+build windows

package wui

import "github.com/gonutz/w32"

func (*Button) isControl() {}

func (control *Button) X() int {
	return control.x
}

func (control *Button) Y() int {
	return control.y
}

func (control *Button) Pos() (x, y int) {
	return control.x, control.y
}

func (control *Button) Width() int {
	return control.width
}

func (control *Button) Height() int {
	return control.height
}

func (control *Button) Size() (width, height int) {
	return control.width, control.height
}

func (control *Button) Bounds() (x, y, width, height int) {
	return control.x, control.y, control.width, control.height
}

func (control *Button) SetX(x int) *Button {
	return control.SetBounds(x, control.y, control.width, control.height)
}

func (control *Button) SetY(y int) *Button {
	return control.SetBounds(control.x, y, control.width, control.height)
}

func (control *Button) SetPos(x, y int) *Button {
	return control.SetBounds(x, y, control.width, control.height)
}

func (control *Button) SetWidth(width int) *Button {
	return control.SetBounds(control.x, control.y, width, control.height)
}

func (control *Button) SetHeight(height int) *Button {
	return control.SetBounds(control.x, control.y, control.width, height)
}

func (control *Button) SetSize(width, height int) *Button {
	return control.SetBounds(control.x, control.y, width, height)
}

func (control *Button) SetBounds(x, y, width, height int) *Button {
	control.x = x
	control.y = y
	control.width = width
	control.height = height
	if control.handle != 0 {
		w32.SetWindowPos(
			control.handle, 0,
			control.x, control.y, control.width, control.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return control
}

func (control *Button) Visible() bool {
	return !control.hidden
}

func (control *Button) SetVisible(v bool) *Button {
	control.hidden = !v
	if control.handle != 0 {
		cmd := w32.SW_SHOW
		if control.hidden {
			cmd = w32.SW_HIDE
		}
		w32.ShowWindow(control.handle, cmd)
	}
	return control
}

func (control *Button) Enabled() bool {
	return !control.disabled
}

func (control *Button) SetEnabled(e bool) *Button {
	control.disabled = !e
	if control.handle != 0 {
		w32.EnableWindow(control.handle, e)
	}
	return control
}

func (control *Button) Text() string {
	return control.text
}

func (control *Button) SetText(text string) *Button {
	control.text = text
	if control.handle != 0 {
		w32.SetWindowText(control.handle, control.text)
	}
	return control
}

func (control *Button) Font() *Font {
	return control.font
}

func (control *Button) SetFont(f *Font) *Button {
	control.font = f
	if control.handle != 0 {
		if control.font != nil {
			control.font.create()
			w32.SendMessage(control.handle, w32.WM_SETFONT, uintptr(control.font.handle), 1)
		}
		if control.font == nil && control.parent != nil && control.parent.font != nil {
			w32.SendMessage(
				control.handle,
				w32.WM_SETFONT,
				uintptr(control.parent.font.handle),
				1,
			)
		}
	}
	return control
}

func (control *Button) afterCreate(parent *Window) {
	control.parent = parent
	if control.hidden {
		w32.ShowWindow(control.handle, w32.SW_HIDE)
	}
	if control.disabled {
		w32.EnableWindow(control.handle, false)
	}
	if control.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.font.handle),
			1,
		)
	} else if control.font == nil && parent != nil && parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(parent.font.handle),
			1,
		)
	}
}

func (control *Button) parentFontChanged() {
	if control.font == nil && control.parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.parent.font.handle),
			1,
		)
	}
}

func (*Checkbox) isControl() {}

func (control *Checkbox) X() int {
	return control.x
}

func (control *Checkbox) Y() int {
	return control.y
}

func (control *Checkbox) Pos() (x, y int) {
	return control.x, control.y
}

func (control *Checkbox) Width() int {
	return control.width
}

func (control *Checkbox) Height() int {
	return control.height
}

func (control *Checkbox) Size() (width, height int) {
	return control.width, control.height
}

func (control *Checkbox) Bounds() (x, y, width, height int) {
	return control.x, control.y, control.width, control.height
}

func (control *Checkbox) SetX(x int) *Checkbox {
	return control.SetBounds(x, control.y, control.width, control.height)
}

func (control *Checkbox) SetY(y int) *Checkbox {
	return control.SetBounds(control.x, y, control.width, control.height)
}

func (control *Checkbox) SetPos(x, y int) *Checkbox {
	return control.SetBounds(x, y, control.width, control.height)
}

func (control *Checkbox) SetWidth(width int) *Checkbox {
	return control.SetBounds(control.x, control.y, width, control.height)
}

func (control *Checkbox) SetHeight(height int) *Checkbox {
	return control.SetBounds(control.x, control.y, control.width, height)
}

func (control *Checkbox) SetSize(width, height int) *Checkbox {
	return control.SetBounds(control.x, control.y, width, height)
}

func (control *Checkbox) SetBounds(x, y, width, height int) *Checkbox {
	control.x = x
	control.y = y
	control.width = width
	control.height = height
	if control.handle != 0 {
		w32.SetWindowPos(
			control.handle, 0,
			control.x, control.y, control.width, control.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return control
}

func (control *Checkbox) Visible() bool {
	return !control.hidden
}

func (control *Checkbox) SetVisible(v bool) *Checkbox {
	control.hidden = !v
	if control.handle != 0 {
		cmd := w32.SW_SHOW
		if control.hidden {
			cmd = w32.SW_HIDE
		}
		w32.ShowWindow(control.handle, cmd)
	}
	return control
}

func (control *Checkbox) Enabled() bool {
	return !control.disabled
}

func (control *Checkbox) SetEnabled(e bool) *Checkbox {
	control.disabled = !e
	if control.handle != 0 {
		w32.EnableWindow(control.handle, e)
	}
	return control
}

func (control *Checkbox) Text() string {
	return control.text
}

func (control *Checkbox) SetText(text string) *Checkbox {
	control.text = text
	if control.handle != 0 {
		w32.SetWindowText(control.handle, control.text)
	}
	return control
}

func (control *Checkbox) Font() *Font {
	return control.font
}

func (control *Checkbox) SetFont(f *Font) *Checkbox {
	control.font = f
	if control.handle != 0 {
		if control.font != nil {
			control.font.create()
			w32.SendMessage(control.handle, w32.WM_SETFONT, uintptr(control.font.handle), 1)
		}
		if control.font == nil && control.parent != nil && control.parent.font != nil {
			w32.SendMessage(
				control.handle,
				w32.WM_SETFONT,
				uintptr(control.parent.font.handle),
				1,
			)
		}
	}
	return control
}

func (control *Checkbox) afterCreate(parent *Window) {
	control.parent = parent
	if control.hidden {
		w32.ShowWindow(control.handle, w32.SW_HIDE)
	}
	if control.disabled {
		w32.EnableWindow(control.handle, false)
	}
	if control.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.font.handle),
			1,
		)
	} else if control.font == nil && parent != nil && parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(parent.font.handle),
			1,
		)
	}
}

func (control *Checkbox) parentFontChanged() {
	if control.font == nil && control.parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.parent.font.handle),
			1,
		)
	}
}

func (*EditLine) isControl() {}

func (control *EditLine) X() int {
	return control.x
}

func (control *EditLine) Y() int {
	return control.y
}

func (control *EditLine) Pos() (x, y int) {
	return control.x, control.y
}

func (control *EditLine) Width() int {
	return control.width
}

func (control *EditLine) Height() int {
	return control.height
}

func (control *EditLine) Size() (width, height int) {
	return control.width, control.height
}

func (control *EditLine) Bounds() (x, y, width, height int) {
	return control.x, control.y, control.width, control.height
}

func (control *EditLine) SetX(x int) *EditLine {
	return control.SetBounds(x, control.y, control.width, control.height)
}

func (control *EditLine) SetY(y int) *EditLine {
	return control.SetBounds(control.x, y, control.width, control.height)
}

func (control *EditLine) SetPos(x, y int) *EditLine {
	return control.SetBounds(x, y, control.width, control.height)
}

func (control *EditLine) SetWidth(width int) *EditLine {
	return control.SetBounds(control.x, control.y, width, control.height)
}

func (control *EditLine) SetHeight(height int) *EditLine {
	return control.SetBounds(control.x, control.y, control.width, height)
}

func (control *EditLine) SetSize(width, height int) *EditLine {
	return control.SetBounds(control.x, control.y, width, height)
}

func (control *EditLine) SetBounds(x, y, width, height int) *EditLine {
	control.x = x
	control.y = y
	control.width = width
	control.height = height
	if control.handle != 0 {
		w32.SetWindowPos(
			control.handle, 0,
			control.x, control.y, control.width, control.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return control
}

func (control *EditLine) Visible() bool {
	return !control.hidden
}

func (control *EditLine) SetVisible(v bool) *EditLine {
	control.hidden = !v
	if control.handle != 0 {
		cmd := w32.SW_SHOW
		if control.hidden {
			cmd = w32.SW_HIDE
		}
		w32.ShowWindow(control.handle, cmd)
	}
	return control
}

func (control *EditLine) Enabled() bool {
	return !control.disabled
}

func (control *EditLine) SetEnabled(e bool) *EditLine {
	control.disabled = !e
	if control.handle != 0 {
		w32.EnableWindow(control.handle, e)
	}
	return control
}

func (control *EditLine) Text() string {
	return control.text
}

func (control *EditLine) SetText(text string) *EditLine {
	control.text = text
	if control.handle != 0 {
		w32.SetWindowText(control.handle, control.text)
	}
	return control
}

func (control *EditLine) Font() *Font {
	return control.font
}

func (control *EditLine) SetFont(f *Font) *EditLine {
	control.font = f
	if control.handle != 0 {
		if control.font != nil {
			control.font.create()
			w32.SendMessage(control.handle, w32.WM_SETFONT, uintptr(control.font.handle), 1)
		}
		if control.font == nil && control.parent != nil && control.parent.font != nil {
			w32.SendMessage(
				control.handle,
				w32.WM_SETFONT,
				uintptr(control.parent.font.handle),
				1,
			)
		}
	}
	return control
}

func (control *EditLine) afterCreate(parent *Window) {
	control.parent = parent
	if control.hidden {
		w32.ShowWindow(control.handle, w32.SW_HIDE)
	}
	if control.disabled {
		w32.EnableWindow(control.handle, false)
	}
	if control.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.font.handle),
			1,
		)
	} else if control.font == nil && parent != nil && parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(parent.font.handle),
			1,
		)
	}
}

func (control *EditLine) parentFontChanged() {
	if control.font == nil && control.parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.parent.font.handle),
			1,
		)
	}
}

func (*Label) isControl() {}

func (control *Label) X() int {
	return control.x
}

func (control *Label) Y() int {
	return control.y
}

func (control *Label) Pos() (x, y int) {
	return control.x, control.y
}

func (control *Label) Width() int {
	return control.width
}

func (control *Label) Height() int {
	return control.height
}

func (control *Label) Size() (width, height int) {
	return control.width, control.height
}

func (control *Label) Bounds() (x, y, width, height int) {
	return control.x, control.y, control.width, control.height
}

func (control *Label) SetX(x int) *Label {
	return control.SetBounds(x, control.y, control.width, control.height)
}

func (control *Label) SetY(y int) *Label {
	return control.SetBounds(control.x, y, control.width, control.height)
}

func (control *Label) SetPos(x, y int) *Label {
	return control.SetBounds(x, y, control.width, control.height)
}

func (control *Label) SetWidth(width int) *Label {
	return control.SetBounds(control.x, control.y, width, control.height)
}

func (control *Label) SetHeight(height int) *Label {
	return control.SetBounds(control.x, control.y, control.width, height)
}

func (control *Label) SetSize(width, height int) *Label {
	return control.SetBounds(control.x, control.y, width, height)
}

func (control *Label) SetBounds(x, y, width, height int) *Label {
	control.x = x
	control.y = y
	control.width = width
	control.height = height
	if control.handle != 0 {
		w32.SetWindowPos(
			control.handle, 0,
			control.x, control.y, control.width, control.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return control
}

func (control *Label) Visible() bool {
	return !control.hidden
}

func (control *Label) SetVisible(v bool) *Label {
	control.hidden = !v
	if control.handle != 0 {
		cmd := w32.SW_SHOW
		if control.hidden {
			cmd = w32.SW_HIDE
		}
		w32.ShowWindow(control.handle, cmd)
	}
	return control
}

func (control *Label) Enabled() bool {
	return !control.disabled
}

func (control *Label) SetEnabled(e bool) *Label {
	control.disabled = !e
	if control.handle != 0 {
		w32.EnableWindow(control.handle, e)
	}
	return control
}

func (control *Label) Text() string {
	return control.text
}

func (control *Label) SetText(text string) *Label {
	control.text = text
	if control.handle != 0 {
		w32.SetWindowText(control.handle, control.text)
	}
	return control
}

func (control *Label) Font() *Font {
	return control.font
}

func (control *Label) SetFont(f *Font) *Label {
	control.font = f
	if control.handle != 0 {
		if control.font != nil {
			control.font.create()
			w32.SendMessage(control.handle, w32.WM_SETFONT, uintptr(control.font.handle), 1)
		}
		if control.font == nil && control.parent != nil && control.parent.font != nil {
			w32.SendMessage(
				control.handle,
				w32.WM_SETFONT,
				uintptr(control.parent.font.handle),
				1,
			)
		}
	}
	return control
}

func (control *Label) afterCreate(parent *Window) {
	control.parent = parent
	if control.hidden {
		w32.ShowWindow(control.handle, w32.SW_HIDE)
	}
	if control.disabled {
		w32.EnableWindow(control.handle, false)
	}
	if control.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.font.handle),
			1,
		)
	} else if control.font == nil && parent != nil && parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(parent.font.handle),
			1,
		)
	}
}

func (control *Label) parentFontChanged() {
	if control.font == nil && control.parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.parent.font.handle),
			1,
		)
	}
}

func (*RadioButton) isControl() {}

func (control *RadioButton) X() int {
	return control.x
}

func (control *RadioButton) Y() int {
	return control.y
}

func (control *RadioButton) Pos() (x, y int) {
	return control.x, control.y
}

func (control *RadioButton) Width() int {
	return control.width
}

func (control *RadioButton) Height() int {
	return control.height
}

func (control *RadioButton) Size() (width, height int) {
	return control.width, control.height
}

func (control *RadioButton) Bounds() (x, y, width, height int) {
	return control.x, control.y, control.width, control.height
}

func (control *RadioButton) SetX(x int) *RadioButton {
	return control.SetBounds(x, control.y, control.width, control.height)
}

func (control *RadioButton) SetY(y int) *RadioButton {
	return control.SetBounds(control.x, y, control.width, control.height)
}

func (control *RadioButton) SetPos(x, y int) *RadioButton {
	return control.SetBounds(x, y, control.width, control.height)
}

func (control *RadioButton) SetWidth(width int) *RadioButton {
	return control.SetBounds(control.x, control.y, width, control.height)
}

func (control *RadioButton) SetHeight(height int) *RadioButton {
	return control.SetBounds(control.x, control.y, control.width, height)
}

func (control *RadioButton) SetSize(width, height int) *RadioButton {
	return control.SetBounds(control.x, control.y, width, height)
}

func (control *RadioButton) SetBounds(x, y, width, height int) *RadioButton {
	control.x = x
	control.y = y
	control.width = width
	control.height = height
	if control.handle != 0 {
		w32.SetWindowPos(
			control.handle, 0,
			control.x, control.y, control.width, control.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return control
}

func (control *RadioButton) Visible() bool {
	return !control.hidden
}

func (control *RadioButton) SetVisible(v bool) *RadioButton {
	control.hidden = !v
	if control.handle != 0 {
		cmd := w32.SW_SHOW
		if control.hidden {
			cmd = w32.SW_HIDE
		}
		w32.ShowWindow(control.handle, cmd)
	}
	return control
}

func (control *RadioButton) Enabled() bool {
	return !control.disabled
}

func (control *RadioButton) SetEnabled(e bool) *RadioButton {
	control.disabled = !e
	if control.handle != 0 {
		w32.EnableWindow(control.handle, e)
	}
	return control
}

func (control *RadioButton) Text() string {
	return control.text
}

func (control *RadioButton) SetText(text string) *RadioButton {
	control.text = text
	if control.handle != 0 {
		w32.SetWindowText(control.handle, control.text)
	}
	return control
}

func (control *RadioButton) Font() *Font {
	return control.font
}

func (control *RadioButton) SetFont(f *Font) *RadioButton {
	control.font = f
	if control.handle != 0 {
		if control.font != nil {
			control.font.create()
			w32.SendMessage(control.handle, w32.WM_SETFONT, uintptr(control.font.handle), 1)
		}
		if control.font == nil && control.parent != nil && control.parent.font != nil {
			w32.SendMessage(
				control.handle,
				w32.WM_SETFONT,
				uintptr(control.parent.font.handle),
				1,
			)
		}
	}
	return control
}

func (control *RadioButton) afterCreate(parent *Window) {
	control.parent = parent
	if control.hidden {
		w32.ShowWindow(control.handle, w32.SW_HIDE)
	}
	if control.disabled {
		w32.EnableWindow(control.handle, false)
	}
	if control.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.font.handle),
			1,
		)
	} else if control.font == nil && parent != nil && parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(parent.font.handle),
			1,
		)
	}
}

func (control *RadioButton) parentFontChanged() {
	if control.font == nil && control.parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.parent.font.handle),
			1,
		)
	}
}

func (*Panel) isControl() {}

func (control *Panel) X() int {
	return control.x
}

func (control *Panel) Y() int {
	return control.y
}

func (control *Panel) Pos() (x, y int) {
	return control.x, control.y
}

func (control *Panel) Width() int {
	return control.width
}

func (control *Panel) Height() int {
	return control.height
}

func (control *Panel) Size() (width, height int) {
	return control.width, control.height
}

func (control *Panel) Bounds() (x, y, width, height int) {
	return control.x, control.y, control.width, control.height
}

func (control *Panel) SetX(x int) *Panel {
	return control.SetBounds(x, control.y, control.width, control.height)
}

func (control *Panel) SetY(y int) *Panel {
	return control.SetBounds(control.x, y, control.width, control.height)
}

func (control *Panel) SetPos(x, y int) *Panel {
	return control.SetBounds(x, y, control.width, control.height)
}

func (control *Panel) SetWidth(width int) *Panel {
	return control.SetBounds(control.x, control.y, width, control.height)
}

func (control *Panel) SetHeight(height int) *Panel {
	return control.SetBounds(control.x, control.y, control.width, height)
}

func (control *Panel) SetSize(width, height int) *Panel {
	return control.SetBounds(control.x, control.y, width, height)
}

func (control *Panel) SetBounds(x, y, width, height int) *Panel {
	control.x = x
	control.y = y
	control.width = width
	control.height = height
	if control.handle != 0 {
		w32.SetWindowPos(
			control.handle, 0,
			control.x, control.y, control.width, control.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return control
}

func (control *Panel) Visible() bool {
	return !control.hidden
}

func (control *Panel) SetVisible(v bool) *Panel {
	control.hidden = !v
	if control.handle != 0 {
		cmd := w32.SW_SHOW
		if control.hidden {
			cmd = w32.SW_HIDE
		}
		w32.ShowWindow(control.handle, cmd)
	}
	return control
}

func (control *Panel) Enabled() bool {
	return !control.disabled
}

func (control *Panel) SetEnabled(e bool) *Panel {
	control.disabled = !e
	if control.handle != 0 {
		w32.EnableWindow(control.handle, e)
	}
	return control
}

func (control *Panel) afterCreate(parent *Window) {
	control.parent = parent
	if control.hidden {
		w32.ShowWindow(control.handle, w32.SW_HIDE)
	}
	if control.disabled {
		w32.EnableWindow(control.handle, false)
	}
}

func (control *Panel) parentFontChanged() {}

func (*Paintbox) isControl() {}

func (control *Paintbox) X() int {
	return control.x
}

func (control *Paintbox) Y() int {
	return control.y
}

func (control *Paintbox) Pos() (x, y int) {
	return control.x, control.y
}

func (control *Paintbox) Width() int {
	return control.width
}

func (control *Paintbox) Height() int {
	return control.height
}

func (control *Paintbox) Size() (width, height int) {
	return control.width, control.height
}

func (control *Paintbox) Bounds() (x, y, width, height int) {
	return control.x, control.y, control.width, control.height
}

func (control *Paintbox) SetX(x int) *Paintbox {
	return control.SetBounds(x, control.y, control.width, control.height)
}

func (control *Paintbox) SetY(y int) *Paintbox {
	return control.SetBounds(control.x, y, control.width, control.height)
}

func (control *Paintbox) SetPos(x, y int) *Paintbox {
	return control.SetBounds(x, y, control.width, control.height)
}

func (control *Paintbox) SetWidth(width int) *Paintbox {
	return control.SetBounds(control.x, control.y, width, control.height)
}

func (control *Paintbox) SetHeight(height int) *Paintbox {
	return control.SetBounds(control.x, control.y, control.width, height)
}

func (control *Paintbox) SetSize(width, height int) *Paintbox {
	return control.SetBounds(control.x, control.y, width, height)
}

func (control *Paintbox) SetBounds(x, y, width, height int) *Paintbox {
	control.x = x
	control.y = y
	control.width = width
	control.height = height
	if control.handle != 0 {
		w32.SetWindowPos(
			control.handle, 0,
			control.x, control.y, control.width, control.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return control
}

func (control *Paintbox) Visible() bool {
	return !control.hidden
}

func (control *Paintbox) SetVisible(v bool) *Paintbox {
	control.hidden = !v
	if control.handle != 0 {
		cmd := w32.SW_SHOW
		if control.hidden {
			cmd = w32.SW_HIDE
		}
		w32.ShowWindow(control.handle, cmd)
	}
	return control
}

func (control *Paintbox) Enabled() bool {
	return !control.disabled
}

func (control *Paintbox) SetEnabled(e bool) *Paintbox {
	control.disabled = !e
	if control.handle != 0 {
		w32.EnableWindow(control.handle, e)
	}
	return control
}

func (control *Paintbox) afterCreate(parent *Window) {
	control.parent = parent
	if control.hidden {
		w32.ShowWindow(control.handle, w32.SW_HIDE)
	}
	if control.disabled {
		w32.EnableWindow(control.handle, false)
	}
}

func (control *Paintbox) parentFontChanged() {}

func (*ProgressBar) isControl() {}

func (control *ProgressBar) X() int {
	return control.x
}

func (control *ProgressBar) Y() int {
	return control.y
}

func (control *ProgressBar) Pos() (x, y int) {
	return control.x, control.y
}

func (control *ProgressBar) Width() int {
	return control.width
}

func (control *ProgressBar) Height() int {
	return control.height
}

func (control *ProgressBar) Size() (width, height int) {
	return control.width, control.height
}

func (control *ProgressBar) Bounds() (x, y, width, height int) {
	return control.x, control.y, control.width, control.height
}

func (control *ProgressBar) SetX(x int) *ProgressBar {
	return control.SetBounds(x, control.y, control.width, control.height)
}

func (control *ProgressBar) SetY(y int) *ProgressBar {
	return control.SetBounds(control.x, y, control.width, control.height)
}

func (control *ProgressBar) SetPos(x, y int) *ProgressBar {
	return control.SetBounds(x, y, control.width, control.height)
}

func (control *ProgressBar) SetWidth(width int) *ProgressBar {
	return control.SetBounds(control.x, control.y, width, control.height)
}

func (control *ProgressBar) SetHeight(height int) *ProgressBar {
	return control.SetBounds(control.x, control.y, control.width, height)
}

func (control *ProgressBar) SetSize(width, height int) *ProgressBar {
	return control.SetBounds(control.x, control.y, width, height)
}

func (control *ProgressBar) SetBounds(x, y, width, height int) *ProgressBar {
	control.x = x
	control.y = y
	control.width = width
	control.height = height
	if control.handle != 0 {
		w32.SetWindowPos(
			control.handle, 0,
			control.x, control.y, control.width, control.height,
			w32.SWP_NOOWNERZORDER|w32.SWP_NOZORDER,
		)
	}
	return control
}

func (control *ProgressBar) Visible() bool {
	return !control.hidden
}

func (control *ProgressBar) SetVisible(v bool) *ProgressBar {
	control.hidden = !v
	if control.handle != 0 {
		cmd := w32.SW_SHOW
		if control.hidden {
			cmd = w32.SW_HIDE
		}
		w32.ShowWindow(control.handle, cmd)
	}
	return control
}

func (control *ProgressBar) Enabled() bool {
	return !control.disabled
}

func (control *ProgressBar) SetEnabled(e bool) *ProgressBar {
	control.disabled = !e
	if control.handle != 0 {
		w32.EnableWindow(control.handle, e)
	}
	return control
}

func (control *ProgressBar) afterCreate(parent *Window) {
	control.parent = parent
	if control.hidden {
		w32.ShowWindow(control.handle, w32.SW_HIDE)
	}
	if control.disabled {
		w32.EnableWindow(control.handle, false)
	}
}

func (control *ProgressBar) parentFontChanged() {}
