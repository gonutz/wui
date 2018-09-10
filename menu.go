//+build windows

package wui

func NewMenu(name string) *Menu {
	return &Menu{name: name}
}

type Menu struct {
	name  string
	items []MenuItem
}

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

type MenuString struct {
	name    string
	onClick func()
}

func (*MenuString) isMenuItem() {}

func (m *MenuString) SetOnClick(f func()) *MenuString {
	m.onClick = f
	return m
}

func NewMenuSeparator() MenuItem {
	return separator
}

type menuSeparator int

func (menuSeparator) isMenuItem() {}

var separator menuSeparator
