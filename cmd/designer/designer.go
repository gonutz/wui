package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

var (
	// names associates variable names with the controls.
	names = make(map[interface{}]string)

	// The build* variables are for previews which create temporary Go and exe
	// files that need to be deleted at the end of the program.
	buildDir    = "."
	buildPrefix = "wui_designer_preview_"
	buildCount  = 0

	events = make(map[event]string)
)

type event struct {
	control interface{}
	name    string
}

func main() {
	// Create a temporary directory to save our preview builds in.
	if dir, err := ioutil.TempDir("", "wui_designer_preview_builds"); err == nil {
		buildDir = dir
		defer os.Remove(dir)
	}
	// After closing the designer, delete all preview builds from this session.
	defer func() {
		if files, err := ioutil.ReadDir(buildDir); err == nil {
			for _, file := range files {
				if !file.IsDir() &&
					strings.HasSuffix(file.Name(), ".exe") &&
					strings.HasPrefix(file.Name(), buildPrefix) {
					// Ignore the error on the remove, one of the builds might
					// still be running and cannot be removed. This is OK, most
					// of the files will get deleted.
					os.Remove(filepath.Join(buildDir, file.Name()))
				}
			}
		}
	}()

	var (
		// The ResizeAreas are the size drag points of the window.
		xResizeArea, yResizeArea, xyResizeArea rectangle
		// innerX and Y is the top-left corner of where theWindow's inner
		// rectangle is drawn, relative to the application window. This means we
		// can use innerX and Y in the application window's mouse events to find
		// the relative mouse position inside theWindow.
		innerX, innerY int
		// active is the highlighted control whose properties are shown in the
		// tool bar.
		active node
		// TODO Move preview somewhere else.
		preview = wui.NewPaintBox()
	)

	const (
		idleMouse = iota
		addingControl
	)
	mouseMode := idleMouse

	theWindow := defaultWindow()
	names[theWindow] = "w"

	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -11})
	w := wui.NewWindow()
	w.SetFont(font)
	w.SetTitle("wui Designer")
	w.SetBackground(w32.GetSysColorBrush(w32.COLOR_BTNFACE))
	w.SetInnerSize(800, 600)

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

	// TODO Doing this after the menu does not work.
	//w.SetInnerSize(800, 600)

	type uiProp struct {
		panel     *wui.Panel
		setter    string
		update    func()
		rightType func(t reflect.Type) bool
	}

	const propMargin = 2
	boolProp := func(name, getterFunc string) uiProp {
		setterFunc := "Set" + getterFunc // By convention.
		c := wui.NewCheckBox()
		c.SetText(name)
		c.SetBounds(100, propMargin, 95, 17)
		p := wui.NewPanel()
		p.SetSize(195, c.Height()+2*propMargin)
		w.Add(p)
		p.Add(c)
		c.SetOnChange(func(on bool) {
			reflect.ValueOf(active).MethodByName(setterFunc).Call(
				[]reflect.Value{reflect.ValueOf(on)},
			)
			preview.Paint()
		})
		update := func() {
			on := reflect.ValueOf(active).MethodByName(getterFunc).Call(nil)[0].Bool()
			c.SetChecked(on)
		}
		rightType := func(t reflect.Type) bool {
			return t.Kind() == reflect.Bool
		}
		return uiProp{
			panel:     p,
			setter:    setterFunc,
			update:    update,
			rightType: rightType,
		}
	}

	intProp := func(name, getterFunc string, minmax ...int) uiProp {
		setterFunc := "Set" + getterFunc // By convention.
		n := wui.NewIntUpDown()
		if len(minmax) == 2 {
			n.SetMinMax(minmax[0], minmax[1])
		}
		n.SetBounds(100, propMargin, 90, 22)
		l := wui.NewLabel()
		l.SetText(name)
		l.SetAlignment(wui.AlignRight)
		// TODO This -1 might have to do with the below TODO about the IntUpDown
		// height.
		l.SetBounds(0, propMargin-1, 95, n.Height())
		p := wui.NewPanel()
		// TODO We add +2 to the height because for some reason setting the
		// height of an IntUpDown does not include the borders. Fix this in the
		// wui library.
		p.SetSize(195, n.Height()+2+2*propMargin)
		w.Add(p)
		p.Add(l)
		p.Add(n)
		n.SetOnValueChange(func(v int) {
			if active == nil {
				return
			}
			if m, ok := reflect.TypeOf(active).MethodByName(setterFunc); ok {
				reflect.ValueOf(active).MethodByName(setterFunc).Call(
					[]reflect.Value{reflect.ValueOf(v).Convert(m.Type.In(1))},
				)
				preview.Paint()
			}
		})
		update := func() {
			v := reflect.ValueOf(active).MethodByName(getterFunc).Call(nil)[0]
			i := v.Convert(reflect.TypeOf(0)).Int()
			n.SetValue(int(i))
		}
		rightType := func(t reflect.Type) bool {
			return t.Kind() == reflect.Int || t.Kind() == reflect.Uint8
		}
		return uiProp{
			panel:     p,
			setter:    setterFunc,
			update:    update,
			rightType: rightType,
		}
	}

	floatProp := func(name, getterFunc string, minmax ...float64) uiProp {
		setterFunc := "Set" + getterFunc // By convention.
		n := wui.NewFloatUpDown()
		if len(minmax) == 2 {
			n.SetMinMax(minmax[0], minmax[1])
		}
		n.SetPrecision(6)
		n.SetBounds(100, propMargin, 90, 22)
		l := wui.NewLabel()
		l.SetText(name)
		l.SetAlignment(wui.AlignRight)
		// TODO This -1 might have to do with the below TODO about the
		// FloatUpDown height.
		l.SetBounds(0, propMargin-1, 95, n.Height())
		p := wui.NewPanel()
		// TODO We add +2 to the height because for some reason setting the
		// height of an FloatUpDown does not include the borders. Fix this in
		// the wui library.
		p.SetSize(195, n.Height()+2+2*propMargin)
		w.Add(p)
		p.Add(l)
		p.Add(n)
		n.SetOnValueChange(func(v float64) {
			if active == nil {
				return
			}
			if m, ok := reflect.TypeOf(active).MethodByName(setterFunc); ok {
				reflect.ValueOf(active).MethodByName(setterFunc).Call(
					[]reflect.Value{reflect.ValueOf(v).Convert(m.Type.In(1))},
				)
				preview.Paint()
			}
		})
		update := func() {
			v := reflect.ValueOf(active).MethodByName(getterFunc).Call(nil)[0]
			i := v.Convert(reflect.TypeOf(0.0)).Float()
			n.SetValue(i)
		}
		rightType := func(t reflect.Type) bool {
			return t.Kind() == reflect.Float32 || t.Kind() == reflect.Float64
		}
		return uiProp{
			panel:     p,
			setter:    setterFunc,
			update:    update,
			rightType: rightType,
		}
	}

	stringProp := func(name, getterFunc string) uiProp {
		setterFunc := "Set" + getterFunc // By convention.
		t := wui.NewEditLine()
		t.SetBounds(100, propMargin, 90, 22)
		l := wui.NewLabel()
		l.SetText(name)
		l.SetAlignment(wui.AlignRight)
		l.SetBounds(0, propMargin-1, 95, t.Height())
		p := wui.NewPanel()
		p.SetSize(195, t.Height()+2*propMargin)
		w.Add(p)
		p.Add(l)
		p.Add(t)
		t.SetOnTextChange(func() {
			if active == nil {
				return
			}
			if _, ok := reflect.TypeOf(active).MethodByName(setterFunc); ok {
				reflect.ValueOf(active).MethodByName(setterFunc).Call(
					[]reflect.Value{reflect.ValueOf(t.Text())},
				)
				preview.Paint()
			}
		})
		update := func() {
			text := reflect.ValueOf(active).MethodByName(getterFunc).Call(nil)[0].String()
			t.SetText(text)
		}
		rightType := func(t reflect.Type) bool {
			return t.Kind() == reflect.String
		}
		return uiProp{
			panel:     p,
			setter:    setterFunc,
			update:    update,
			rightType: rightType,
		}
	}

	stringListProp := func(name, getterFunc string) uiProp {
		setterFunc := "Set" + getterFunc // By convention.
		l := wui.NewLabel()
		l.SetBounds(10, 5, 180, 13)
		l.SetText(name)
		l.SetAlignment(wui.AlignCenter)
		list := wui.NewTextEdit()
		list.SetBounds(10, 20, 180, 80)
		list.SetAnchors(wui.AnchorMinAndMax, wui.AnchorMinAndMax)
		p := wui.NewPanel()
		p.SetSize(195, list.Height()+2*propMargin)
		w.Add(p)
		p.Add(l)
		p.Add(list)
		list.SetOnTextChange(func() {
			if active == nil {
				return
			}
			if _, ok := reflect.TypeOf(active).MethodByName(setterFunc); ok {
				items := strings.Split(list.Text(), "\r\n")
				items = removeEmptyStrings(items)
				l.SetText(fmt.Sprintf("%s (%d)", name, len(items)))
				reflect.ValueOf(active).MethodByName(setterFunc).Call(
					[]reflect.Value{reflect.ValueOf(items)},
				)
				preview.Paint()
			}
		})
		update := func() {
			items := reflect.ValueOf(active).MethodByName(getterFunc).Call(nil)[0].Interface().([]string)
			l.SetText(fmt.Sprintf("%s (%d)", name, len(items)))
			list.SetText(strings.Join(items, "\r\n") + "\r\n")
		}
		rightType := func(t reflect.Type) bool {
			// NOTE that currently there is only []string, we might have to
			// check for the underlying slice type if we support others in the
			// future.
			return t.Kind() == reflect.Slice
		}
		return uiProp{
			panel:     p,
			setter:    setterFunc,
			update:    update,
			rightType: rightType,
		}
	}

	// enumNames must correspond to the respective const, the order is important
	// and the consts must be iota'd, i.e. start with 0 and increment by 1.
	enumProp := func(name, getterFunc string, enumNames ...string) uiProp {
		setterFunc := "Set" + getterFunc // By convention.
		c := wui.NewComboBox()
		for _, name := range enumNames {
			c.AddItem(name)
		}
		c.SetBounds(100, propMargin, 90, 22)
		l := wui.NewLabel()
		l.SetText(name)
		l.SetAlignment(wui.AlignRight)
		l.SetBounds(0, propMargin-1, 95, c.Height())
		p := wui.NewPanel()
		p.SetSize(195, c.Height()+2*propMargin)
		w.Add(p)
		p.Add(l)
		p.Add(c)
		c.SetOnChange(func(index int) {
			m, ok := reflect.TypeOf(active).MethodByName(setterFunc)
			if ok {
				reflect.ValueOf(active).MethodByName(setterFunc).Call(
					[]reflect.Value{reflect.ValueOf(index).Convert(m.Type.In(1))},
				)
				preview.Paint()
			}
		})
		update := func() {
			v := reflect.ValueOf(active).MethodByName(getterFunc).Call(nil)[0]
			index := v.Convert(reflect.TypeOf(0)).Int()
			c.SetSelectedIndex(int(index))
		}
		rightType := func(t reflect.Type) bool {
			return true
		}
		return uiProp{
			panel:     p,
			setter:    setterFunc,
			update:    update,
			rightType: rightType,
		}
	}

	uiProps := []uiProp{
		stringProp("Title", "Title"),
		enumProp("Window State", "State",
			"Normal", "Maximized", "Minimized",
		),
		intProp("Alpha", "Alpha", 0, 255),
		boolProp("Enabled", "Enabled"),
		boolProp("Visible", "Visible"),
		enumProp("Horizontal Anchor", "HorizontalAnchor",
			"Left", "Right", "Center", "Left+Right", "Left+Center", "Right+Center",
		),
		enumProp("Vertical Anchor", "VerticalAnchor",
			"Top", "Bottom", "Center", "Top+Bottom", "Top+Center", "Bottom+Center",
		),
		intProp("X", "X"),
		intProp("Y", "Y"),
		intProp("Width", "Width"),
		intProp("Height", "Height"),
		stringProp("Text", "Text"),
		enumProp("Alignment", "Alignment",
			"Left", "Center", "Right",
		),
		boolProp("Checked", "Checked"),
		intProp("Arrow Increment", "ArrowIncrement"),
		intProp("Mouse Increment", "MouseIncrement"),
		intProp("Min", "Min"),
		intProp("Max", "Max"),
		intProp("Value", "Value"),
		floatProp("Min", "Min"),
		floatProp("Max", "Max"),
		floatProp("Value", "Value"),
		intProp("Precision", "Precision", 1, 6),
		enumProp("Orientation", "Orientation",
			"Horizontal", "Vertical",
		),
		intProp("Tick Frequency", "TickFrequency"),
		enumProp("Tick Position", "TickPosition",
			"Right/Bottom", "Left/Top", "Both Sides",
		),
		boolProp("Ticks Visible", "TicksVisible"),
		enumProp("Border Style", "BorderStyle",
			"None", "Single Line", "Sunken", "Sunken Thick", "Raised",
		),
		intProp("Character Limit", "CharacterLimit", 1, 0x7FFFFFFE),
		boolProp("Is Password", "IsPassword"),
		boolProp("Read Only", "ReadOnly"),
		stringListProp("Items", "Items"),
		intProp("Selected Index", "SelectedIndex", -1, math.MaxInt32),
		boolProp("Vertical", "Vertical"),
		boolProp("Moves Forever", "MovesForever"),
	}

	appIcon := w32.LoadIcon(0, w32.MakeIntResource(w32.IDI_APPLICATION))
	appIconWidth := w32.GetSystemMetrics(w32.SM_CXICON)
	appIconHeight := w32.GetSystemMetrics(w32.SM_CYICON)
	appIconWidth, appIconHeight = 17, 17

	defaultCursor := w.Cursor()

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
	checkBoxTemplate.SetText("Check Box")
	checkBoxTemplate.SetChecked(true)
	checkBoxTemplate.SetBounds(10, 44, 100, 17)

	radioButtonTemplate := wui.NewRadioButton()
	radioButtonTemplate.SetText("Radio Button")
	radioButtonTemplate.SetChecked(true)
	radioButtonTemplate.SetBounds(10, 67, 100, 17)

	sliderTemplate := wui.NewSlider()
	sliderTemplate.SetBounds(10, 95, 150, 45)

	panelTemplate := wui.NewPanel()
	panelTemplate.SetBounds(10, 150, 150, 50)
	panelTemplate.SetBorderStyle(wui.PanelBorderSingleLine)
	panelText := wui.NewLabel()
	panelText.SetText("Panel")
	panelText.SetAlignment(wui.AlignCenter)
	panelText.SetSize(panelTemplate.InnerWidth(), panelTemplate.InnerHeight())
	panelTemplate.Add(panelText)

	labelTemplate := wui.NewLabel()
	labelTemplate.SetText("Text Label")
	labelTemplate.SetBounds(10, 210, 150, 13)

	paintBoxTemplate := wui.NewPaintBox()
	paintBoxTemplate.SetBounds(10, 230, 150, 50)

	editLineTemplate := wui.NewEditLine()
	editLineTemplate.SetBounds(10, 290, 150, 20)
	editLineTemplate.SetText("Text Edit Line")

	intTemplate := wui.NewIntUpDown()
	intTemplate.SetBounds(10, 320, 80, 22)

	comboTemplate := wui.NewComboBox()
	comboTemplate.SetBounds(10, 350, 150, 21)
	comboTemplate.AddItem("Combo Box")
	comboTemplate.SetSelectedIndex(0)

	progressTemplate := wui.NewProgressBar()
	progressTemplate.SetBounds(10, 380, 150, 25)
	progressTemplate.SetValue(0.5)

	floatTemplate := wui.NewFloatUpDown()
	floatTemplate.SetBounds(10, 410, 80, 22)

	allTemplates := []wui.Control{
		buttonTemplate,
		checkBoxTemplate,
		radioButtonTemplate,
		sliderTemplate,
		panelTemplate,
		labelTemplate,
		paintBoxTemplate,
		editLineTemplate,
		intTemplate,
		comboTemplate,
		progressTemplate,
		floatTemplate,
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
		for _, template := range allTemplates {
			drawControl(template, c)
		}
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
	nameText.SetAlignment(wui.AlignRight)
	nameText.SetBounds(10, 10, 85, 20)
	w.Add(nameText)
	name := wui.NewEditLine()
	name.SetBounds(100, 10, 90, 22)
	w.Add(name)

	preview.SetBounds(200, 0, 400, 600)
	preview.SetHorizontalAnchor(wui.AnchorMinAndMax)
	preview.SetVerticalAnchor(wui.AnchorMinAndMax)
	white := wui.RGB(255, 255, 255)
	black := wui.RGB(0, 0, 0)

	editOnPaint := wui.NewButton()
	editOnPaint.SetText("OnPaint")
	editOnPaint.SetBounds(105, 500, 85, 25)
	editOnPaint.SetVisible(false) // TODO Bring this back.
	w.Add(editOnPaint)

	name.SetOnTextChange(func() {
		names[active] = name.Text()
	})
	editOnPaint.SetOnClick(func() {
		p, valid := active.(*wui.PaintBox)
		if !valid {
			panic("OnPaint only valid for paint boxes")
		}

		dlg := wui.NewWindow()
		dlg.SetPosition(w32.ClientToScreen(w32.HWND(preview.Handle()), 0, 0))
		dlg.SetSize(preview.Size())

		code := wui.NewTextEdit()
		font, _ := wui.NewFont(wui.FontDesc{Name: "Courier New", Height: -15})
		code.SetFont(font)
		// TODO code.SetWriteTabs(true)
		// TODO code.SetLineBreaks("\n")
		code.SetBounds(0, 0, dlg.InnerWidth(), dlg.InnerHeight()-30)
		code.SetAnchors(wui.AnchorMinAndMax, wui.AnchorMinAndMax)
		dlg.Add(code)

		onPaint := event{p, "OnPaint"}
		if events[onPaint] == "" {
			events[onPaint] = "func(canvas *wui.Canvas) {\n\t\n}"
		}
		code.SetText(strings.Replace(events[onPaint], "\n", "\r\n", -1))
		code.SetCursorPosition(29)

		ok := wui.NewButton()
		ok.SetText("OK")
		ok.SetBounds(dlg.InnerWidth()/2-87, dlg.InnerHeight()-28, 85, 25)
		ok.SetAnchors(wui.AnchorCenter, wui.AnchorMax)
		ok.SetOnClick(func() {
			events[onPaint] = strings.Replace(code.Text(), "\r", "", -1)
			dlg.Close()
		})
		dlg.Add(ok)

		cancel := wui.NewButton()
		cancel.SetText("Cancel")
		cancel.SetBounds(dlg.InnerWidth()/2+2, dlg.InnerHeight()-28, 85, 25)
		cancel.SetAnchors(wui.AnchorCenter, wui.AnchorMax)
		cancel.SetOnClick(dlg.Close)
		dlg.Add(cancel)

		dlg.SetOnShow(code.Focus)

		dlg.ShowModal()
	})

	activate := func(newActive node) {
		active = newActive

		name.SetText(names[active])
		y := name.Y() + name.Height() + propMargin
		for _, prop := range uiProps {
			m, hasProp := reflect.TypeOf(active).MethodByName(prop.setter)
			show := hasProp && prop.rightType(m.Type.In(1))
			prop.panel.SetVisible(show)
			if show {
				prop.panel.SetY(y)
				y += prop.panel.Height()
				prop.update()
			}
		}
	}
	activate(theWindow)

	const xOffset, yOffset = 5, 5
	preview.SetOnPaint(func(c *wui.Canvas) {
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
		if active != nil && active != theWindow {
			x, y, w, h := active.Bounds()
			parent := active.Parent()
			for parent != theWindow {
				dx, dy, _, _ := parent.InnerBounds()
				x += dx
				y += dy
				parent = parent.Parent()
			}
			if w < 0 {
				w = 0
			}
			if h < 0 {
				h = 0
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

	lastX, lastY := -999, -999
	w.SetOnMouseMove(func(x, y int) {
		if x == lastX && y == lastY {
			return
		}
		lastX, lastY = x, y

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
				w.SetCursor(wui.CursorSizeWE)
			} else if yResizeArea.contains(x, y) {
				w.SetCursor(wui.CursorSizeNS)
			} else if xyResizeArea.contains(x, y) {
				w.SetCursor(wui.CursorSizeNWSE)
			} else {
				w.SetCursor(defaultCursor)
			}
		} else {
			if dragMode == dragX || dragMode == dragXY {
				dx := x - dragStartX
				theWindow.SetWidth(dragStartWidth + dx)
			}
			if dragMode == dragY || dragMode == dragXY {
				dy := y - dragStartY
				theWindow.SetHeight(dragStartHeight + dy)
			}
			activate(theWindow) // Update the size in the UI.
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
				innerX, innerY, _, _ := theWindow.InnerBounds()
				outerX, outerY, _, _ := theWindow.Bounds()
				x, y, w, h := controlToAdd.Bounds()
				relX := x - (xOffset + innerX - outerX)
				relY := y - (yOffset + innerY - outerY)
				// Find the sub-container that this is to be placed in. Use the
				// center of the new control to determine where to add it.
				addToThis, x, y := findContainerAt(theWindow, relX+w/2, relY+h/2)
				controlToAdd.SetBounds(x-w/2, y-h/2, w, h)
				addToThis.Add(controlToAdd)
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
			wui.MessageBoxError("TODO", "Open is not yet implemented")
		}
	})

	saveCodeTo := func(path string) {
		code := generatePreviewCode(theWindow, theWindow.X(), theWindow.Y())
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
		// We place the window such that it lies exactly over our drawing.
		x, y := w32.ClientToScreen(w32.HWND(w.Handle()), preview.X(), preview.Y())
		showPreview(w, theWindow, x+xOffset, y+yOffset)
	})
	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl, Rune: 'O'}, fileOpenMenu.OnClick())
	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl, Rune: 'S'}, fileSaveMenu.OnClick())
	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl | wui.ModShift, Rune: 'S'}, fileSaveAsMenu.OnClick())

	w.SetShortcut(wui.ShortcutKeys{Rune: 27}, w.Close) // TODO ESC for debugging

	w.SetState(wui.WindowMaximized)
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

func innerContains(b innerBounder, atX, atY int) bool {
	x, y, w, h := b.InnerBounds()
	return atX >= x && atY >= y && atX < x+w && atY < y+h
}

type innerBounder interface {
	InnerBounds() (x, y, width, height int)
}

type drawer interface {
	PushDrawRegion(x, y, width, height int)
	PopDrawRegion()
	Line(x1, y1, x2, y2 int, color wui.Color)
	DrawRect(x, y, w, h int, color wui.Color)
	FillRect(x, y, w, h int, color wui.Color)
	DrawEllipse(x, y, w, h int, color wui.Color)
	FillEllipse(x, y, w, h int, color wui.Color)
	TextRectFormat(x, y, w, h int, s string, format wui.Format, color wui.Color)
	TextExtent(s string) (width, height int)
	TextOut(x, y int, s string, color wui.Color)
	Polygon(p []w32.POINT, color wui.Color)
}

func makeOffsetDrawer(base drawer, dx, dy int) drawer {
	return &offsetDrawer{base: base, dx: dx, dy: dy}
}

type offsetDrawer struct {
	base   drawer
	dx, dy int
}

func (d *offsetDrawer) PushDrawRegion(x, y, width, height int) {
	d.base.PushDrawRegion(x+d.dx, y+d.dy, width, height)
}

func (d *offsetDrawer) PopDrawRegion() {
	d.base.PopDrawRegion()
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

func (d *offsetDrawer) TextExtent(s string) (width, height int) {
	return d.base.TextExtent(s)
}

func (d *offsetDrawer) TextOut(x, y int, s string, color wui.Color) {
	d.base.TextOut(x+d.dx, y+d.dy, s, color)
}

func (d *offsetDrawer) Polygon(p []w32.POINT, color wui.Color) {
	for i := range p {
		p[i].X += int32(d.dx)
		p[i].Y += int32(d.dy)
	}
	d.base.Polygon(p, color)
}

func (d *offsetDrawer) Line(x1, y1, x2, y2 int, color wui.Color) {
	d.base.Line(x1+d.dx, y1+d.dy, x2+d.dx, y2+d.dy, color)
}

func drawContainer(container wui.Container, d drawer) {
	_, _, w, h := container.InnerBounds()
	d.PushDrawRegion(0, 0, w, h)
	for _, child := range container.Children() {
		drawControl(child, d)
	}
	d.PopDrawRegion()
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
	case *wui.Slider:
		drawSlider(x, d)
	case *wui.Label:
		drawLabel(x, d)
	case *wui.PaintBox:
		drawPaintBox(x, d)
	case *wui.EditLine:
		drawEditLine(x, d)
	case *wui.IntUpDown:
		drawIntUpDown(x, d)
	case *wui.ComboBox:
		drawComboBox(x, d)
	case *wui.ProgressBar:
		drawProgressBar(x, d)
	case *wui.FloatUpDown:
		drawFloatUpDown(x, d)
	default:
		panic("unhandled control type")
	}
}

func drawButton(b *wui.Button, d drawer) {
	x, y, w, h := b.Bounds()
	if w > 0 && h > 0 {
		d.DrawRect(x, y, w, h, wui.RGB(240, 240, 240))
	}
	if w > 2 && h > 2 {
		d.FillRect(x+1, y+1, w-2, h-2, wui.RGB(173, 173, 173))
	}
	if w > 4 && h > 4 {
		d.FillRect(x+2, y+2, w-4, h-4, wui.RGB(225, 225, 225))
	}
	if w > 6 && h > 6 {
		textW, textH := d.TextExtent(b.Text())
		d.PushDrawRegion(x+3, y+3, w-6, h-6)
		d.TextOut(x+(w-textW)/2, y+(h-textH)/2, b.Text(), wui.RGB(0, 0, 0))
		d.PopDrawRegion()
	}
}

func drawRadioButton(r *wui.RadioButton, d drawer) {
	x, y, w, h := r.Bounds()
	d.PushDrawRegion(x, y, w, h)
	d.FillRect(x, y, w, h, wui.RGB(240, 240, 240))
	d.FillEllipse(x, y+(h-13)/2, 13, 13, wui.RGB(255, 255, 255))
	d.DrawEllipse(x, y+(h-13)/2, 13, 13, wui.RGB(0, 0, 0))
	if r.Checked() {
		d.FillEllipse(x+3, y+(h-13)/2+3, 7, 7, wui.RGB(0, 0, 0))
	}
	_, textH := d.TextExtent(r.Text())
	d.TextOut(x+16, y+(h-textH)/2, r.Text(), wui.RGB(0, 0, 0))
	d.PopDrawRegion()
}

func drawCheckBox(c *wui.CheckBox, d drawer) {
	x, y, w, h := c.Bounds()
	d.PushDrawRegion(x, y, w, h)
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
	_, textH := d.TextExtent(c.Text())
	d.TextOut(x+16, y+(h-textH)/2, c.Text(), wui.RGB(0, 0, 0))
	d.PopDrawRegion()
}

func drawPanel(p *wui.Panel, d drawer) {
	x, y, w, h := p.Bounds()
	if w <= 0 || h <= 0 {
		return
	}
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
	innerX, innerY, _, _ := p.InnerBounds()
	drawContainer(p, makeOffsetDrawer(d, innerX, innerY))
}

func drawSlider(s *wui.Slider, d drawer) {
	var (
		drawSlideBar    func(offset int)
		drawCursorBody  func(offset, size int)
		drawCursorArrow func(offset int)
		// drawEndTicks and drawMiddleTicks are only assigned if ticks are
		// visible for this slider.
		drawEndTicks    = func(offset int) {}
		drawMiddleTicks = func(offset int) {}
	)

	cursorColor := wui.RGB(0, 120, 215)
	tickColor := wui.RGB(196, 196, 196)
	slideBarBorder := wui.RGB(214, 214, 214)
	slideBarBackground := wui.RGB(231, 231, 234)

	x, y, w, h := s.Bounds()
	if w <= 0 || h <= 0 {
		return
	}
	d.PushDrawRegion(x, y, w, h)
	defer d.PopDrawRegion()
	min, max := s.MinMax()
	innerTickCount := max - min - 1
	freq := s.TickFrequency()
	relCursor := s.CursorPosition() - min

	if s.Orientation() == wui.HorizontalSlider {
		xLeft := x + 13
		xRight := x + w - 14
		scale := 1.0 / float64(innerTickCount+1) * float64(xRight-xLeft)
		if xRight < xLeft {
			xRight = xLeft
			scale = 0
		}
		xOffset := int(float64(relCursor)*scale + 0.5)
		cursorCenter := xLeft + xOffset

		drawSlideBar = func(offset int) {
			if xLeft != xRight {
				d.DrawRect(x+8, y+offset, w-16, 4, slideBarBorder)
				d.FillRect(x+9, y+offset+1, w-18, 2, slideBarBackground)
			}
		}
		drawCursorBody = func(offset, size int) {
			d.FillRect(cursorCenter-5, y+offset, 11, size, cursorColor)
		}
		drawCursorArrow = func(offset int) {
			d.Polygon([]w32.POINT{
				{int32(cursorCenter - 5), int32(y + 15)},
				{int32(cursorCenter), int32(y + 15 + offset)},
				{int32(cursorCenter + 5), int32(y + 15)},
			}, cursorColor)
		}

		if s.TicksVisible() {
			drawEndTicks = func(offset int) {
				d.Line(xLeft, y+offset, xLeft, y+offset+4, tickColor)
				d.Line(xRight, y+offset, xRight, y+offset+4, tickColor)
			}
			drawMiddleTicks = func(offset int) {
				for i := freq; i <= innerTickCount; i += freq {
					x := xLeft + int(float64(i)*scale+0.5)
					d.Line(x, y+offset, x, y+offset+3, tickColor)
				}
			}
		}
	} else {
		yTop := y + 13
		yBottom := y + h - 14
		scale := 1.0 / float64(innerTickCount+1) * float64(yBottom-yTop)
		if yBottom < yTop {
			yBottom = yTop
			scale = 0
		}
		yOffset := int(float64(relCursor)*scale + 0.5)
		cursorCenter := yTop + yOffset

		drawSlideBar = func(offset int) {
			if yTop != yBottom {
				d.DrawRect(x+offset, y+8, 4, h-16, slideBarBorder)
				d.FillRect(x+offset+1, y+9, 2, h-18, slideBarBackground)
			}
		}
		drawCursorBody = func(offset, size int) {
			d.FillRect(x+offset, cursorCenter-5, size, 11, cursorColor)
		}
		drawCursorArrow = func(offset int) {
			d.Polygon([]w32.POINT{
				{int32(x + 15), int32(cursorCenter - 5)},
				{int32(x + 15 + offset), int32(cursorCenter)},
				{int32(x + 15), int32(cursorCenter + 5)},
			}, cursorColor)
		}

		if s.TicksVisible() {
			drawEndTicks = func(offset int) {
				d.Line(x+offset, yTop, x+offset+4, yTop, tickColor)
				d.Line(x+offset, yBottom, x+offset+4, yBottom, tickColor)
			}
			drawMiddleTicks = func(offset int) {
				for i := freq; i <= innerTickCount; i += freq {
					y := yTop + int(float64(i)*scale+0.5)
					d.Line(x+offset, y, x+offset+3, y, tickColor)
				}
			}
		}
	}

	switch s.TickPosition() {
	case wui.TicksBottomOrRight:
		drawSlideBar(8)
		drawCursorBody(2, 14)
		drawCursorArrow(5)
		drawEndTicks(22)
		drawMiddleTicks(22)
	case wui.TicksTopOrLeft:
		drawSlideBar(18)
		drawCursorBody(15, 14)
		drawCursorArrow(-5)
		drawEndTicks(5)
		drawMiddleTicks(6)
	case wui.TicksOnBothSides:
		drawSlideBar(19)
		drawCursorBody(10, 21)
		drawEndTicks(5)
		drawEndTicks(33)
		drawMiddleTicks(6)
		drawMiddleTicks(33)
	default:
		panic("unhandled tick position")
	}
}

func drawLabel(l *wui.Label, d drawer) {
	x, y, w, h := l.Bounds()
	textW, textH := d.TextExtent(l.Text())
	textX := x
	switch l.Alignment() {
	case wui.AlignCenter:
		textX = x + (w-textW)/2
	case wui.AlignRight:
		textX = x + w - textW
	}
	d.PushDrawRegion(x, y, w, h)
	d.TextOut(textX, y+(h-textH)/2, l.Text(), wui.RGB(0, 0, 0))
	d.PopDrawRegion()
}

func drawPaintBox(p *wui.PaintBox, d drawer) {
	x, y, w, h := p.Bounds()
	if w > 0 && h > 0 {
		d.DrawRect(x, y, w, h, wui.RGB(0, 0, 0))
		d.TextRectFormat(x, y, w, h, "Paint Box", wui.FormatCenter, wui.RGB(0, 0, 0))
	}
}

func drawIntUpDown(e *wui.IntUpDown, d drawer) {
	x, y, w, h := e.Bounds()
	if w > 0 && h > 0 {
		d.PushDrawRegion(x, y, w, h)
		d.DrawRect(x, y, w, h, wui.RGB(122, 122, 122))
		d.FillRect(x+1, y+1, w-2, h-2, wui.RGB(255, 255, 255))

		text := strconv.Itoa(e.Value())
		color := wui.RGB(0, 0, 0)
		d.TextOut(x+6, y+3, text, color)

		d.FillRect(x+w-19, y, 19, h, wui.RGB(231, 231, 231))
		d.DrawRect(x+w-19, y, 19, h, wui.RGB(172, 172, 172))
		d.DrawRect(x+w-19+2, y+2, 19-4, h-4, wui.RGB(172, 172, 172))
		d.DrawRect(x+w-19+2, y+h/2-1, 19-4, 2, wui.RGB(172, 172, 172))
		y1 := y + h/4
		d.Line(x+w-12, y1+2, x+w-12+5, y1+2, wui.RGB(0, 0, 0))
		d.Line(x+w-11, y1+1, x+w-11+3, y1+1, wui.RGB(0, 0, 0))
		d.Line(x+w-10, y1+0, x+w-10+1, y1+0, wui.RGB(0, 0, 0))
		y2 := y + 3*h/4 - 2
		d.Line(x+w-12, y2+0, x+w-12+5, y2+0, wui.RGB(0, 0, 0))
		d.Line(x+w-11, y2+1, x+w-11+3, y2+1, wui.RGB(0, 0, 0))
		d.Line(x+w-10, y2+2, x+w-10+1, y2+2, wui.RGB(0, 0, 0))
		d.PopDrawRegion()
	}
}

func drawFloatUpDown(e *wui.FloatUpDown, d drawer) {
	x, y, w, h := e.Bounds()
	if w > 0 && h > 0 {
		d.PushDrawRegion(x, y, w, h)
		d.DrawRect(x, y, w, h, wui.RGB(122, 122, 122))
		d.FillRect(x+1, y+1, w-2, h-2, wui.RGB(255, 255, 255))

		text := fmt.Sprintf("%."+strconv.Itoa(e.Precision())+"f", e.Value())
		color := wui.RGB(0, 0, 0)
		d.TextOut(x+6, y+3, text, color)

		d.FillRect(x+w-19, y, 19, h, wui.RGB(231, 231, 231))
		d.DrawRect(x+w-19, y, 19, h, wui.RGB(172, 172, 172))
		d.DrawRect(x+w-19+2, y+2, 19-4, h-4, wui.RGB(172, 172, 172))
		d.DrawRect(x+w-19+2, y+h/2-1, 19-4, 2, wui.RGB(172, 172, 172))
		y1 := y + h/4
		d.Line(x+w-12, y1+2, x+w-12+5, y1+2, wui.RGB(0, 0, 0))
		d.Line(x+w-11, y1+1, x+w-11+3, y1+1, wui.RGB(0, 0, 0))
		d.Line(x+w-10, y1+0, x+w-10+1, y1+0, wui.RGB(0, 0, 0))
		y2 := y + 3*h/4 - 2
		d.Line(x+w-12, y2+0, x+w-12+5, y2+0, wui.RGB(0, 0, 0))
		d.Line(x+w-11, y2+1, x+w-11+3, y2+1, wui.RGB(0, 0, 0))
		d.Line(x+w-10, y2+2, x+w-10+1, y2+2, wui.RGB(0, 0, 0))
		d.PopDrawRegion()
	}
}

func drawComboBox(c *wui.ComboBox, d drawer) {
	x, y, w, h := c.Bounds()
	if w > 0 && h > 0 {
		d.PushDrawRegion(x, y, w, h)
		d.DrawRect(x, y, w, h, wui.RGB(173, 173, 173))
		d.FillRect(x+1, y+1, w-2, h-2, wui.RGB(225, 225, 225))
		arrowX := x + w - 13
		arrowY := y + 9
		d.Line(arrowX, arrowY, arrowX+4, arrowY+4, wui.RGB(86, 86, 86))
		d.Line(arrowX+4, arrowY+3, arrowX+8, arrowY-1, wui.RGB(86, 86, 86))
		if w > 20 {
			i := c.SelectedIndex()
			items := c.Items()
			if 0 <= i && i < len(items) {
				text := items[i]
				d.PushDrawRegion(x, y, w-20, h)
				d.TextOut(x+4, y+4, text, wui.RGB(0, 0, 0))
				d.PopDrawRegion()
			}
		}
		d.PopDrawRegion()
	}
}

func drawProgressBar(p *wui.ProgressBar, d drawer) {
	x, y, w, h := p.Bounds()
	if w > 0 && h > 0 {
		d.PushDrawRegion(x, y, w, h)
		d.DrawRect(x, y, w, h, wui.RGB(188, 188, 188))
		d.FillRect(x+1, y+1, w-2, h-2, wui.RGB(230, 230, 230))
		filledW := int(float64(w-2)*p.Value() + 0.5)
		d.FillRect(x+1, y+1, filledW, h-2, wui.RGB(0, 180, 40))
		d.PopDrawRegion()
	}
}

func drawEditLine(e *wui.EditLine, d drawer) {
	x, y, w, h := e.Bounds()
	if w > 0 && h > 0 {
		d.PushDrawRegion(x, y, w, h)
		if e.Enabled() {
			d.DrawRect(x, y, w, h, wui.RGB(122, 122, 122))
		} else {
			d.DrawRect(x, y, w, h, wui.RGB(204, 204, 204))
		}
		d.FillRect(x+1, y+1, w-2, h-2, wui.RGB(255, 255, 255))
		if e.ReadOnly() || !e.Enabled() {
			d.FillRect(x+2, y+2, w-4, h-4, wui.RGB(240, 240, 240))
		}
		text := e.Text()
		if e.IsPassword() {
			text = strings.Repeat("●", utf8.RuneCountInString(text))
		}
		color := wui.RGB(0, 0, 0)
		if !e.Enabled() {
			color = wui.RGB(109, 109, 109)
		}
		d.TextOut(x+6, y+3, text, color)
		d.PopDrawRegion()
	}
}

type node interface {
	Parent() wui.Container
	Bounds() (x, y, width, height int)
	SetBounds(x, y, width, height int)
}

func showPreview(parent, w *wui.Window, x, y int) {
	// Create a centered progress dialog that cannot be closed until the preview
	// is shown.
	canClose := make(chan bool, 1)
	progress := wui.NewDialogWindow()
	progress.SetTitle("Generating Preview...")
	progress.DisableAltF4()
	progress.SetOnCanClose(func() bool {
		return <-canClose
	})
	progress.SetInnerSize(420, 50)
	progress.SetX(parent.X() + (parent.Width()-progress.Width())/2)
	progress.SetY(parent.Y() + (parent.Height()-progress.Height())/2)
	p := wui.NewProgressBar()
	p.SetMovesForever(true)
	p.SetBounds(10, 10, 400, 30)
	progress.Add(p)

	// Generate the code in a different go routine while the progress bar is
	// showing.
	go func() {
		defer func() {
			canClose <- true
			progress.Close()
		}()

		code := generatePreviewCode(w, x, y)

		// Write the Go file to our temporary build dir.
		goFile := filepath.Join(buildDir, "wui_designer_temp_file.go")
		err := ioutil.WriteFile(goFile, code, 0666)
		if err != nil {
			wui.MessageBoxError("Error", err.Error())
			return
		}
		defer os.Remove(goFile)

		// Build the executable into our temporary build dir.
		exeFile := filepath.Join(buildDir, buildPrefix+strconv.Itoa(buildCount)+".exe")
		buildCount++

		// Do the build synchronously and report any build errors.
		output, err := exec.Command("go", "build", "-o", exeFile, goFile).CombinedOutput()
		if err != nil {
			wui.MessageBoxError("Error", err.Error()+"\r\n"+string(output))
			return
		}

		// Start the program in parallel so we can have multiple previews open at
		// once.
		exec.Command(exeFile).Start()
	}()

	progress.ShowModal()
}

func generatePreviewCode(w *wui.Window, x, y int) []byte {
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
	// TODO Use reflection for the window as well, like the other controls.
	line(name + " := wui.NewWindow()")
	line(name+".SetTitle(%q)", w.Title())
	line(name+".SetPosition(%d, %d)", x, y)
	line(name+".SetSize(%d, %d)", w.Width(), w.Height())
	if w.Alpha() != 255 {
		line(name+".SetAlpha(%d)", w.Alpha())
	}
	if w.State() != wui.WindowNormal {
		line(name+".SetState(%s)", w.State().String())
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

		typeName := reflect.TypeOf(child).Elem().Name()
		do(" := wui.New%s()", typeName)

		setters := generateProperties(name, child)
		for _, setter := range setters {
			line("\t" + setter)
		}

		// TODO Generate ALL events.
		if p, ok := child.(*wui.PaintBox); ok {
			onPaint := event{p, "OnPaint"}
			if events[onPaint] != "" {
				do(".SetOnPaint(%s)", events[onPaint])
			}
		}

		line("%s.Add(%s)", parent, name)

		if p, ok := child.(*wui.Panel); ok {
			// TODO We would want to fill in the panel content above, before
			// adding the panel to the parent, but there is a bug in Panel.Add,
			// see the comment there.
			writeContainer(p, name, line)
		}
	}
}

func cloneControl(c wui.Control) wui.Control {
	// TODO Use the properties for this.
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
	case *wui.Slider:
		s := wui.NewSlider()
		s.SetMinMax(x.MinMax())
		s.SetCursorPosition(x.CursorPosition())
		s.SetTickFrequency(x.TickFrequency())
		s.SetArrowIncrement(x.ArrowIncrement())
		s.SetMouseIncrement(x.MouseIncrement())
		s.SetTicksVisible(x.TicksVisible())
		s.SetOrientation(x.Orientation())
		s.SetTickPosition(x.TickPosition())
		s.SetBounds(0, 0, x.Width(), x.Height())
		return s
	case *wui.Panel:
		p := wui.NewPanel()
		p.SetBorderStyle(x.BorderStyle())
		p.SetBounds(0, 0, x.Width(), x.Height())
		return p
	case *wui.Label:
		l := wui.NewLabel()
		l.SetText(x.Text())
		l.SetAlignment(x.Alignment())
		l.SetBounds(0, 0, x.Width(), x.Height())
		return l
	case *wui.PaintBox:
		p := wui.NewPaintBox()
		p.SetBounds(0, 0, x.Width(), x.Height())
		return p
	case *wui.EditLine:
		e := wui.NewEditLine()
		e.SetBounds(0, 0, x.Width(), x.Height())
		e.SetText(x.Text())
		e.SetIsPassword(x.IsPassword())
		e.SetCharacterLimit(x.CharacterLimit())
		e.SetReadOnly(x.ReadOnly())
		return e
	case *wui.IntUpDown:
		n := wui.NewIntUpDown()
		n.SetBounds(0, 0, x.Width(), x.Height())
		n.SetMinMax(x.MinMax())
		n.SetValue(x.Value())
		return n
	case *wui.ComboBox:
		c := wui.NewComboBox()
		c.SetItems(x.Items())
		c.SetSelectedIndex(x.SelectedIndex())
		c.SetBounds(0, 0, x.Width(), x.Height())
		return c
	case *wui.ProgressBar:
		p := wui.NewProgressBar()
		p.SetValue(x.Value())
		p.SetVertical(x.Vertical())
		p.SetMovesForever(x.MovesForever())
		p.SetBounds(0, 0, x.Width(), x.Height())
		return p
	case *wui.FloatUpDown:
		f := wui.NewFloatUpDown()
		f.SetBounds(0, 0, x.Width(), x.Height())
		f.SetMinMax(x.MinMax())
		f.SetPrecision(x.Precision())
		f.SetValue(x.Value())
		return f
	default:
		panic("unhandled control type in cloneControl")
	}
}

type enabler interface {
	Enabled() bool
	SetEnabled(bool)
}

type visibler interface {
	Visible() bool
	SetVisible(bool)
}

func findContainerAt(c wui.Container, x, y int) (innerMost wui.Container, atX, atY int) {
	for _, child := range c.Children() {
		if container, ok := child.(wui.Container); ok {
			if innerContains(container, x, y) {
				dx, dy, _, _ := container.InnerBounds()
				return findContainerAt(container, x-dx, y-dy)
			}
		}
	}
	return c, x, y
}

// removeEmptyStrings changes the given input slice.
func removeEmptyStrings(items []string) []string {
	n := 0
	for i := range items {
		if items[i] != "" {
			items[n] = items[i]
			n++
		}
	}
	return items[:n]
}
