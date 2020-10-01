//+build windows

package wui

import (
	"syscall"
	"unsafe"

	"github.com/gonutz/w32"
)

// NewMainMenu returns a new menu bar that can be added to a Window. You can add
// sub-menus to it with Menu.Add.
func NewMainMenu() *Menu {
	return &Menu{}
}

// NewMenu returns a new Menu with the given text. Add MenuItems to it using
// Menu.Add.
//
// Place an ampersand in the name to underline the following character, e.g.
// "&File" to underline the "F". This will be the character used when the user
// uses the alt and arrow keys to navigate the menu.
func NewMenu(name string) *Menu {
	return &Menu{name: name}
}

// Menu is a named container for MenuItems. It is not executable, clicking a
// Menu will expand it and show its children. See NewMainMenu and NewMenu.
type Menu struct {
	name  string
	items []MenuItem
}

// MenuItem is something that can go into a menu. Possible such things can be
// created via: NewMenu, NewMenuString, NewMenuSeparator.
type MenuItem interface {
	isMenuItem()
}

func (*Menu) isMenuItem() {}

// Add appends the given MenuItem to the Menu.
func (m *Menu) Add(item MenuItem) *Menu {
	m.items = append(m.items, item)
	return m
}

// NewMenuString creates a new executable menu item with the given text.
//
// Insert an ampersand to underline the following character, e.g. "New &File" to
// underline the "F". This will be the character used when the user uses the alt
// and arrow keys to navigate the menu.
//
// If you want to display a shortcut text next to the menu, right aligned,
// insert a tab and then your text, e.g. "Open File\tCtrl+O" to have "Open File"
// on the left and "Ctrl+O" on the right.
func NewMenuString(text string) *MenuString {
	return &MenuString{text: text}
}

// MenuString is an executable menu item, see NewMenuString.
type MenuString struct {
	window  w32.HWND
	menu    w32.HMENU
	id      uint
	text    string
	checked bool
	onClick func()
}

func (*MenuString) isMenuItem() {}

func (m *MenuString) SetOnClick(f func()) *MenuString {
	m.onClick = f
	return m
}

func (m *MenuString) OnClick() func() {
	return m.onClick
}

func (m *MenuString) Checked() bool {
	return m.checked
}

func (m *MenuString) SetChecked(c bool) {
	m.checked = c
	if m.menu != 0 {
		var info w32.MENUITEMINFO
		info.Mask = w32.MIIM_STATE
		if c {
			info.State = w32.MFS_CHECKED
		} else {
			info.State = w32.MFS_UNCHECKED
		}
		w32.SetMenuItemInfo(m.menu, m.id, false, &info)
	}
}

func (m *MenuString) Text() string {
	return m.text
}

func (m *MenuString) SetText(s string) {
	m.text = s
	if m.menu != 0 {
		var info w32.MENUITEMINFO
		info.Mask = w32.MIIM_STRING
		info.TypeData = uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
		w32.SetMenuItemInfo(m.menu, m.id, false, &info)
		w32.DrawMenuBar(m.window)
	}
}

// NewMenuSeparator returns a horizontal line separating regions in a menu.
func NewMenuSeparator() MenuItem {
	return separator
}

type menuSeparator int

func (menuSeparator) isMenuItem() {}

var separator menuSeparator
