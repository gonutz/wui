package main

import (
	"testing"

	"github.com/gonutz/check"
	"github.com/gonutz/wui"
)

func TestLoadEmptyDefaultWindow(t *testing.T) {
	code := `package main

import "github.com/gonutz/wui"

func main() {
	w := wui.NewWindow()
	w.Show()
}
`
	windows, err := extractWindowsFromCode(code)
	check.Eq(t, err, nil)
	check.Eq(t, len(windows), 1)
	w := windows[0]
	check.Eq(t, w.creationLineNumber, 6)
	check.Eq(t, w.window.Width(), wui.NewWindow().Width()) // Default value.
}

func TestParseErrorsCancelLoading(t *testing.T) {
	windows, err := extractWindowsFromCode("")
	check.Neq(t, err, nil)
	check.Eq(t, windows, nil)
}

func TestOnlySingleVariablesAssignedFromNewWindowAreConsidered(t *testing.T) {
	checkNone := func(codeInMain string) {
		t.Helper()
		code := `package main

import "github.com/gonutz/wui"

func main() {
` + codeInMain + `
}
`
		windows, err := extractWindowsFromCode(code)
		check.Eq(t, err, nil)
		check.Eq(t, windows, nil)
	}

	checkNone("                            // no window created at all")
	checkNone("wui.NewWindow()             // need a named variable")
	checkNone("_ = wui.NewWindow()         // need a named variable")
	checkNone("a, b := wui.NewWindow()     // too many on left side")
	checkNone("a := wui.NewWindow(), wat() // too many on right side")
	checkNone("w := wui.NewWindow          // function not called")
	checkNone("w := wui.NewWindow(1)       // NewWindows wants no arguments")
	checkNone("w := NewWindow()            // not called on wui package")
	checkNone("w := wat().NewWindow()      // not called on wui package")
	checkNone("w := schmui.NewWindow()     // not called on wui package")
	checkNone("w := wui.OldWidow()         // wrong function name")
	checkNone(`
	wui := somethingElse()
	w := wui.NewWindow()
	`)
}

func TestWuiMayBeImportedByAnotherName(t *testing.T) {
	code := `package main

import other "github.com/gonutz/wui"

func main() {
	w := other.NewWindow()
	w.Show()
}
`
	windows, err := extractWindowsFromCode(code)
	check.Eq(t, err, nil)
	check.Eq(t, len(windows), 1)
	w := windows[0]
	check.Eq(t, w.creationLineNumber, 6)
	check.Eq(t, w.window.Width(), wui.NewWindow().Width()) // Default value.
}

func TestWuiMustBeImported(t *testing.T) {
	code := `package main

func main() {
	w := other.NewWindow()
	w.Show()
}
`
	windows, err := extractWindowsFromCode(code)
	check.Neq(t, err, nil)
	check.Eq(t, windows, nil)
}

func TestWuiMustBeImportedExactlyOnce(t *testing.T) {
	code := `package main
	
import "github.com/gonutz/wui"
import other "github.com/gonutz/wui"

func main() {
	w := other.NewWindow()
	w.Show()
}
`
	windows, err := extractWindowsFromCode(code)
	check.Neq(t, err, nil)
	check.Eq(t, windows, nil)
}

// TODO What about
//
//     import . "github.com/gonutz/wui"
//
// ? Do we support it?

func TestWindowSettingsAreReadFromGoFile(t *testing.T) {
	code := `package main
	
import "github.com/gonutz/wui"

func main() {
	w := wui.NewWindow()
	w.SetSize(800, 600)
	w.SetAlpha(127)
	w.Show()
}
`
	windows, err := extractWindowsFromCode(code)
	check.Eq(t, err, nil)
	check.Eq(t, len(windows), 1)
	w := windows[0].window
	check.Eq(t, w.Width(), 800)
	check.Eq(t, w.Height(), 600)
	check.Eq(t, w.Alpha(), 127)
}

func TestMultipleWindowsMightBeCreatedWithTheSameVariable(t *testing.T) {
	code := `package main

import "github.com/gonutz/wui"

func main() {
	w := wui.NewWindow()
	w.SetSize(800, 600)
	w.SetAlpha(127)
	w.Show()

	w = wui.NewWindow()
	w.SetSize(1000, 555)
	w.SetAlpha(50)
	w.Show()
}
`
	windows, err := extractWindowsFromCode(code)
	check.Eq(t, err, nil)
	check.Eq(t, len(windows), 2)

	w := windows[0].window
	check.Eq(t, w.Width(), 800)
	check.Eq(t, w.Height(), 600)
	check.Eq(t, w.Alpha(), 127)

	w = windows[1].window
	check.Eq(t, w.Width(), 1000)
	check.Eq(t, w.Height(), 555)
	check.Eq(t, w.Alpha(), 50)
}
