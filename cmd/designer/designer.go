package main

import (
	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

func main() {
	theWindow := defaultWindow()

	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -11})
	w := wui.NewWindow()
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

	preview := wui.NewPaintbox()
	preview.SetBounds(200, 0, 400, 600)
	preview.AnchorTopAndBottom()
	preview.AnchorLeftAndRight()
	white := wui.RGB(255, 255, 255)
	black := wui.RGB(0, 0, 0)
	preview.SetOnPaint(func(c *wui.Canvas) {
		const xOffset, yOffset = 5, 5
		w, h := c.Size()
		c.FillRect(0, 0, w, h, white)

		width, height := theWindow.Size()
		clientWidth, clientHeight := theWindow.ClientSize()
		c.FillRect(xOffset, yOffset, width, height, wui.RGB(100, 200, 255))
		borderSize := (width - clientWidth) / 2
		topBorderSize := height - borderSize - clientHeight
		c.FillRect(
			xOffset+borderSize,
			yOffset+topBorderSize,
			clientWidth,
			clientHeight,
			wui.RGB(240, 240, 240),
		)
		c.SetFont(theWindow.Font())
		_, textH := c.TextExtent(theWindow.Title())
		c.TextOut(
			xOffset+borderSize,
			yOffset+(topBorderSize-textH)/2,
			theWindow.Title(),
			wui.RGB(0, 0, 0),
		)
		outlineRect := func(x, y, w, h int, inner, outer wui.Color) {
			c.FillRect(x, y, w, h, inner)
			c.DrawRect(x, y, w, h, outer)
		}
		outlineRect(xOffset+width-5, yOffset+height-5, 10, 10, white, black)
		outlineRect(xOffset+width/2-5, yOffset+height-5, 10, 10, white, black)
		outlineRect(xOffset+width-5, yOffset+height/2-5, 10, 10, white, black)
	})
	w.Add(preview)

	w.Show()
}

func defaultWindow() *wui.Window {
	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -11})
	w := wui.NewWindow()
	w.SetFont(font)
	w.SetTitle("Window")
	w.SetClientSize(300, 300)
	return w
}
