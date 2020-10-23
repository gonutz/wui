package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gonutz/wui"
)

type property struct {
	name     string
	combines []string
}

func prop(name string, combines ...string) property {
	return property{name: name, combines: combines}
}

func commonPropertiesPlus(plus ...property) []property {
	return append([]property{
		prop("Enabled"),
		prop("Visible"),
		prop("HorizontalAnchor"),
		prop("VerticalAnchor"),
		prop("Anchors", "HorizontalAnchor", "VerticalAnchor"),
		prop("X"),
		prop("Y"),
		prop("Position", "X", "Y"),
		prop("Width"),
		prop("Height"),
		prop("Size", "Width", "Height"),
		prop("Bounds", "Position", "Size"),
	},
		plus...)
}

var properties = map[interface{}][]property{
	wui.NewButton(): commonPropertiesPlus(
		prop("Text"),
	),

	wui.NewLabel(): commonPropertiesPlus(
		prop("Text"),
		prop("Alignment"),
	),

	wui.NewCheckBox(): commonPropertiesPlus(
		prop("Text"),
		prop("Checked"),
	),

	wui.NewRadioButton(): commonPropertiesPlus(
		prop("Text"),
		prop("Checked"),
	),

	wui.NewSlider(): commonPropertiesPlus(
		prop("ArrowIncrement"),
		prop("MouseIncrement"),
		prop("Cursor"),
		prop("Min"),
		prop("Max"),
		prop("MinMax", "Min", "Max"),
		prop("Orientation"),
		prop("TickFrequency"),
		prop("TickPosition"),
		prop("TicksVisible"),
	),

	wui.NewPanel(): commonPropertiesPlus(
		prop("BorderStyle"),
	),

	wui.NewPaintBox(): commonPropertiesPlus(),

	wui.NewEditLine(): commonPropertiesPlus(
		prop("Text"),
		prop("CharacterLimit"),
		prop("IsPassword"),
		prop("ReadOnly"),
	),
}

func generateProperties(variable string, control interface{}) []string {
	for def, props := range properties {
		if reflect.TypeOf(control) == reflect.TypeOf(def) {
			return genProps(variable, control, def, props)
		}
	}
	panic("no properties found for type " + reflect.TypeOf(control).String())

}

func genProps(variable string, c, def interface{}, props []property) []string {
	var s []string
	control := reflect.ValueOf(c)
	wasSet := make(map[string]bool)
	for _, p := range props {
		if len(p.combines) > 0 {
			if !containsAll(wasSet, p.combines) {
				continue
			}
		}
		ours := control.MethodByName(p.name).Call(nil)
		defaults := reflect.ValueOf(def).MethodByName(p.name).Call(nil)
		if !equal(ours, defaults) {
			s = append(s, variable+".Set"+p.name+"("+toGo(ours)+")")
			wasSet[p.name] = true
			for _, previous := range p.combines {
				s = removeFirstThatContains(s, variable+".Set"+previous+"(")
			}
		}
	}
	return s
}

func containsAll(set map[string]bool, list []string) bool {
	for _, s := range list {
		if !set[s] {
			return false
		}
	}
	return true
}

func removeFirstThatContains(from []string, pattern string) []string {
	for i := range from {
		if strings.Contains(from[i], pattern) {
			return append(from[:i], from[i+1:]...)
		}
	}
	return from
}

func equal(a, b []reflect.Value) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i].Interface(), b[i].Interface()) {
			return false
		}
	}
	return true
}

func toGo(args []reflect.Value) string {
	var asGo []string
	for _, arg := range args {
		var s string
		if arg.Kind() == reflect.String {
			s = fmt.Sprintf("%q", arg.String())
		} else {
			s = fmt.Sprint(arg)
		}
		asGo = append(asGo, s)
	}
	return strings.Join(asGo, ", ")
}
