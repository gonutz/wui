package main

import (
	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

func main() {
	theWindow := defaultWindow()

	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -11})
	w := wui.NewWindow()

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

	preview := wui.NewPaintbox()
	preview.SetBounds(200, 0, 400, 600)
	preview.AnchorTopAndBottom()
	preview.AnchorLeftAndRight()
	white := wui.RGB(255, 255, 255)
	black := wui.RGB(0, 0, 0)

	var xResizeArea, yResizeArea, xyResizeArea square

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
			if dragMode == dragX || dragMode == dragXY {
				theWindow.SetWidth(dragStartWidth + dx)
			}
			if dragMode == dragY || dragMode == dragXY {
				theWindow.SetHeight(dragStartHeight + dy)
			}
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
	return w
}
