package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

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
	leftSlider.AnchorTopAndBottom()
	w.Add(leftSlider)

	rightSlider := wui.NewPanel()
	rightSlider.SetBounds(600, -1, 5, 602)
	rightSlider.SetSingleLineBorder()
	rightSlider.AnchorTopAndBottom()
	rightSlider.AnchorRight()
	w.Add(rightSlider)

	l := wui.NewLabel()
	l.SetText("Border Style")
	l.SetRightAlign()
	l.SetBounds(5, 0, 60, 30)
	w.Add(l)
	borderStyle := wui.NewCombobox()
	borderStyle.SetBounds(70, 5, 120, 30)
	borderStyle.Add("Sizeable")
	borderStyle.Add("Dialog")
	borderStyle.Add("Single")
	borderStyle.Add("None")
	borderStyle.SetSelectedIndex(0)
	w.Add(borderStyle)

	preview := wui.NewPaintbox()
	preview.SetBounds(200, 0, 400, 600)
	preview.AnchorTopAndBottom()
	preview.AnchorLeftAndRight()
	white := wui.RGB(255, 255, 255)
	black := wui.RGB(0, 0, 0)

	var xResizeArea, yResizeArea, xyResizeArea square

	preview.SetOnPaint(func(c *wui.Canvas) {
		const xOffset, yOffset = 5, 5
		width, height := theWindow.Size()
		clientWidth, clientHeight := theWindow.ClientSize()
		borderSize := (width - clientWidth) / 2
		topBorderSize := height - borderSize - clientHeight
		clientX := xOffset + borderSize
		clientY := yOffset + topBorderSize

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

		for _, child := range theWindow.Children() {
			if button, ok := child.(*wui.Button); ok {
				x, y, w, h := button.Bounds()
				c.FillRect(clientX+x+1, clientY+y+1, w-2, h-2, wui.RGB(173, 173, 173))
				c.FillRect(clientX+x+2, clientY+y+2, w-4, h-4, wui.RGB(225, 225, 225))
				c.TextRectFormat(clientX+x, clientY+y, w, h, button.Text(), wui.FormatCenter, black)
			} else {
				panic("unhandled child control")
			}
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
			// TODO Refactor this, we want to go through SetBounds for now.
			theWindow.SetBounds(theWindow.X(), theWindow.Y(), newW, newH)
			preview.Paint()
		}
	})
	w.SetOnMouseDown(func(button wui.MouseButton, x, y int) {
		if button == wui.MouseButtonLeft {
			dragStartX = x
			dragStartY = y
			dragStartWidth, dragStartHeight = theWindow.Size()
			x -= preview.X()
			y -= preview.Y()
			if xResizeArea.contains(x, y) {
				dragMode = dragX
			} else if yResizeArea.contains(x, y) {
				dragMode = dragY
			} else if xyResizeArea.contains(x, y) {
				dragMode = dragXY
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
		for i, child := range theWindow.Children() {
			name := fmt.Sprintf("child%d", i)
			if button, ok := child.(*wui.Button); ok {
				line("%s := wui.NewButton()", name)
				line("%s.SetBounds(%d, %d, %d, %d)", name, button.X(), button.Y(), button.Width(), button.Height())
				line("%s.SetText(%q)", name, button.Text())
				if !button.Enabled() {
					line("%s.SetEnabled(false)", name)
				}
				if !button.Visible() {
					line("%s.SetVisible(false)", name)
				}
				line("w.Add(%s)", name)
			} else {
				panic("unhandled child control")
			}
		}
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
		defer os.Remove(fileName)
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
	b.AnchorRight()
	b.AnchorBottom()
	b.SetText("BottomRight")
	w.Add(b)
	//
	return w
}
