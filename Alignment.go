package giu

import (
	"fmt"
	"image"

	"github.com/AllenDang/imgui-go"
)

type AlignmentType byte

const (
	AlignLeft AlignmentType = iota
	AlignCenter
	AlignRight
)

type AlignmentSetter struct {
	alignType AlignmentType
	layout    Layout
	id        string
}

// Align sets widgets alignment.
// usage: see examples/align
//
// - BUG: DatePickerWidget doesn't work properly
// - BUG: there is some bug with SelectableWidget
// - BUG: ComboWidget and ComboCustomWidgets doesn't work properly.
func Align(at AlignmentType) *AlignmentSetter {
	return &AlignmentSetter{
		alignType: at,
		id:        GenAutoID("alignSetter"),
	}
}

// To sets a layout, alignment should be applied to.
func (a *AlignmentSetter) To(widgets ...Widget) *AlignmentSetter {
	a.layout = Layout(widgets)
	return a
}

// ID allows to manually set AlignmentSetter ID (it shouldn't be used
// in a normal conditions).
func (a *AlignmentSetter) ID(id string) *AlignmentSetter {
	a.id = id
	return a
}

func (a *AlignmentSetter) Build() {
	if a.layout == nil {
		return
	}

	a.layout.Range(func(item Widget) {
		// if item is inil, just skip it
		if item == nil {
			return
		}

		switch item.(type) {
		// ok, it doesn't make sense to align again :-)
		case *AlignmentSetter:
			item.Build()
			return
		// there is a bug with selectables and combos, so skip them for now
		case *SelectableWidget, *ComboWidget, *ComboCustomWidget:
			item.Build()
			return
		}

		currentPos := GetCursorPos()
		w := GetWidgetWidth(item)
		availableW, _ := GetAvailableRegion()
		// we need to increase available region by 2 * window padding (X),
		// because GetCursorPos considers it
		paddingW, _ := GetWindowPadding()
		availableW += 2 * paddingW

		// set cursor position to align the widget
		switch a.alignType {
		case AlignLeft:
			SetCursorPos(currentPos)
		case AlignCenter:
			SetCursorPos(image.Pt(int(availableW/2-w/2), currentPos.Y))
		case AlignRight:
			SetCursorPos(image.Pt(int(availableW-w), currentPos.Y))
		default:
			panic(fmt.Sprintf("giu: (*AlignSetter).Build: unknown align type %d", a.alignType))
		}

		// build aligned widget
		item.Build()
	})
}

// GetWidgetWidth returns a width of widget
// NOTE: THIS IS A BETA SOLUTION and may contain bugs
// in most cases, you may want to use supported by imgui GetItemRectSize.
// There is an upstream issue for this problem:
// https://github.com/ocornut/imgui/issues/3714
//
// This function is just a workaround used in giu.
//
// NOTE: user-definied widgets, which contains more than one
// giu widget will be processed incorrectly (only width of the last built
// widget will be processed)
//
// here is a list of known bugs:
// - BUG: clicking bug - when widget is clickable, it is unable to be
// clicked see:
//   - https://github.com/AllenDang/giu/issues/341
//   - https://github.com/ocornut/imgui/issues/4588
// - BUG: text pasted into input text is pasted twice
//   (see: https://github.com/AllenDang/giu/issues/340)
//
// if you find anything else, please report it on
// https://github.com/AllenDang/giu Any contribution is appreciated!
func GetWidgetWidth(w Widget) (result float32) {
	// save cursor position before rendering
	currentPos := GetCursorPos()

	// render widget in `dry` mode
	imgui.PushStyleVarFloat(imgui.StyleVarAlpha, 0)
	w.Build()
	imgui.PopStyleVar()

	// save widget's width
	// check cursor position
	imgui.SameLine()
	spacingW, _ := GetItemSpacing()
	result = float32(GetCursorPos().X-currentPos.X) - spacingW

	SetCursorPos(currentPos)

	return result
}
