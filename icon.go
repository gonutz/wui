package wui

import (
	"encoding/binary"
	"errors"
	"image"
	"io"
	"io/ioutil"
	"os"
	"syscall"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

// Icon holds a window icon. You can use a pre-defined Icon... variable (see
// below) or create a custom icon with NewIconFromImage, NewIconFromExeResource,
// NewIconFromFile or NewIconFromReader.
type Icon struct {
	handle w32.HICON
}

var (
	// IconApplication is the default application icon.
	IconApplication = loadIcon(w32.IDI_APPLICATION)
	// IconQuestion is a question mark icon.
	IconQuestion = loadIcon(w32.IDI_QUESTION)
	// IconWinLogo is usually the same as IconApplication but on older Windows
	// versions (like Windows 2000) this was a Windows icon.
	IconWinLogo = loadIcon(w32.IDI_WINLOGO)
	// IconShield is the icon that comes up when Windows asks admin permission.
	// It looks like a knight's shield.
	IconShield = loadIcon(w32.IDI_SHIELD)
	// IconWarning is an exclamation mark icon.
	IconWarning = loadIcon(w32.IDI_WARNING)
	// IconError is a red cross icon.
	IconError = loadIcon(w32.IDI_ERROR)
	// IconInformation is a blue 'i' for information.
	IconInformation = loadIcon(w32.IDI_INFORMATION)
)

func loadIcon(id uint16) *Icon {
	return &Icon{handle: w32.LoadIcon(0, w32.MakeIntResource(id))}
}

// IconInformation creates an icon from an image. The image should have a
// standard icon size, typically square and a power of 2, see
// https://docs.microsoft.com/en-us/windows/win32/uxguide/vis-icons
func NewIconFromImage(img image.Image) (*Icon, error) {
	// We create an icon structure in the form Windows likes which consists of a
	// BITMAPINFOHEADER (see
	// https://docs.microsoft.com/en-us/previous-versions/dd183376(v=vs.85))
	// followed by the image data. We need 4 byte BGRA color order while the Go
	// image gives use RGBA, see the re-ordering in the for loop below.
	// After the image data comes a mask which has 1 bit for each pixel. We want
	// to use each bit so we set them all to 1.
	// All this is put into one single byte array and then passed to
	// CreateIconFromResource.

	size := img.Bounds().Size()
	const headerLen = 40                   // Size of BITMAPINFOHEADER.
	maskLen := (size.Y * (size.X + 7) / 8) // Round up to whole bytes.
	iconLen := headerLen + size.X*size.Y*4 + maskLen
	iconData := make([]byte, iconLen)

	// Write the BITMAPINFOHEADER.
	binary.LittleEndian.PutUint32(iconData[0:], headerLen)
	binary.LittleEndian.PutUint32(iconData[4:], uint32(size.X))
	binary.LittleEndian.PutUint32(iconData[8:], uint32(size.Y*2))
	binary.LittleEndian.PutUint16(iconData[12:], 1)
	binary.LittleEndian.PutUint16(iconData[14:], 32)
	binary.LittleEndian.PutUint32(iconData[16:], w32.BI_RGB)
	binary.LittleEndian.PutUint32(iconData[20:], uint32(size.X*size.Y*4))
	// 4 uint32 0s follow, iconData[40:] is where the image data starts.
	dest := iconData[headerLen:]
	// Write the pixels upside down into the bitmap buffer.
	b := img.Bounds()
	for y := b.Max.Y - 1; y >= b.Min.Y; y-- {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			dest[0] = byte(b >> 8)
			dest[1] = byte(g >> 8)
			dest[2] = byte(r >> 8)
			dest[3] = byte(a >> 8)
			dest = dest[4:]
		}
	}

	// Write the mask. Transparency comes from the image's alpha channel, thus
	// we can set the mask to all 1s.
	for i := range dest {
		dest[i] = 0xFF
	}

	icon := w32.CreateIconFromResource(
		unsafe.Pointer(&iconData[0]),
		uint32(len(iconData)),
		true, // true for icons, false for cursors.
		// 0x30000 is a magic constant from the docs:
		// https://docs.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-createiconfromresource
		0x30000,
	)
	if icon == 0 {
		return nil, errors.New("wui.NewIconFromImage: CreateIconFromResource returned 0 handle")
	}
	return &Icon{handle: icon}, nil
}

// NewIconFromExeResource loads an icon that was compiled into the executable.
// In Go you can create .syso files that contain resources. Each resource will
// get a unique ID. See for example the rsrc tool which can create .syso files:
// https://github.com/gonutz/rsrc
func NewIconFromExeResource(resourceID int) (*Icon, error) {
	icon := w32.HICON(w32.LoadImage(
		w32.GetModuleHandle(""),
		w32.MakeIntResource(uint16(resourceID)),
		w32.IMAGE_ICON,
		0,
		0,
		w32.LR_DEFAULTSIZE|w32.LR_SHARED,
	))
	if icon == 0 {
		return nil, errors.New("wui.NewIconFromExeResource: LoadImage returned 0 handle")
	}
	return &Icon{handle: icon}, nil
}

// NewIconFromFile loads an icon from disk. The format must be .ico, not an
// image.
func NewIconFromFile(path string) (*Icon, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, errors.New("wui.NewIconFromFile: invalid path: " + err.Error())
	}
	icon := w32.HICON(w32.LoadImage(
		0,
		p,
		w32.IMAGE_ICON,
		0, 0,
		w32.LR_LOADFROMFILE|w32.LR_DEFAULTSIZE,
	))
	if icon == 0 {
		return nil, errors.New("wui.NewIconFromFile: LoadImage returned 0 handle")
	}
	return &Icon{handle: icon}, nil
}

// NewIconFromReader loads an icon from the given reader. The format must be
// .ico, not an image.
func NewIconFromReader(r io.Reader) (*Icon, error) {
	// For some reason CreateIconFromResource does not work for multi-resolution
	// icons. It will only ever use the first of the icons.
	// Instead create a temporary icon file, load it, then delete it.
	f, err := ioutil.TempFile("", "icon_")
	if err != nil {
		return nil, errors.New("wui.NewIconFromReader: unable to create temporary icon file: " + err.Error())
	}
	iconPath := f.Name()
	defer os.Remove(iconPath)
	_, err = io.Copy(f, r)
	f.Close() // Close before trying to load from it.
	if err != nil {
		return nil, errors.New("wui.NewIconFromReader: copying data to temporary icon file failed: " + err.Error())
	}
	icon, err := NewIconFromFile(iconPath)
	if err != nil {
		return nil, errors.New("wui.NewIconFromReader: " + err.Error())
	}
	return icon, nil
}
