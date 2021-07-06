package wui

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/gonutz/w32/v2"
)

type FileOpenDialog struct {
	filters     []uint16
	filterCount int
	filterIndex int
	initPath    string
	title       string
	defaultExt  string
}

func NewFileOpenDialog() *FileOpenDialog {
	return &FileOpenDialog{}
}

func (dlg *FileOpenDialog) SetTitle(title string) {
	dlg.title = title
}

func (dlg *FileOpenDialog) SetInitialPath(path string) {
	dlg.initPath = path
}

func (dlg *FileOpenDialog) AddFilter(text, ext1 string, exts ...string) {
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
}

// SetFilterIndex sets the active filter, 0-indexed.
func (dlg *FileOpenDialog) SetFilterIndex(i int) {
	dlg.filterIndex = i
}

func (dlg *FileOpenDialog) ExecuteSingleSelection(parent *Window) (bool, string) {
	ok, buf := dlg.getOpenFileName(parent, w32.MAX_PATH+2, 0)
	if ok {
		return true, syscall.UTF16ToString(buf)
	}
	return false, ""
}

func (dlg *FileOpenDialog) ExecuteMultiSelection(parent *Window) (bool, []string) {
	ok, buf := dlg.getOpenFileName(parent, 65535, w32.OFN_ALLOWMULTISELECT)
	if ok {
		// parse multiple files, the format is 0-separated UTF-16 strings, first
		// comes the directory, then the file names, after the last file name
		// there are two zeros
		var dir string
		var files []string
		var start int
		for i := range buf[:len(buf)-1] {
			if buf[i] == 0 {
				part := buf[start:i]
				if start == 0 {
					dir = syscall.UTF16ToString(part)
				} else {
					file := syscall.UTF16ToString(part)
					files = append(files, filepath.Join(dir, file))
				}
				start = i + 1
				if buf[i+1] == 0 {
					break
				}
			}
		}
		if dir != "" && files == nil {
			// in this case, only one file was selected
			return true, []string{dir}
		}
		return true, files
	}
	return false, nil
}

func (dlg *FileOpenDialog) getOpenFileName(
	parent *Window, bufLen int, flags uint32,
) (bool, []uint16) {
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
			path, err := syscall.UTF16FromString(dlg.initPath)
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

	ok := w32.GetOpenFileName(&w32.OPENFILENAME{
		Owner:       owner,
		Filter:      &dlg.filters[0],
		FilterIndex: uint32(dlg.filterIndex + 1), // NOTE one-indexed
		File:        &filenameBuf[0],
		MaxFile:     uint32(len(filenameBuf)),
		InitialDir:  initDir,
		Title:       title,
		Flags: w32.OFN_ENABLESIZING | w32.OFN_EXPLORER |
			w32.OFN_FILEMUSTEXIST | w32.OFN_LONGNAMES | w32.OFN_PATHMUSTEXIST |
			w32.OFN_HIDEREADONLY | flags,
	})
	return ok, filenameBuf
}
