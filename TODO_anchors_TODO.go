//+build ignore

package main

import "github.com/gonutz/wui"

func main() {
	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -11})

	w := wui.NewWindow()
	w.SetFont(font)
	w.SetClientSize(400, 300)

	b := wui.NewButton()
	b.SetBounds(10, 10, 380, 25)
	b.AnchorLeftAndRight()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(10, 50, 380, 25)
	b.AnchorRight()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(10, 90, 75, 25)
	b.AnchorTopAndBottom()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(110, 90, 75, 25)
	b.AnchorBottom()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(210, 130, 75, 25)
	b.AnchorVerticalCenter()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(310, 170, 75, 25)
	b.AnchorBottomAndCenter()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(310, 130, 75, 25)
	b.AnchorTopAndCenter()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(210, 170, 75, 25)
	b.AnchorVerticalCenter()
	b.SetText("Hello World!")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(120-2, 200, 75, 25)
	b.AnchorBottom()
	b.AnchorLeftAndCenter()
	b.SetText("OK")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(200+2, 200, 75, 25)
	b.AnchorBottom()
	b.AnchorRightAndCenter()
	b.SetText("Cancel")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(120-2, 250, 75, 25)
	b.AnchorBottom()
	b.AnchorHorizontalCenter()
	b.SetText("OK")
	w.Add(b)

	b = wui.NewButton()
	b.SetBounds(200+2, 250, 75, 25)
	b.AnchorBottom()
	b.AnchorHorizontalCenter()
	b.SetText("Cancel")
	w.Add(b)

	w.Show()
}
