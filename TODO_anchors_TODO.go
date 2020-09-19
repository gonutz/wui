//+build ignore

package main

import "github.com/gonutz/wui"

func main() {
	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -11})

	w := wui.NewWindow()
	w.SetFont(font)
	w.SetClientSize(300, 300)

	b := wui.NewButton()
	b.SetBounds(10, 10, 280, 30)
	b.AnchorLeftAndRight()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(10, 50, 280, 30)
	b.AnchorRight()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(10, 90, 100, 30)
	b.AnchorTopAndBottom()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(120, 90, 100, 30)
	b.AnchorBottom()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(50-2, 200, 100, 30)
	b.AnchorBottom()
	b.AnchorLeftAndCenter()
	b.SetText("OK")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(150+2, 200, 100, 30)
	b.AnchorBottom()
	b.AnchorRightAndCenter()
	b.SetText("Cancel")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(50-2, 250, 100, 30)
	b.AnchorBottom()
	b.AnchorHorizontalCenter()
	b.SetText("OK")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(150+2, 250, 100, 30)
	b.AnchorBottom()
	b.AnchorHorizontalCenter()
	b.SetText("Cancel")
	w.Add(b)

	w.Show()
}
