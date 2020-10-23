package main

import (
	"testing"

	"github.com/gonutz/check"
	"github.com/gonutz/wui"
)

func TestButtonPropertyGeneration(t *testing.T) {
	var b *wui.Button
	checkProperties := func(want ...string) {
		t.Helper()
		check.Eq(t, generateProperties("b", b), want)
	}

	b = wui.NewButton()
	checkProperties()

	b = wui.NewButton()
	b.SetEnabled(false)
	checkProperties(`b.SetEnabled(false)`)

	b = wui.NewButton()
	b.SetText("a button")
	checkProperties(`b.SetText("a button")`)

	b = wui.NewButton()
	b.SetX(123)
	checkProperties(`b.SetX(123)`)

	b = wui.NewButton()
	b.SetPosition(123, 456)
	checkProperties(`b.SetPosition(123, 456)`)

	b = wui.NewButton()
	b.SetBounds(123, 456, 789, 654)
	checkProperties(`b.SetBounds(123, 456, 789, 654)`)

	b = wui.NewButton()
	b.SetWidth(123)
	checkProperties(`b.SetWidth(123)`)

	b = wui.NewButton()
	b.SetX(123)
	b.SetWidth(456)
	checkProperties(`b.SetX(123)`, `b.SetWidth(456)`)

	b = wui.NewButton()
	b.SetHorizontalAnchor(wui.AnchorMinAndMax)
	checkProperties(`b.SetHorizontalAnchor(wui.AnchorMinAndMax)`)

	b = wui.NewButton()
	b.SetAnchors(wui.AnchorCenter, wui.AnchorMaxAndCenter)
	checkProperties(`b.SetAnchors(wui.AnchorCenter, wui.AnchorMaxAndCenter)`)
}

func TestCheckBoxPropertyGeneration(t *testing.T) {
	var c *wui.CheckBox
	checkProperties := func(want ...string) {
		t.Helper()
		check.Eq(t, generateProperties("c", c), want)
	}

	c = wui.NewCheckBox()
	checkProperties()

	c = wui.NewCheckBox()
	c.SetChecked(true)
	checkProperties(`c.SetChecked(true)`)
}
