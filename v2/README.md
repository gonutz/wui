Windows GUI Library
===================

This is a pure Go library to create native Windows GUIs. See [the online documentation](https://pkg.go.dev/github.com/gonutz/wui/v2) for details.

# Minimal Example

This is all the code you need to create a window (which does not do much).

```Go
package main

import "github.com/gonutz/wui/v2"

func main() {
	wui.NewWindow().Show()
}
```

# The Designer

I am currently working on a graphical designer. It is located under `github.com/gonutz/wui/v2/cmd/designer`.

At the moment it lets you place widgets graphically and generate a Go main file from it (using the `Save` menu). You can also run a preview with `Ctrl+R`.

There is no way to read the generated code back in at the moment. Right now it is a tool to place things and generate code from it.
