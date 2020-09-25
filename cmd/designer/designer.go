package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

const deleteTempDesignerFile = false

func main() {
	theWindow := defaultWindow()

	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -11})
	w := wui.NewWindow()

	appIcon := w32.LoadIcon(0, w32.MakeIntResource(w32.IDI_APPLICATION))
	appIconWidth := w32.GetSystemMetrics(w32.SM_CXICON)
	appIconHeight := w32.GetSystemMetrics(w32.SM_CYICON)
	appIconWidth, appIconHeight = 17, 17

	stdCursor := w.Cursor()
	upDownCursor := w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_SIZENS))
	leftRightCursor := w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_SIZEWE))
	diagonalCursor := w32.LoadCursor(0, w32.MakeIntResource(w32.IDC_SIZENWSE))

	w.SetFont(font)
	w.SetTitle("wui Designer")
	w.SetBackground(w32.GetSysColorBrush(w32.COLOR_BTNFACE))
	w.SetClientSize(800, 600)

	leftSlider := wui.NewPanel()
	leftSlider.SetBounds(195, -1, 5, 602)
	leftSlider.SetSingleLineBorder()
	leftSlider.SetVerticalAnchor(wui.AnchorMinAndMax)
	w.Add(leftSlider)

	rightSlider := wui.NewPanel()
	rightSlider.SetBounds(600, -1, 5, 602)
	rightSlider.SetSingleLineBorder()
	leftSlider.SetVerticalAnchor(wui.AnchorMinAndMax)
	rightSlider.SetHorizontalAnchor(wui.AnchorMax)
	w.Add(rightSlider)

	alphaText := wui.NewLabel()
	alphaText.SetText("Alpha")
	alphaText.SetRightAlign()
	alphaText.SetBounds(10, 10, 85, 25)
	w.Add(alphaText)
	alpha := wui.NewIntUpDown()
	alpha.SetBounds(105, 10, 85, 25)
	alpha.SetMinMaxValues(0, 255)
	w.Add(alpha)

	preview := wui.NewPaintbox()
	preview.SetBounds(200, 0, 400, 600)
	preview.SetHorizontalAnchor(wui.AnchorMinAndMax)
	preview.SetVerticalAnchor(wui.AnchorMinAndMax)
	white := wui.RGB(255, 255, 255)
	black := wui.RGB(0, 0, 0)

	var (
		// The ResizeAreas are the size drag points of the window.
		xResizeArea, yResizeArea, xyResizeArea square
		// clientX and Y is the top-left corner where the client area of the
		// window is drawn. The coordinates are relative to the application
		// window so we can use it in mouse events to find the relative mouse
		// position inside the window. TODO Say this with fewer "window"s.
		clientX, clientY int
		// active is the highlighted control whose properties are shown in the
		// tool bar.
		active node
	)

	alpha.SetOnValueChange(func(n int) {
		if w, ok := active.(*wui.Window); ok {
			w.SetAlpha(uint8(n))
		} else {
			panic("alpha value changed for non-Window")
		}
	})

	activate := func(newActive node) {
		active = newActive
		preview.Paint()
		switch x := active.(type) {
		case *wui.Window:
			alphaText.SetVisible(true)
			alpha.SetVisible(true)
			alpha.SetValue(int(x.Alpha()))
		default:
			alphaText.SetVisible(false)
			alpha.SetVisible(false)
		}
	}
	activate(theWindow)

	preview.SetOnPaint(func(c *wui.Canvas) {
		const xOffset, yOffset = 5, 5
		width, height := theWindow.Size()
		clientWidth, clientHeight := theWindow.ClientSize()
		borderSize := (width - clientWidth) / 2
		topBorderSize := height - borderSize - clientHeight
		clientX = xOffset + borderSize
		clientY = yOffset + topBorderSize
		client := makeOffsetDrawer(c, clientX, clientY)

		// Clear client area.
		c.FillRect(
			clientX,
			clientY,
			clientWidth,
			clientHeight,
			wui.RGB(240, 240, 240),
		)

		xResizeArea = square{
			x:    xOffset + width - 6,
			y:    yOffset + height/2 - 6,
			size: 12,
		}
		yResizeArea = square{
			x:    xOffset + width/2 - 6,
			y:    yOffset + height - 6,
			size: 12,
		}
		xyResizeArea = square{
			x:    xOffset + width - 6,
			y:    yOffset + height - 6,
			size: 12,
		}

		drawContainer(theWindow, client)

		// Highlight the currently selected child control.
		if active != theWindow {
			x, y, w, h := active.Bounds()
			parent := active.Parent()
			for parent != theWindow {
				dx, dy, _, _ := parent.Bounds()
				x += dx
				y += dy
				parent = parent.Parent()
			}
			client.DrawRect(x-1, y-1, w+2, h+2, wui.RGB(255, 0, 255))
			client.DrawRect(x-2, y-2, w+4, h+4, wui.RGB(255, 0, 255))
		}

		// Draw the window border, icon and title.
		borderColor := wui.RGB(100, 200, 255)
		c.FillRect(xOffset, yOffset, width, topBorderSize, borderColor)
		c.FillRect(xOffset, yOffset, borderSize, height, borderColor)
		c.FillRect(xOffset, yOffset+height-borderSize, width, borderSize, borderColor)
		c.FillRect(xOffset+width-borderSize, yOffset, borderSize, height, borderColor)

		w32.DrawIconEx(
			c.Handle(),
			xOffset+borderSize,
			yOffset+(topBorderSize-appIconHeight)/2,
			appIcon,
			appIconWidth, appIconHeight,
			0, 0, w32.DI_NORMAL,
		)

		c.SetFont(theWindow.Font())
		_, textH := c.TextExtent(theWindow.Title())
		c.TextOut(
			xOffset+borderSize+appIconWidth+5,
			yOffset+(topBorderSize-textH)/2,
			theWindow.Title(),
			wui.RGB(0, 0, 0),
		)

		// Clear the background behind the window.
		w, h := c.Size()
		c.FillRect(0, 0, w, yOffset, white)
		c.FillRect(0, 0, xOffset, h, white)
		right := xOffset + width
		c.FillRect(right, 0, w-right, h, white)
		bottom := yOffset + height
		c.FillRect(0, bottom, w, h-bottom, white)

		// Add drag markers to resize window.
		outlineSquare := func(s square) {
			c.FillRect(s.x, s.y, s.size, s.size, white)
			c.DrawRect(s.x, s.y, s.size, s.size, black)
		}
		outlineSquare(xResizeArea)
		outlineSquare(yResizeArea)
		outlineSquare(xyResizeArea)
	})
	w.Add(preview)

	var dragMode int
	const (
		dragNone = 0
		dragX    = 1
		dragY    = 2
		dragXY   = 3
	)

	var dragStartX, dragStartY, dragStartWidth, dragStartHeight int

	w.SetOnMouseMove(func(x, y int) {
		if dragMode == dragNone {
			x -= preview.X()
			y -= preview.Y()
			if xResizeArea.contains(x, y) {
				w.SetCursor(leftRightCursor)
			} else if yResizeArea.contains(x, y) {
				w.SetCursor(upDownCursor)
			} else if xyResizeArea.contains(x, y) {
				w.SetCursor(diagonalCursor)
			} else {
				w.SetCursor(stdCursor)
			}
		} else {
			dx, dy := x-dragStartX, y-dragStartY
			newW := dragStartWidth
			newH := dragStartHeight
			if dragMode == dragX || dragMode == dragXY {
				newW += dx
			}
			if dragMode == dragY || dragMode == dragXY {
				newH += dy
			}
			// TODO Refactor this, we want to go through SetBounds for now since
			// only it does updating children by anchor at the moment.
			theWindow.SetBounds(theWindow.X(), theWindow.Y(), newW, newH)
			preview.Paint()
		}
	})

	w.SetOnMouseDown(func(button wui.MouseButton, x, y int) {
		if button == wui.MouseButtonLeft {
			dragStartX = x
			dragStartY = y
			dragStartWidth, dragStartHeight = theWindow.Size()
			if xResizeArea.contains(x-preview.X(), y-preview.Y()) {
				dragMode = dragX
			} else if yResizeArea.contains(x-preview.X(), y-preview.Y()) {
				dragMode = dragY
			} else if xyResizeArea.contains(x-preview.X(), y-preview.Y()) {
				dragMode = dragXY
			} else {
				newActive := findControlAt(
					theWindow,
					x-(preview.X()+clientX),
					y-(preview.Y()+clientY),
				)
				if newActive != active {
					activate(newActive)
				}
			}
		}
	})

	w.SetOnMouseUp(func(button wui.MouseButton, x, y int) {
		if button == wui.MouseButtonLeft {
			dragMode = dragNone
		}
	})

	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl, Rune: 'R'}, func() {
		var wuiCode string
		line := func(format string, a ...interface{}) {
			format = "\t" + format
			if wuiCode != "" {
				format = "\n" + format
			}
			wuiCode += fmt.Sprintf(format, a...)
		}

		line("w := wui.NewWindow()")
		line("w.SetTitle(%q)", theWindow.Title())
		line("w.SetSize(%d, %d)", theWindow.Width(), theWindow.Height())
		if theWindow.Alpha() != 255 {
			line("w.SetAlpha(%d)", theWindow.Alpha())
		}
		font := theWindow.Font()
		if font != nil {
			line("font, _ := wui.NewFont(wui.FontDesc{")
			if font.Desc.Name != "" {
				line("Name: %q,", font.Desc.Name)
			}
			if font.Desc.Height != 0 {
				line("Height: %d,", font.Desc.Height)
			}
			if font.Desc.Bold {
				line("Bold: true,")
			}
			if font.Desc.Italic {
				line("Italic: true,")
			}
			if font.Desc.Underlined {
				line("Underlined: true,")
			}
			if font.Desc.StrikedOut {
				line("StrikedOut: true,")
			}
			line("})")
			line("w.SetFont(font)")
		}
		writeContainer(theWindow, "w", line)
		line("w.Show()")

		mainProgram := `//+build ignore

package main

import "github.com/gonutz/wui"

func main() {
` + wuiCode + `
}
`
		const fileName = "wui_designer_temp_file.go"
		err := ioutil.WriteFile(fileName, []byte(mainProgram), 0666)
		if err != nil {
			wui.MessageBoxError("Error", err.Error())
			return
		}
		if deleteTempDesignerFile {
			defer os.Remove(fileName)
		}
		// TODO This blocks and freezes the designer:
		output, err := exec.Command("go", "run", fileName).CombinedOutput()
		if err != nil {
			wui.MessageBoxError("Error", err.Error()+"\r\n"+string(output))
		} else if len(output) > 0 {
			wui.MessageBoxInfo("go output", string(output))
		}
	})

	w.SetShortcut(wui.ShortcutKeys{Rune: 27}, w.Close) // TODO for debugging

	w.Maximize()
	w.Show()
}

type square struct {
	x, y, size int
}

func (s square) contains(x, y int) bool {
	return x >= s.x && y >= s.y && x < s.x+s.size && y < s.y+s.size
}

func defaultWindow() *wui.Window {
	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -11})
	w := wui.NewWindow()
	w.SetFont(font)
	w.SetTitle("Window")
	w.SetClientSize(300, 300)
	// TODO
	b := wui.NewButton()
	b.SetBounds(10, 10, 75, 25)
	b.SetText("TopLeft")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(215, 265, 75, 25)
	b.SetHorizontalAnchor(wui.AnchorMax)
	b.SetVerticalAnchor(wui.AnchorMax)
	b.SetText("BottomRight")
	w.Add(b)

	p := wui.NewPanel()
	p.SetBounds(100, 100, 100, 100)
	p.SetHorizontalAnchor(wui.AnchorMinAndMax)
	p.SetVerticalAnchor(wui.AnchorMinAndMax)
	w.Add(p)

	b = wui.NewButton()
	b.SetBounds(10, 100-30, 80, 25)
	b.SetVerticalAnchor(wui.AnchorMax)
	b.SetText("In here")
	p.Add(b)
	//
	return w
}

func findControlAt(parent wui.Container, x, y int) node {
	for _, child := range parent.Children() {
		if contains(child, x, y) {
			if container, ok := child.(wui.Container); ok {
				dx, dy, _, _ := container.Bounds()
				return findControlAt(container, x-dx, y-dy)
			}
			return child
		}
	}
	return parent
}

func contains(b bounder, atX, atY int) bool {
	x, y, w, h := b.Bounds()
	return atX >= x && atY >= y && atX < x+w && atY < y+h
}

type bounder interface {
	Bounds() (x, y, width, height int)
}

type drawer interface {
	Line(x1, y1, x2, y2 int, color wui.Color)
	DrawRect(x, y, w, h int, color wui.Color)
	FillRect(x, y, w, h int, color wui.Color)
	TextRectFormat(x, y, w, h int, s string, format wui.Format, color wui.Color)
}

func makeOffsetDrawer(base drawer, dx, dy int) drawer {
	return &offsetDrawer{base: base, dx: dx, dy: dy}
}

type offsetDrawer struct {
	base   drawer
	dx, dy int
}

func (d *offsetDrawer) DrawRect(x, y, w, h int, color wui.Color) {
	d.base.DrawRect(x+d.dx, y+d.dy, w, h, color)
}

func (d *offsetDrawer) FillRect(x, y, w, h int, color wui.Color) {
	d.base.FillRect(x+d.dx, y+d.dy, w, h, color)
}

func (d *offsetDrawer) TextRectFormat(
	x, y, w, h int, s string, format wui.Format, color wui.Color,
) {
	d.base.TextRectFormat(x+d.dx, y+d.dy, w, h, s, format, color)
}

func (d *offsetDrawer) Line(x1, y1, x2, y2 int, color wui.Color) {
	d.base.Line(x1+d.dx, y1+d.dy, x2+d.dx, y2+d.dy, color)
}

func drawContainer(container wui.Container, d drawer) {
	for _, child := range container.Children() {
		if button, ok := child.(*wui.Button); ok {
			x, y, w, h := button.Bounds()
			d.FillRect(x+1, y+1, w-2, h-2, wui.RGB(173, 173, 173))
			d.FillRect(x+2, y+2, w-4, h-4, wui.RGB(225, 225, 225))
			d.TextRectFormat(x, y, w, h, button.Text(), wui.FormatCenter, wui.RGB(0, 0, 0))
		} else if panel, ok := child.(*wui.Panel); ok {
			x, y, w, h := panel.Bounds()
			border := "none"
			switch border {
			case "none":
				d.DrawRect(x, y, w, h, wui.RGB(230, 230, 230))
			case "single":
				d.DrawRect(x, y, w, h, wui.RGB(100, 100, 100))
			case "raised":
				d.Line(x, y, x+w, y, wui.RGB(227, 227, 227))
				d.Line(x, y, x, y+h, wui.RGB(227, 227, 227))
				d.Line(x+w-1, y, x+w-1, y+h, wui.RGB(105, 105, 105))
				d.Line(x, y+h-1, x+w, y+h-1, wui.RGB(105, 105, 105))
				d.Line(x+1, y+1, x+w-1, y+1, wui.RGB(255, 255, 255))
				d.Line(x+1, y+1, x+1, y+h-1, wui.RGB(255, 255, 255))
				d.Line(x+w-2, y+1, x+w-2, y+h-1, wui.RGB(160, 160, 160))
				d.Line(x+1, y+h-2, x+w-1, y+h-2, wui.RGB(160, 160, 160))
			case "sunken":
				d.Line(x, y, x+w, y, wui.RGB(160, 160, 160))
				d.Line(x, y, x, y+h, wui.RGB(160, 160, 160))
				d.Line(x+w-1, y, x+w-1, y+h, wui.RGB(255, 255, 255))
				d.Line(x, y+h-1, x+w, y+h-1, wui.RGB(255, 255, 255))
			case "sunken_thick":
				d.Line(x, y, x+w, y, wui.RGB(160, 160, 160))
				d.Line(x, y, x, y+h, wui.RGB(160, 160, 160))
				d.Line(x+w-1, y, x+w-1, y+h, wui.RGB(255, 255, 255))
				d.Line(x, y+h-1, x+w, y+h-1, wui.RGB(255, 255, 255))
				d.Line(x+1, y+1, x+w-1, y+1, wui.RGB(105, 105, 105))
				d.Line(x+1, y+1, x+1, y+h-1, wui.RGB(105, 105, 105))
				d.Line(x+w-2, y+1, x+w-2, y+h-1, wui.RGB(227, 227, 227))
				d.Line(x+1, y+h-2, x+w-1, y+h-2, wui.RGB(227, 227, 227))
			}
			drawContainer(panel, makeOffsetDrawer(d, panel.X(), panel.Y()))
		} else {
			panic("unhandled child control")
		}
	}
}

func writeContainer(c wui.Container, parent string, line func(format string, a ...interface{})) {
	for i, child := range c.Children() {
		name := fmt.Sprintf("%s_child%d", parent, i)
		do := func(format string, a ...interface{}) {
			line(name+format, a...)
		}
		if button, ok := child.(*wui.Button); ok {
			do(" := wui.NewButton()")
			do(".SetBounds(%d, %d, %d, %d)", button.X(), button.Y(), button.Width(), button.Height())
			do(".SetText(%q)", button.Text())
			if !button.Enabled() {
				do(".SetEnabled(false)")
			}
			if !button.Visible() {
				do(".SetVisible(false)")
			}
			line("%s.Add(%s)", parent, name)
		} else if panel, ok := child.(*wui.Panel); ok {
			do(" := wui.NewPanel()")
			// TODO
			//do(".SetNoBorder()")
			//do(".SetRaisedBorder()")
			//do(".SetSingleLineBorder()")
			//do(".SetSunkenBorder()")
			do(".SetSunkenThickBorder()")
			do(".SetBounds(%d, %d, %d, %d)", panel.X(), panel.Y(), panel.Width(), panel.Height())
			if !panel.Enabled() {
				do(".SetEnabled(false)")
			}
			if !panel.Visible() {
				do(".SetVisible(false)")
			}
			// TODO We would want to fill in the panel content here, before
			// adding the panel to the parent, but there is a bug in Panel.Add,
			// see the comment there.
			line("%s.Add(%s)", parent, name)
			writeContainer(panel, name, line)
		} else {
			panic("unhandled child control")
		}
	}
}

type node interface {
	Parent() wui.Container
	Bounds() (x, y, width, height int)
}
