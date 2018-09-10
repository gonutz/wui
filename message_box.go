//+build windows

package wui

import "github.com/gonutz/w32"

func MessageBox(caption, text string) {
	w32.MessageBox(0, text, caption, w32.MB_OK|w32.MB_TOPMOST)
}

func MessageBoxError(caption, text string) {
	w32.MessageBox(0, text, caption, w32.MB_OK|w32.MB_ICONERROR|w32.MB_TOPMOST)
}

func MessageBoxOKCancel(caption, text string) bool {
	return w32.MessageBox(0, text, caption, w32.MB_OKCANCEL|w32.MB_TOPMOST) == w32.IDOK
}

func MessageBoxYesNo(caption, text string) bool {
	return w32.MessageBox(0, text, caption, w32.MB_YESNO|w32.MB_TOPMOST) == w32.IDYES
}
