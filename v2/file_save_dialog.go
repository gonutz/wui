package wui

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gonutz/w32/v2"
)

type FileSaveDialog struct {
	appendExt   bool
	filters     []uint16
	filterCount int
	filterIndex int
	initPath    string
	title       string
	exts        []string
}

func NewFileSaveDialog() *FileSaveDialog {
	return &FileSaveDialog{appendExt: true}
}

func (dlg *FileSaveDialog) SetAppendExt(ext bool) {
	dlg.appendExt = ext
}

func (dlg *FileSaveDialog) SetTitle(title string) {
	dlg.title = title
}

func (dlg *FileSaveDialog) SetInitialPath(path string) {
	dlg.initPath = path
}

func (dlg *FileSaveDialog) AddFilter(text, ext1 string, exts ...string) {
	text16, err := syscall.UTF16FromString(text)
	if err != nil {
		return
	}
	validateMask := func(ext string) string {
		ext = strings.TrimSpace(ext)
		if ext == "" {
			return "*.*"
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "*." + ext
		} else if !strings.HasPrefix(ext, "*") {
			ext = "*" + ext
		}
		return ext
	}
	mask := validateMask(ext1)
	for _, ext := range exts {
		mask += ";" + validateMask(ext)
	}
	mask16, err := syscall.UTF16FromString(mask)
	if err != nil {
		return
	}
	dlg.filters = append(dlg.filters, text16...)
	dlg.filters = append(dlg.filters, mask16...)
	dlg.filterCount++
	if ext1 == "" {
		dlg.exts = append(dlg.exts, "")
	} else {
		dlg.exts = append(
			dlg.exts,
			"."+strings.TrimPrefix(strings.ToLower(ext1), "."),
		)
	}
}

// SetFilterIndex sets the active filter, 0-indexed.
func (dlg *FileSaveDialog) SetFilterIndex(i int) {
	dlg.filterIndex = i
}

func (dlg *FileSaveDialog) Execute(parent *Window) (bool, string) {
	ok, buf, filterIndex := dlg.getSaveFileName(parent, w32.MAX_PATH+2)
	if ok {
		path := syscall.UTF16ToString(buf)
		if dlg.appendExt && 0 <= filterIndex && filterIndex < len(dlg.exts) {
			ext := dlg.exts[filterIndex]
			if !strings.HasSuffix(path, ext) {
				path += ext
			}
		}
		return true, path
	}
	return false, ""
}

func (dlg *FileSaveDialog) getSaveFileName(parent *Window, bufLen int) (bool, []uint16, int) {
	var owner w32.HWND
	if parent != nil {
		owner = parent.handle
	}

	dlg.filters = append(dlg.filters, 0)
	if dlg.filterIndex < 0 {
		dlg.filterIndex = 0
	}
	if dlg.filterIndex >= dlg.filterCount {
		dlg.filterIndex = dlg.filterCount - 1
	}

	var initDir *uint16
	var initDir16 []uint16
	filenameBuf := make([]uint16, bufLen)
	if dlg.initPath != "" {
		if info, err := os.Stat(dlg.initPath); err == nil && info.IsDir() {
			initDir16, err = syscall.UTF16FromString(dlg.initPath)
			if err == nil {
				initDir = &initDir16[0]
			}
		} else {
			dir, file := filepath.Split(dlg.initPath)

			initDir16, err = syscall.UTF16FromString(dir)
			if err == nil {
				initDir = &initDir16[0]
			}

			path, err := syscall.UTF16FromString(file)
			if err == nil {
				copy(filenameBuf, path)
			}
		}
	}

	var title16 []uint16
	var title *uint16
	if dlg.title != "" {
		var err error
		title16, err = syscall.UTF16FromString(dlg.title)
		if err == nil {
			title = &title16[0]
		}
	}

	ofn := &w32.OPENFILENAME{
		Owner:       owner,
		Filter:      &dlg.filters[0],
		FilterIndex: uint32(dlg.filterIndex + 1), // NOTE one-indexed
		File:        &filenameBuf[0],
		MaxFile:     uint32(len(filenameBuf)),
		InitialDir:  initDir,
		Title:       title,
		Flags: w32.OFN_ENABLESIZING | w32.OFN_EXPLORER | w32.OFN_LONGNAMES |
			w32.OFN_OVERWRITEPROMPT,
	}
	ok := w32.GetSaveFileName(ofn)
	return ok, filenameBuf, int(ofn.FilterIndex) - 1
}
