package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

// TODO Handle negative widths/heights, they display in the preview but the real
// window does not allow them.

// TODO Clamp the drawing canvas for each container.

var names = make(map[interface{}]string)

func main() {
	const (
		idleMouse = iota
		addingControl
	)
	mouseMode := idleMouse

	theWindow := defaultWindow()
	names[theWindow] = "w"

	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -11})
	w := wui.NewWindow()

	menu := wui.NewMainMenu()
	fileMenu := wui.NewMenu("&File")
	fileOpenMenu := wui.NewMenuString("&Open File...\tCtrl+O")
	fileMenu.Add(fileOpenMenu)
	fileSaveMenu := wui.NewMenuString("&Save File\tCtrl+S")
	fileMenu.Add(fileSaveMenu)
	fileSaveAsMenu := wui.NewMenuString("Save File &As...\tCtrl+Shift+S")
	fileMenu.Add(fileSaveAsMenu)
	menu.Add(fileMenu)
	w.SetMenu(menu)

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
	w.SetInnerSize(800, 600)

	leftSlider := wui.NewPanel()
	leftSlider.SetBounds(195, -1, 5, 602)
	leftSlider.SetBorderStyle(wui.PanelBorderSingleLine)
	leftSlider.SetVerticalAnchor(wui.AnchorMinAndMax)
	w.Add(leftSlider)

	rightSlider := wui.NewPanel()
	rightSlider.SetBounds(600, -1, 5, 602)
	rightSlider.SetBorderStyle(wui.PanelBorderSingleLine)
	rightSlider.SetVerticalAnchor(wui.AnchorMinAndMax)
	rightSlider.SetHorizontalAnchor(wui.AnchorMax)
	w.Add(rightSlider)

	buttonTemplate := wui.NewButton()
	buttonTemplate.SetText("Button")
	buttonTemplate.SetBounds(9, 11, 85, 25)

	checkBoxTemplate := wui.NewCheckBox()
	checkBoxTemplate.SetText("CheckBox")
	checkBoxTemplate.SetChecked(true)
	checkBoxTemplate.SetBounds(10, 44, 100, 17)

	radioButtonTemplate := wui.NewRadioButton()
	radioButtonTemplate.SetText("RadioButton")
	radioButtonTemplate.SetChecked(true)
	radioButtonTemplate.SetBounds(10, 67, 100, 17)

	allTemplates := []wui.Control{
		buttonTemplate,
		checkBoxTemplate,
		radioButtonTemplate,
	}

	var highlightedTemplate, controlToAdd wui.Control
	var templateDx, templateDy int

	palette := wui.NewPaintBox()
	palette.SetBounds(605, 0, 195, 600)
	palette.SetHorizontalAnchor(wui.AnchorMax)
	palette.SetVerticalAnchor(wui.AnchorMinAndMax)
	palette.SetOnPaint(func(c *wui.Canvas) {
		w, h := c.Size()
		c.FillRect(0, 0, w, h, wui.RGB(240, 240, 240))
		// Draw all template controls.
		drawButton(buttonTemplate, c)
		drawCheckBox(checkBoxTemplate, c)
		drawRadioButton(radioButtonTemplate, c)
		// Highlight what is under the mouse.
		if highlightedTemplate != nil {
			x, y, w, h := highlightedTemplate.Bounds()
			c.DrawRect(x-1, y-1, w+2, h+2, wui.RGB(255, 0, 255))
			c.DrawRect(x-2, y-2, w+4, h+4, wui.RGB(255, 0, 255))
		}
	})
	palette.SetOnMouseMove(func(x, y int) {
		oldHighlight := highlightedTemplate
		highlightedTemplate = nil
		for _, c := range allTemplates {
			if contains(c, x, y) {
				highlightedTemplate = c
			}
		}
		if highlightedTemplate != oldHighlight {
			palette.Paint()
		}
	})
	w.Add(palette)

	nameText := wui.NewLabel()
	nameText.SetText("Variable Name")
	nameText.SetRightAlign()
	nameText.SetBounds(10, 10, 85, 25)
	w.Add(nameText)
	name := wui.NewEditLine()
	name.SetBounds(105, 10, 85, 25)
	w.Add(name)

	alphaText := wui.NewLabel()
	alphaText.SetText("Alpha")
	alphaText.SetRightAlign()
	alphaText.SetBounds(10, 40, 85, 25)
	w.Add(alphaText)
	alpha := wui.NewIntUpDown()
	alpha.SetBounds(105, 40, 85, 25)
	alpha.SetMinMaxValues(0, 255)
	w.Add(alpha)

	anchorToIndex := map[wui.Anchor]int{
		wui.AnchorMin:          0,
		wui.AnchorMax:          1,
		wui.AnchorCenter:       2,
		wui.AnchorMinAndMax:    3,
		wui.AnchorMinAndCenter: 4,
		wui.AnchorMaxAndCenter: 5,
	}
	indexToAnchor := make(map[int]wui.Anchor)
	for a, i := range anchorToIndex {
		indexToAnchor[i] = a
	}

	hAnchorText := wui.NewLabel()
	hAnchorText.SetText("Horizontal Anchor")
	hAnchorText.SetRightAlign()
	hAnchorText.SetBounds(10, 40, 85, 25)
	w.Add(hAnchorText)
	hAnchor := wui.NewComboBox()
	hAnchor.Add("Left")
	hAnchor.Add("Right")
	hAnchor.Add("Center")
	hAnchor.Add("Left+Right")
	hAnchor.Add("Left+Center")
	hAnchor.Add("Right+Center")
	hAnchor.SetBounds(105, 40, 85, 25)
	w.Add(hAnchor)

	vAnchorText := wui.NewLabel()
	vAnchorText.SetText("Vertical Anchor")
	vAnchorText.SetRightAlign()
	vAnchorText.SetBounds(10, 70, 85, 25)
	w.Add(vAnchorText)
	vAnchor := wui.NewComboBox()
	vAnchor.Add("Top")
	vAnchor.Add("Bottom")
	vAnchor.Add("Center")
	vAnchor.Add("Top+Bottom")
	vAnchor.Add("Top+Center")
	vAnchor.Add("Bottom+Center")
	vAnchor.SetBounds(105, 70, 85, 25)
	w.Add(vAnchor)

	checked := wui.NewCheckBox()
	checked.SetText("Checked")
	checked.SetBounds(105, 100, 85, 17)
	w.Add(checked)

	panelBorderToIndex := map[wui.PanelBorderStyle]int{
		wui.PanelBorderNone:        0,
		wui.PanelBorderSingleLine:  1,
		wui.PanelBorderRaised:      2,
		wui.PanelBorderSunken:      3,
		wui.PanelBorderSunkenThick: 4,
	}
	indexToPanelBorder := make(map[int]wui.PanelBorderStyle)
	for a, i := range panelBorderToIndex {
		indexToPanelBorder[i] = a
	}

	panelBorderStyleText := wui.NewLabel()
	panelBorderStyleText.SetText("Border Style")
	panelBorderStyleText.SetRightAlign()
	panelBorderStyleText.SetBounds(10, 100, 85, 25)
	w.Add(panelBorderStyleText)
	panelBorderStyle := wui.NewComboBox()
	panelBorderStyle.Add("None")
	panelBorderStyle.Add("Single")
	panelBorderStyle.Add("Raised")
	panelBorderStyle.Add("Sunken")
	panelBorderStyle.Add("Sunken Thick")
	panelBorderStyle.SetBounds(105, 100, 85, 25)
	w.Add(panelBorderStyle)

	preview := wui.NewPaintBox()
	preview.SetBounds(200, 0, 400, 600)
	preview.SetHorizontalAnchor(wui.AnchorMinAndMax)
	preview.SetVerticalAnchor(wui.AnchorMinAndMax)
	white := wui.RGB(255, 255, 255)
	black := wui.RGB(0, 0, 0)

	var (
		// The ResizeAreas are the size drag points of the window.
		xResizeArea, yResizeArea, xyResizeArea rectangle
		// innerX and Y is the top-left corner where the inner area of the
		// window is drawn. The coordinates are relative to the application
		// window so we can use it in mouse events to find the relative mouse
		// position inside the window. TODO Say this with fewer "window"s.
		innerX, innerY int
		// active is the highlighted control whose properties are shown in the
		// tool bar.
		active node
	)

	name.SetOnTextChange(func() {
		names[active] = name.Text()
	})
	alpha.SetOnValueChange(func(n int) {
		if w, ok := active.(*wui.Window); ok {
			w.SetAlpha(uint8(n))
		} else {
			panic("alpha value changed for non-Window")
		}
	})
	hAnchor.SetOnChange(func(i int) {
		if c, ok := active.(wui.Control); ok {
			c.SetHorizontalAnchor(indexToAnchor[i])
		} else {
			panic("anchor set on non-Control")
		}
	})
	vAnchor.SetOnChange(func(i int) {
		if c, ok := active.(wui.Control); ok {
			c.SetVerticalAnchor(indexToAnchor[i])
		} else {
			panic("anchor set on non-Control")
		}
	})
	checked.SetOnChange(func(check bool) {
		if r, ok := active.(*wui.RadioButton); ok {
			r.SetChecked(check)
			preview.Paint()
		} else if c, ok := active.(*wui.CheckBox); ok {
			c.SetChecked(check)
			preview.Paint()
		} else {
			panic("check is for radio buttons and check boxes only")
		}
	})
	panelBorderStyle.SetOnChange(func(i int) {
		if p, ok := active.(*wui.Panel); ok {
			p.SetBorderStyle(indexToPanelBorder[i])
			preview.Paint()
		} else {
			panic("panel border style only for panels")
		}
	})

	activate := func(newActive node) {
		active = newActive
		name.SetText(names[active])

		_, isWindow := active.(*wui.Window)
		_, isCheckBox := active.(*wui.CheckBox)
		_, isRadioButton := active.(*wui.RadioButton)
		_, isPanel := active.(*wui.Panel)

		alphaText.SetVisible(isWindow)
		alpha.SetVisible(isWindow)
		hAnchorText.SetVisible(!isWindow)
		hAnchor.SetVisible(!isWindow)
		vAnchorText.SetVisible(!isWindow)
		vAnchor.SetVisible(!isWindow)
		checked.SetVisible(isCheckBox || isRadioButton)
		panelBorderStyleText.SetVisible(isPanel)
		panelBorderStyle.SetVisible(isPanel)

		if isWindow {
			alpha.SetValue(int(active.(*wui.Window).Alpha()))
		} else {
			h, v := active.(wui.Control).Anchors()
			hAnchor.SetSelectedIndex(anchorToIndex[h])
			vAnchor.SetSelectedIndex(anchorToIndex[v])
		}
		if isCheckBox {
			checked.SetChecked(active.(*wui.CheckBox).Checked())
		}
		if isRadioButton {
			checked.SetChecked(active.(*wui.RadioButton).Checked())
		}
		if isPanel {
			b := active.(*wui.Panel).BorderStyle()
			panelBorderStyle.SetSelectedIndex(panelBorderToIndex[b])
		}

		preview.Paint()
	}
	activate(theWindow)

	const xOffset, yOffset = 5, 5
	preview.SetOnPaint(func(c *wui.Canvas) {
		// TODO Use Bounds and InnerBounds to determine the real borders, they
		// do not necessarily have to be the same left/right and bottom.
		width, height := theWindow.Size()
		innerWidth, innerHeight := theWindow.InnerSize()
		borderSize := (width - innerWidth) / 2
		topBorderSize := height - borderSize - innerHeight
		innerX = xOffset + borderSize
		innerY = yOffset + topBorderSize
		inner := makeOffsetDrawer(c, innerX, innerY)

		// Clear inner area.
		c.FillRect(innerX, innerY, innerWidth, innerHeight, wui.RGB(240, 240, 240))

		xResizeArea = rectangle{
			x: xOffset + width - 6,
			y: yOffset + height/2 - 12,
			w: 12,
			h: 24,
		}
		yResizeArea = rectangle{
			x: xOffset + width/2 - 12,
			y: yOffset + height - 6,
			w: 24,
			h: 12,
		}
		xyResizeArea = rectangle{
			x: xOffset + width - 6,
			y: yOffset + height - 6,
			w: 12,
			h: 12,
		}

		// Draw all the window contents.
		drawContainer(theWindow, inner)

		// Draw the window border, icon and title.
		borderColor := wui.RGB(100, 200, 255)
		c.FillRect(xOffset, yOffset, width, topBorderSize, borderColor)
		c.FillRect(xOffset, yOffset, borderSize, height, borderColor)
		c.FillRect(xOffset, yOffset+height-borderSize, width, borderSize, borderColor)
		c.FillRect(xOffset+width-borderSize, yOffset, borderSize, height, borderColor)

		c.SetFont(theWindow.Font())
		_, textH := c.TextExtent(theWindow.Title())
		c.TextOut(
			xOffset+borderSize+appIconWidth+5,
			yOffset+(topBorderSize-textH)/2,
			theWindow.Title(),
			wui.RGB(0, 0, 0),
		)

		// TODO Handle combinations of borders and top-right corner buttons. For
		// now we just draw minimize, maximize and close buttons.
		{
			w := topBorderSize
			h := w - 8
			y := yOffset + 4
			right := xOffset + width - borderSize
			x0 := right - 3*w - 2
			x1 := right - 2*w - 1
			x2 := right - 1*w - 0
			iconSize := h / 2
			// Minimize button.
			c.FillRect(x0, y, w, h, wui.RGB(240, 240, 240))
			cx := x0 + (w-iconSize)/2
			cy := y + h - 1 - (iconSize+1)/2
			c.Line(cx, cy, cx+iconSize, cy, wui.RGB(0, 0, 0))
			// Maximize button.
			c.FillRect(x1, y, w, h, wui.RGB(240, 240, 240))
			cx = x1 + (w-iconSize)/2
			cy = y + (h-iconSize)/2
			c.DrawRect(cx, cy, iconSize, iconSize, wui.RGB(0, 0, 0))
			// Close button.
			c.FillRect(x2, y, w, h, wui.RGB(255, 128, 128))
			cx = x2 + (w-iconSize)/2
			cy = y + (h-iconSize)/2
			c.Line(cx, cy, cx+iconSize, cy+iconSize, wui.RGB(0, 0, 0))
			c.Line(cx, cy+iconSize-1, cx+iconSize, cy-1, wui.RGB(0, 0, 0))
		}

		w32.DrawIconEx(
			c.Handle(),
			xOffset+borderSize,
			yOffset+(topBorderSize-appIconHeight)/2,
			appIcon,
			appIconWidth, appIconHeight,
			0, 0, w32.DI_NORMAL,
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
		outlineSquare := func(r rectangle) {
			c.FillRect(r.x, r.y, r.w, r.h, white)
			c.DrawRect(r.x, r.y, r.w, r.h, black)
		}
		outlineSquare(xResizeArea)
		outlineSquare(yResizeArea)
		outlineSquare(xyResizeArea)

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
			inner.DrawRect(x-1, y-1, w+2, h+2, wui.RGB(255, 0, 255))
			inner.DrawRect(x-2, y-2, w+4, h+4, wui.RGB(255, 0, 255))
		}

		if controlToAdd != nil {
			drawControl(controlToAdd, c)
		}
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
		if mouseMode == addingControl {
			if contains(preview, x, y) {
				_, _, w, h := controlToAdd.Bounds()
				relX := x - preview.X()
				relY := y - preview.Y()
				if false {
					// TODO Align to some nice-looking grid unless Ctrl is held
					// down for example. NOTE that this right now contains a
					// bug, relX is not in window client coordinates, it is
					// relative to the preview paint box and thus we can never
					// get to 0,0 with this.
					relX = relX / 5 * 5
					relY = relY / 5 * 5
				}
				relX += templateDx
				relY += templateDy
				controlToAdd.SetBounds(relX, relY, w, h)
			}
			preview.Paint()
		} else if dragMode == dragNone {
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
			if contains(palette, x, y) && highlightedTemplate != nil {
				controlToAdd = cloneControl(highlightedTemplate)
				hx, hy, _, _ := highlightedTemplate.Bounds()
				templateDx = hx - (x - palette.X())
				templateDy = hy - (y - palette.Y())
				mouseMode = addingControl
				preview.Paint()
			} else if mouseMode == addingControl {
				// TODO Find the sub-container that this is to be placed in.
				innerX, innerY, _, _ := theWindow.InnerBounds()
				outerX, outerY, _, _ := theWindow.Bounds()
				x, y, w, h := controlToAdd.Bounds()
				controlToAdd.SetBounds(
					x-(xOffset+innerX-outerX),
					y-(yOffset+innerY-outerY),
					w, h,
				)
				theWindow.Add(controlToAdd)
				activate(controlToAdd)
				controlToAdd = nil
				mouseMode = idleMouse
				preview.Paint()
			} else {
				dragStartX = x
				dragStartY = y
				dragStartWidth, dragStartHeight = theWindow.Size()
				windowArea := rectangle{
					x: preview.X() + xOffset,
					y: preview.Y() + yOffset,
					w: theWindow.Width(),
					h: theWindow.Height(),
				}
				if xResizeArea.contains(x-preview.X(), y-preview.Y()) {
					// TODO Combine dragMode and mouseMode?
					dragMode = dragX
				} else if yResizeArea.contains(x-preview.X(), y-preview.Y()) {
					dragMode = dragY
				} else if xyResizeArea.contains(x-preview.X(), y-preview.Y()) {
					dragMode = dragXY
				} else if windowArea.contains(x, y) {
					newActive := findControlAt(
						theWindow,
						x-(preview.X()+innerX),
						y-(preview.Y()+innerY),
					)
					if newActive != active {
						activate(newActive)
					}
				}
			}
		}
	})

	w.SetOnMouseUp(func(button wui.MouseButton, x, y int) {
		if button == wui.MouseButtonLeft {
			dragMode = dragNone
		}
	})

	workingPath := ""
	setWorkingPath := func(path string) {
		workingPath = path
		title := "wui Designer"
		if path != "" {
			title += " - " + path
		}
		w.SetTitle(title)
	}
	setWorkingPath("")

	fileOpenMenu.SetOnClick(func() {
		open := wui.NewFileOpenDialog()
		open.SetTitle("Select a Go file containing one or more wui.Windows")
		open.AddFilter("Go file", ".go")
		if accept, path := open.ExecuteSingleSelection(w); accept {
			setWorkingPath(path)
			windows, err := openFile(path)
			if err != nil {
				wui.MessageBoxError("Error", err.Error())
			} else {
				if len(windows) != 1 {
					panic("TODO Handle more than one windows, present a select dialog")
				}
				theWindow = windows[0]
				activate(theWindow)
			}
		}
	})

	saveCodeTo := func(path string) {
		code := generatePreviewCode(theWindow)
		err := ioutil.WriteFile(path, code, 0666)
		if err != nil {
			wui.MessageBoxError("Error", err.Error())
		} else {
			workingPath = path
		}
	}

	fileSaveAsMenu.SetOnClick(func() {
		save := wui.NewFileSaveDialog()
		save.SetAppendExt(true)
		save.AddFilter("Go file", ".go")
		if accept, path := save.Execute(w); accept {
			saveCodeTo(path)
		}
	})

	fileSaveMenu.SetOnClick(func() {
		if workingPath != "" {
			saveCodeTo(workingPath)
		} else {
			fileSaveAsMenu.OnClick()()
		}
	})

	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl, Rune: 'R'}, func() {
		showPreview(theWindow)
	})
	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl, Rune: 'O'}, fileOpenMenu.OnClick())
	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl, Rune: 'S'}, fileSaveMenu.OnClick())
	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl | wui.ModShift, Rune: 'S'}, fileSaveAsMenu.OnClick())

	w.SetShortcut(wui.ShortcutKeys{Rune: 27}, w.Close) // TODO ESC for debugging

	w.Maximize()
	w.Show()
}

type rectangle struct {
	x, y, w, h int
}

func (r rectangle) contains(x, y int) bool {
	return x >= r.x && y >= r.y && x < r.x+r.w && y < r.y+r.h
}

func defaultWindow() *wui.Window {
	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -11})
	w := wui.NewWindow()
	w.SetFont(font)
	w.SetTitle("Window")
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
	DrawEllipse(x, y, w, h int, color wui.Color)
	FillEllipse(x, y, w, h int, color wui.Color)
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

func (d *offsetDrawer) DrawEllipse(x, y, w, h int, color wui.Color) {
	d.base.DrawEllipse(x+d.dx, y+d.dy, w, h, color)
}

func (d *offsetDrawer) FillEllipse(x, y, w, h int, color wui.Color) {
	d.base.FillEllipse(x+d.dx, y+d.dy, w, h, color)
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
		drawControl(child, d)
	}
}

func drawControl(c wui.Control, d drawer) {
	switch x := c.(type) {
	case *wui.Button:
		drawButton(x, d)
	case *wui.RadioButton:
		drawRadioButton(x, d)
	case *wui.CheckBox:
		drawCheckBox(x, d)
	case *wui.Panel:
		drawPanel(x, d)
	default:
		panic("unhandled control type")
	}
}

func drawButton(b *wui.Button, d drawer) {
	x, y, w, h := b.Bounds()
	d.DrawRect(x, y, w, h, wui.RGB(240, 240, 240))
	d.FillRect(x+1, y+1, w-2, h-2, wui.RGB(173, 173, 173))
	d.FillRect(x+2, y+2, w-4, h-4, wui.RGB(225, 225, 225))
	d.TextRectFormat(x, y, w, h, b.Text(), wui.FormatCenter, wui.RGB(0, 0, 0))
}

func drawRadioButton(r *wui.RadioButton, d drawer) {
	x, y, w, h := r.Bounds()
	d.FillRect(x, y, w, h, wui.RGB(240, 240, 240))
	d.FillEllipse(x, y+(h-13)/2, 13, 13, wui.RGB(255, 255, 255))
	d.DrawEllipse(x, y+(h-13)/2, 13, 13, wui.RGB(0, 0, 0))
	if r.Checked() {
		d.FillEllipse(x+3, y+(h-13)/2+3, 7, 7, wui.RGB(0, 0, 0))
	}
	d.TextRectFormat(x+16, y, w-16, h, r.Text(), wui.FormatCenterLeft, wui.RGB(0, 0, 0))
}

func drawCheckBox(c *wui.CheckBox, d drawer) {
	x, y, w, h := c.Bounds()
	d.FillRect(x, y, w, h, wui.RGB(240, 240, 240))
	d.FillRect(x, y+(h-13)/2, 13, 13, wui.RGB(255, 255, 255))
	d.DrawRect(x, y+(h-13)/2, 13, 13, wui.RGB(0, 0, 0))
	if c.Checked() {
		// Draw two lines for the check mark. ✓
		startX := x + 2
		startY := y + (h-13)/2 + 6
		d.Line(startX, startY, startX+3, startY+3, wui.RGB(0, 0, 0))
		d.Line(startX+3, startY+2, startX+9, startY-4, wui.RGB(0, 0, 0))
	}
	d.TextRectFormat(x+16, y, w-16, h, c.Text(), wui.FormatCenterLeft, wui.RGB(0, 0, 0))
}

func drawPanel(p *wui.Panel, d drawer) {
	x, y, w, h := p.Bounds()
	switch p.BorderStyle() {
	case wui.PanelBorderNone:
		d.DrawRect(x, y, w, h, wui.RGB(230, 230, 230))
	case wui.PanelBorderSingleLine:
		d.DrawRect(x, y, w, h, wui.RGB(100, 100, 100))
	case wui.PanelBorderRaised:
		d.Line(x, y, x+w, y, wui.RGB(227, 227, 227))
		d.Line(x, y, x, y+h, wui.RGB(227, 227, 227))
		d.Line(x+w-1, y, x+w-1, y+h, wui.RGB(105, 105, 105))
		d.Line(x, y+h-1, x+w, y+h-1, wui.RGB(105, 105, 105))
		d.Line(x+1, y+1, x+w-1, y+1, wui.RGB(255, 255, 255))
		d.Line(x+1, y+1, x+1, y+h-1, wui.RGB(255, 255, 255))
		d.Line(x+w-2, y+1, x+w-2, y+h-1, wui.RGB(160, 160, 160))
		d.Line(x+1, y+h-2, x+w-1, y+h-2, wui.RGB(160, 160, 160))
	case wui.PanelBorderSunken:
		d.Line(x, y, x+w, y, wui.RGB(160, 160, 160))
		d.Line(x, y, x, y+h, wui.RGB(160, 160, 160))
		d.Line(x+w-1, y, x+w-1, y+h, wui.RGB(255, 255, 255))
		d.Line(x, y+h-1, x+w, y+h-1, wui.RGB(255, 255, 255))
	case wui.PanelBorderSunkenThick:
		d.Line(x, y, x+w, y, wui.RGB(160, 160, 160))
		d.Line(x, y, x, y+h, wui.RGB(160, 160, 160))
		d.Line(x+w-1, y, x+w-1, y+h, wui.RGB(255, 255, 255))
		d.Line(x, y+h-1, x+w, y+h-1, wui.RGB(255, 255, 255))
		d.Line(x+1, y+1, x+w-1, y+1, wui.RGB(105, 105, 105))
		d.Line(x+1, y+1, x+1, y+h-1, wui.RGB(105, 105, 105))
		d.Line(x+w-2, y+1, x+w-2, y+h-1, wui.RGB(227, 227, 227))
		d.Line(x+1, y+h-2, x+w-1, y+h-2, wui.RGB(227, 227, 227))
	}
	// TODO Use inner coordinates for drawing panels once they are
	// supported in the library.
	drawContainer(p, makeOffsetDrawer(d, p.X(), p.Y()))
}

func anchorToString(a wui.Anchor) string {
	switch a {
	case wui.AnchorMin:
		return "AnchorMin"
	case wui.AnchorMax:
		return "AnchorMax"
	case wui.AnchorCenter:
		return "AnchorCenter"
	case wui.AnchorMinAndMax:
		return "AnchorMinAndMax"
	case wui.AnchorMinAndCenter:
		return "AnchorMinAndCenter"
	case wui.AnchorMaxAndCenter:
		return "AnchorMaxAndCenter"
	default:
		panic("unhandled anchor type")
	}
}

func panelBorderToString(s wui.PanelBorderStyle) string {
	switch s {
	case wui.PanelBorderNone:
		return "PanelBorderNone"
	case wui.PanelBorderSingleLine:
		return "PanelBorderSingleLine"
	case wui.PanelBorderSunken:
		return "PanelBorderSunken"
	case wui.PanelBorderSunkenThick:
		return "PanelBorderSunkenThick"
	case wui.PanelBorderRaised:
		return "PanelBorderRaised"
	default:
		panic("unhandled panel border style")
	}
}

type node interface {
	Parent() wui.Container
	Bounds() (x, y, width, height int)
}

func showPreview(w *wui.Window) {
	code := generatePreviewCode(w)
	const fileName = "wui_designer_temp_file.go"
	err := ioutil.WriteFile(fileName, code, 0666)
	if err != nil {
		wui.MessageBoxError("Error", err.Error())
		return
	}
	defer os.Remove(fileName)
	// TODO This blocks and freezes the designer, instead build into a temporary
	// directory and check that the build worked fine, then start the process
	// non-blocking.
	output, err := exec.Command("go", "run", fileName).CombinedOutput()
	if err != nil {
		wui.MessageBoxError("Error", err.Error()+"\r\n"+string(output))
	} else if len(output) > 0 {
		wui.MessageBoxInfo("go output", string(output))
	}
}

func generatePreviewCode(w *wui.Window) []byte {
	var code bytes.Buffer
	code.WriteString(`//+build ignore

package main

import "github.com/gonutz/wui"

func main() {`)

	line := func(format string, a ...interface{}) {
		fmt.Fprint(&code, "\n")
		fmt.Fprintf(&code, format, a...)
	}

	name := names[w]
	if name == "" {
		name = "w"
	}
	line(name + " := wui.NewWindow()")
	line(name+".SetTitle(%q)", w.Title())
	line(name+".SetSize(%d, %d)", w.Width(), w.Height())
	if w.Alpha() != 255 {
		line(name+".SetAlpha(%d)", w.Alpha())
	}
	font := w.Font()
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
		line(name + ".SetFont(font)")
	}

	writeContainer(w, name, line)

	line("")
	line(name + ".Show()")
	code.WriteString("\n}")

	formatted, err := format.Source(code.Bytes())
	if err != nil {
		panic("We generated wrong code: " + err.Error())
	}
	return formatted
}

func writeContainer(c wui.Container, parent string, line func(format string, a ...interface{})) {
	for i, child := range c.Children() {
		line("")
		name := names[child]
		if name == "" {
			name = fmt.Sprintf("%s_child%d", parent, i)
		}
		do := func(format string, a ...interface{}) {
			line(name+format, a...)
		}
		if button, ok := child.(*wui.Button); ok {
			do(" := wui.NewButton()")
			do(".SetBounds(%d, %d, %d, %d)", button.X(), button.Y(), button.Width(), button.Height())
			h, v := button.Anchors()
			if h != wui.Anchor(0) {
				do(".SetHorizontalAnchor(wui.%s)", anchorToString(h))
			}
			if v != wui.Anchor(0) {
				do(".SetVerticalAnchor(wui.%s)", anchorToString(v))
			}
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
			border := panel.BorderStyle()
			if border != wui.PanelBorderNone {
				do(".SetBorderStyle(wui.%s)", panelBorderToString(border))
			}
			do(".SetBounds(%d, %d, %d, %d)", panel.X(), panel.Y(), panel.Width(), panel.Height())
			h, v := panel.Anchors()
			if h != wui.Anchor(0) {
				do(".SetHorizontalAnchor(wui.%s)", anchorToString(h))
			}
			if v != wui.Anchor(0) {
				do(".SetVerticalAnchor(wui.%s)", anchorToString(v))
			}
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
		} else if r, ok := child.(*wui.RadioButton); ok {
			do(" := wui.NewRadioButton()")
			do(".SetText(%q)", r.Text())
			do(".SetBounds(%d, %d, %d, %d)", r.X(), r.Y(), r.Width(), r.Height())
			h, v := r.Anchors()
			if h != wui.Anchor(0) {
				do(".SetHorizontalAnchor(wui.%s)", anchorToString(h))
			}
			if v != wui.Anchor(0) {
				do(".SetVerticalAnchor(wui.%s)", anchorToString(v))
			}
			if !r.Enabled() {
				do(".SetEnabled(false)")
			}
			if !r.Visible() {
				do(".SetVisible(false)")
			}
			if r.Checked() {
				do(".SetChecked(true)")
			}
			line("%s.Add(%s)", parent, name)
		} else if r, ok := child.(*wui.CheckBox); ok {
			do(" := wui.NewCheckBox()")
			do(".SetText(%q)", r.Text())
			do(".SetBounds(%d, %d, %d, %d)", r.X(), r.Y(), r.Width(), r.Height())
			h, v := r.Anchors()
			if h != wui.Anchor(0) {
				do(".SetHorizontalAnchor(wui.%s)", anchorToString(h))
			}
			if v != wui.Anchor(0) {
				do(".SetVerticalAnchor(wui.%s)", anchorToString(v))
			}
			if !r.Enabled() {
				do(".SetEnabled(false)")
			}
			if !r.Visible() {
				do(".SetVisible(false)")
			}
			if r.Checked() {
				do(".SetChecked(true)")
			}
			line("%s.Add(%s)", parent, name)
		} else {
			panic("unhandled child control")
		}
	}
}

func cloneControl(c wui.Control) wui.Control {
	switch x := c.(type) {
	case *wui.Button:
		b := wui.NewButton()
		b.SetText(x.Text())
		b.SetBounds(0, 0, x.Width(), x.Height())
		return b
	case *wui.CheckBox:
		c := wui.NewCheckBox()
		c.SetText(x.Text())
		c.SetChecked(x.Checked())
		c.SetBounds(0, 0, x.Width(), x.Height())
		return c
	case *wui.RadioButton:
		r := wui.NewRadioButton()
		r.SetText(x.Text())
		r.SetBounds(0, 0, x.Width(), x.Height())
		return r
	default:
		panic("unhandled control type in cloneControl")
	}
}
