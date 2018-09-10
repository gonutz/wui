//+build ignore

package main

import "github.com/gonutz/wui"

func main() {
	var window *wui.Window

	close := wui.NewButton()
	close.SetBounds(320-50, 240-12, 100, 24)
	close.SetText("Close")
	close.SetOnClick(func() {
		close.SetBounds(0, 0, 100, 50)
	})

	window = wui.NewWindow()
	window.SetTitle("Test")
	window.SetClientSize(640, 480)
	window.Add(close)
	window.Show()
}
