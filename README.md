Windows GUI Library
===================

This is a pure Go library to create native Windows GUIs.

See the [documentation](https://godoc.org/github.com/gonutz/wui) for details.

# Minimal Example

This is all the code you need to create a window (which does not do much).

```Go
package main

import "github.com/gonutz/wui"

func main() {
	wui.NewWindow().Show()
}
```
