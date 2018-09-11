//+build windows

package wui

import "github.com/gonutz/w32"

func MessageBox(parent *Window, caption, text string) {
	msgBox(parent, caption, text, w32.MB_OK)
}

func MessageBoxError(parent *Window, caption, text string) {
	msgBox(parent, caption, text, w32.MB_OK|w32.MB_ICONERROR)
}

func MessageBoxOKCancel(parent *Window, caption, text string) bool {
	return msgBox(parent, caption, text, w32.MB_OKCANCEL) == w32.IDOK
}

func MessageBoxYesNo(parent *Window, caption, text string) bool {
	return msgBox(parent, caption, text, w32.MB_YESNO) == w32.IDYES
}

func msgBox(parent *Window, caption, text string, flags uint) int {
	var handle w32.HWND
	if parent != nil {
		handle = parent.handle
	}
	return w32.MessageBox(handle, text, caption, w32.MB_TOPMOST|flags)
}
