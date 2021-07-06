package wui

import "github.com/gonutz/w32/v2"

func MessageBox(caption, text string) {
	msgBox(caption, text, w32.MB_OK)
}

func MessageBoxError(caption, text string) {
	msgBox(caption, text, w32.MB_OK|w32.MB_ICONERROR)
}

func MessageBoxWarning(caption, text string) {
	msgBox(caption, text, w32.MB_OK|w32.MB_ICONWARNING)
}

func MessageBoxInfo(caption, text string) {
	msgBox(caption, text, w32.MB_OK|w32.MB_ICONINFORMATION)
}

func MessageBoxQuestion(caption, text string) {
	msgBox(caption, text, w32.MB_OK|w32.MB_ICONQUESTION)
}

func MessageBoxOKCancel(caption, text string) bool {
	return msgBox(caption, text, w32.MB_OKCANCEL) == w32.IDOK
}

func MessageBoxYesNo(caption, text string) bool {
	return msgBox(caption, text, w32.MB_YESNO|w32.MB_ICONQUESTION) == w32.IDYES
}

func MessageBoxCustom(caption, text string, flags uint) int {
	return msgBox(caption, text, flags)
}

func msgBox(caption, text string, flags uint) int {
	var handle w32.HWND
	parent := windows.top()
	if parent != nil {
		handle = parent.handle
	}
	return w32.MessageBox(handle, text, caption, w32.MB_TOPMOST|flags)
}
