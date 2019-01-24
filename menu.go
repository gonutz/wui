//+build windows

package wui

import "github.com/gonutz/w32"

func NewMenu(name string) *Menu {
	return &Menu{name: name}
}

func NewMainMenu() *Menu {
	return &Menu{}
}

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

func (m *Menu) Add(item MenuItem) *Menu {
	m.items = append(m.items, item)
	return m
}

func NewMenuString(name string) *MenuString {
	return &MenuString{name: name}
}

// menu string

type MenuString struct {
	menu    w32.HMENU
	id      uint
	name    string
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
		w32.GetMenuItemInfo(m.menu, m.id, false, &info)
		if c {
			info.State = w32.MFS_CHECKED
		} else {
			info.State = w32.MFS_UNCHECKED
		}
		w32.SetMenuItemInfo(m.menu, m.id, false, &info)
	}
}

// separator

func NewMenuSeparator() MenuItem {
	return separator
}

type menuSeparator int

func (menuSeparator) isMenuItem() {}

var separator menuSeparator
