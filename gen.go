//+build ignore

package main

import (
	"fmt"
	"os"
	"text/template"
)

type control struct {
	TypeName string
	props    []string
}

func main() {
	// init generation data
	commonProps := []string{
		"control",
		"bounds",
		"visible",
		"enabled",
	}
	controls := []control{
		{"Button",
			[]string{
				"text",
			}},
		{"Checkbox",
			[]string{
				"text",
			}},
		{"EditLine",
			[]string{
				"text",
			}},
		{"Label",
			[]string{
				"text",
			}},
		{"RadioButton",
			[]string{
				"text",
			}},
		{"Panel",
			[]string{}},
		{"Paintbox",
			[]string{}},
		{"ProgressBar",
			[]string{}},
	}
	for i := range controls {
		var props []string
		props = append(props, commonProps...)
		props = append(props, controls[i].props...)
		controls[i].props = props
	}

	templates := make(map[string]*template.Template)
	for name, code := range allTemplates {
		templates[name] = template.Must(template.New(name).Parse(code))
	}

	// generate code
	fmt.Println(`//+build windows`)
	fmt.Println()
	fmt.Println(`package wui`)
	fmt.Println()
	fmt.Println(`import "github.com/gonutz/w32"`)
	for _, c := range controls {
		for _, prop := range c.props {
			if t, ok := templates[prop]; ok {
				t.Execute(os.Stdout, c)
			} else {
				panic("no template for " + prop)
			}
		}
		afterCreate := templates["after_create"]
		if contains(c.props, "text") {
			afterCreate = templates["after_create_with_text"]
		}
		afterCreate.Execute(os.Stdout, c)
	}
}

func contains(list []string, find string) bool {
	for _, s := range list {
		if s == find {
			return true
		}
	}
	return false
}

var allTemplates = map[string]string{
	"control": `
func (*{{.TypeName}}) isControl() {}
`,

	"visible": `
func (control *{{.TypeName}}) Visible() bool {
	return !control.hidden
}

func (control *{{.TypeName}}) SetVisible(v bool) *{{.TypeName}} {
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
`,

	"enabled": `
func (control *{{.TypeName}}) Enabled() bool {
	return !control.disabled
}

func (control *{{.TypeName}}) SetEnabled(e bool) *{{.TypeName}} {
	control.disabled = !e
	if control.handle != 0 {
		w32.EnableWindow(control.handle, e)
	}
	return control
}
`,

	"bounds": `
func (control *{{.TypeName}}) X() int {
	return control.x
}

func (control *{{.TypeName}}) Y() int {
	return control.y
}

func (control *{{.TypeName}}) Pos() (x, y int) {
	return control.x, control.y
}

func (control *{{.TypeName}}) Width() int {
	return control.width
}

func (control *{{.TypeName}}) Height() int {
	return control.height
}

func (control *{{.TypeName}}) Size() (width, height int) {
	return control.width, control.height
}

func (control *{{.TypeName}}) Bounds() (x, y, width, height int) {
	return control.x, control.y, control.width, control.height
}

func (control *{{.TypeName}}) SetX(x int) *{{.TypeName}} {
	return control.SetBounds(x, control.y, control.width, control.height)
}

func (control *{{.TypeName}}) SetY(y int) *{{.TypeName}} {
	return control.SetBounds(control.x, y, control.width, control.height)
}

func (control *{{.TypeName}}) SetPos(x, y int) *{{.TypeName}} {
	return control.SetBounds(x, y, control.width, control.height)
}

func (control *{{.TypeName}}) SetWidth(width int) *{{.TypeName}} {
	return control.SetBounds(control.x, control.y, width, control.height)
}

func (control *{{.TypeName}}) SetHeight(height int) *{{.TypeName}} {
	return control.SetBounds(control.x, control.y, control.width, height)
}

func (control *{{.TypeName}}) SetSize(width, height int) *{{.TypeName}} {
	return control.SetBounds(control.x, control.y, width, height)
}

func (control *{{.TypeName}}) SetBounds(x, y, width, height int) *{{.TypeName}} {
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
`,

	"text": `
func (control *{{.TypeName}}) Text() string {
	return control.text
}

func (control *{{.TypeName}}) SetText(text string) *{{.TypeName}} {
	control.text = text
	if control.handle != 0 {
		w32.SetWindowText(control.handle, control.text)
	}
	return control
}

func (control *{{.TypeName}}) Font() *Font {
	return control.font
}

func (control *{{.TypeName}}) SetFont(f *Font) *{{.TypeName}} {
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
`,

	"after_create": `
func (control *{{.TypeName}}) afterCreate(parent *Window) {
	control.parent = parent
	if control.hidden {
		w32.ShowWindow(control.handle, w32.SW_HIDE)
	}
	if control.disabled {
		w32.EnableWindow(control.handle, false)
	}
}

func (control *{{.TypeName}}) parentFontChanged() {}
`,

	"after_create_with_text": `
func (control *{{.TypeName}}) afterCreate(parent *Window) {
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

func (control *{{.TypeName}}) parentFontChanged() {
	if control.font == nil && control.parent.font != nil {
		w32.SendMessage(
			control.handle,
			w32.WM_SETFONT,
			uintptr(control.parent.font.handle),
			1,
		)
	}
}
`,
}
