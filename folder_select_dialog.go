//go:build windows
// +build windows

package wui

import (
	"syscall"

	"github.com/gonutz/w32/v2"
)

type FolderSelectDialog struct {
	title string
}

func NewFolderSelectDialog() *FolderSelectDialog {
	return &FolderSelectDialog{}
}

func (dlg *FolderSelectDialog) SetTitle(title string) {
	dlg.title = title
}

func (dlg *FolderSelectDialog) Execute(parent *Window) (bool, string) {
	var owner w32.HWND
	if parent != nil {
		owner = parent.handle
	}

	title, _ := syscall.UTF16PtrFromString(dlg.title)

	idl := w32.SHBrowseForFolder(&w32.BROWSEINFO{
		Owner: owner,
		Title: title,
		Flags: w32.BIF_NEWDIALOGSTYLE,
	})
	folder := w32.SHGetPathFromIDList(idl)

	return idl != 0, folder

}
