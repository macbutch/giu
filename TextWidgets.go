package giu

import (
	"fmt"
	"math"

	"github.com/AllenDang/imgui-go"
	"github.com/sahilm/fuzzy"
)

var _ Widget = &InputTextMultilineWidget{}

// InputTextMultilineWidget represents multiline text input widget
// see examples/widgets/.
type InputTextMultilineWidget struct {
	label         string
	text          *string
	width, height float32
	flags         InputTextFlags
	cb            imgui.InputTextCallback
	onChange      func()
}

// InputTextMultiline creates InputTextMultilineWidget.
func InputTextMultiline(text *string) *InputTextMultilineWidget {
	return &InputTextMultilineWidget{
		text:     text,
		width:    0,
		height:   0,
		flags:    0,
		cb:       nil,
		onChange: nil,
		label:    GenAutoID("##InputTextMultiline"),
	}
}

// Label sets input field label.
func (i *InputTextMultilineWidget) Label(label string) *InputTextMultilineWidget {
	i.label = label
	return i
}

// Labelf is formatting version of Label.
func (i *InputTextMultilineWidget) Labelf(format string, args ...interface{}) *InputTextMultilineWidget {
	return i.Label(fmt.Sprintf(format, args...))
}

// Build implements Widget interface.
func (i *InputTextMultilineWidget) Build() {
	if imgui.InputTextMultilineV(
		tStr(i.label),
		tStrPtr(i.text),
		imgui.Vec2{
			X: i.width,
			Y: i.height,
		},
		int(i.flags), i.cb,
	) && i.onChange != nil {
		i.onChange()
	}
}

// Flags sets InputTextFlags (see Flags.go).
func (i *InputTextMultilineWidget) Flags(flags InputTextFlags) *InputTextMultilineWidget {
	i.flags = flags
	return i
}

// Callback sets imgui.InputTextCallback.
func (i *InputTextMultilineWidget) Callback(cb imgui.InputTextCallback) *InputTextMultilineWidget {
	i.cb = cb
	return i
}

// OnChange set callback called when user action taken on input text field (when text was changed).
func (i *InputTextMultilineWidget) OnChange(onChange func()) *InputTextMultilineWidget {
	i.onChange = onChange
	return i
}

// Size sets input field size.
func (i *InputTextMultilineWidget) Size(width, height float32) *InputTextMultilineWidget {
	i.width, i.height = width, height
	return i
}

var _ Widget = &BulletWidget{}

// BulletWidget adds a small, white dot (bullet).
// useful in enumerations.
type BulletWidget struct{}

// Bullet creates a bullet widget.
func Bullet() *BulletWidget {
	return &BulletWidget{}
}

// Build implements Widget interface.
func (b *BulletWidget) Build() {
	imgui.Bullet()
}

var _ Widget = &BulletTextWidget{}

// BulletTextWidget does similar to BulletWidget, but allows
// to add a text after a bullet. Very useful to create lists.
type BulletTextWidget struct {
	text string
}

// BulletText creates bulletTextWidget.
func BulletText(text string) *BulletTextWidget {
	return &BulletTextWidget{
		text: tStr(text),
	}
}

// BulletTextf is a formatting version of BulletText.
func BulletTextf(format string, args ...interface{}) *BulletTextWidget {
	return BulletText(fmt.Sprintf(format, args...))
}

// Build implements Widget interface.
func (bt *BulletTextWidget) Build() {
	imgui.BulletText(bt.text)
}

var _ Widget = &InputTextWidget{}

type InputTextWidget struct {
	label      string
	hint       string
	value      *string
	width      float32
	candidates []string
	flags      InputTextFlags
	cb         imgui.InputTextCallback
	onChange   func()
}

type inputTextState struct {
	autoCompleteCandidates fuzzy.Matches
}

func (s *inputTextState) Dispose() {
	s.autoCompleteCandidates = nil
}

func InputText(value *string) *InputTextWidget {
	return &InputTextWidget{
		label:    GenAutoID("##InputText"),
		hint:     "",
		value:    value,
		width:    0,
		flags:    0,
		cb:       nil,
		onChange: nil,
	}
}

func (i *InputTextWidget) Label(label string) *InputTextWidget {
	i.label = tStr(label)
	return i
}

func (i *InputTextWidget) Labelf(format string, args ...interface{}) *InputTextWidget {
	return i.Label(fmt.Sprintf(format, args...))
}

// AutoComplete enables auto complete popup by using fuzzy search of current value against candidates
// Press enter to confirm the first candidate.
func (i *InputTextWidget) AutoComplete(candidates []string) *InputTextWidget {
	i.candidates = candidates
	return i
}

func (i *InputTextWidget) Hint(hint string) *InputTextWidget {
	i.hint = tStr(hint)
	return i
}

func (i *InputTextWidget) Size(width float32) *InputTextWidget {
	i.width = width
	return i
}

func (i *InputTextWidget) Flags(flags InputTextFlags) *InputTextWidget {
	i.flags = flags
	return i
}

func (i *InputTextWidget) Callback(cb imgui.InputTextCallback) *InputTextWidget {
	i.cb = cb
	return i
}

func (i *InputTextWidget) OnChange(onChange func()) *InputTextWidget {
	i.onChange = onChange
	return i
}

// Build implements Widget interface.
func (i *InputTextWidget) Build() {
	// Get state
	var state *inputTextState
	if s := Context.GetState(i.label); s == nil {
		state = &inputTextState{}
		Context.SetState(i.label, state)
	} else {
		var isOk bool
		state, isOk = s.(*inputTextState)
		Assert(isOk, "InputTextWidget", "Build", "wrong state type recovered.")
	}

	if i.width != 0 {
		PushItemWidth(i.width)
		defer PopItemWidth()
	}

	isChanged := imgui.InputTextWithHint(i.label, i.hint, tStrPtr(i.value), int(i.flags), i.cb)

	if isChanged && i.onChange != nil {
		i.onChange()
	}

	if isChanged {
		// Enable auto complete
		if len(i.candidates) > 0 {
			matches := fuzzy.Find(*i.value, i.candidates)
			if matches.Len() > 0 {
				size := int(math.Min(5, float64(matches.Len())))
				matches = matches[:size]

				state.autoCompleteCandidates = matches
			}
		}
	}

	// Draw autocomplete list
	if len(state.autoCompleteCandidates) > 0 {
		labels := make(Layout, len(state.autoCompleteCandidates))
		for i, m := range state.autoCompleteCandidates {
			labels[i] = Label(m.Str)
		}

		SetNextWindowPos(imgui.GetItemRectMin().X, imgui.GetItemRectMax().Y)
		imgui.BeginTooltip()
		labels.Build()
		imgui.EndTooltip()

		// Press enter will replace value string with first match candidate
		if IsKeyPressed(KeyEnter) {
			*i.value = state.autoCompleteCandidates[0].Str
			state.autoCompleteCandidates = nil
		}
	}
}

var _ Widget = &InputIntWidget{}

type InputIntWidget struct {
	label    string
	value    *int32
	width    float32
	flags    InputTextFlags
	onChange func()
}

func InputInt(value *int32) *InputIntWidget {
	return &InputIntWidget{
		label:    GenAutoID("##InputInt"),
		value:    value,
		width:    0,
		flags:    0,
		onChange: nil,
	}
}

func (i *InputIntWidget) Label(label string) *InputIntWidget {
	i.label = tStr(label)
	return i
}

func (i *InputIntWidget) Labelf(format string, args ...interface{}) *InputIntWidget {
	return i.Label(fmt.Sprintf(format, args...))
}

func (i *InputIntWidget) Size(width float32) *InputIntWidget {
	i.width = width
	return i
}

func (i *InputIntWidget) Flags(flags InputTextFlags) *InputIntWidget {
	i.flags = flags
	return i
}

func (i *InputIntWidget) OnChange(onChange func()) *InputIntWidget {
	i.onChange = onChange
	return i
}

// Build implements Widget interface.
func (i *InputIntWidget) Build() {
	if i.width != 0 {
		PushItemWidth(i.width)
		defer PopItemWidth()
	}

	if imgui.InputIntV(i.label, i.value, 0, 100, int(i.flags)) && i.onChange != nil {
		i.onChange()
	}
}

var _ Widget = &InputFloatWidget{}

type InputFloatWidget struct {
	label    string
	value    *float32
	width    float32
	flags    InputTextFlags
	format   string
	onChange func()
}

func InputFloat(value *float32) *InputFloatWidget {
	return &InputFloatWidget{
		label:    GenAutoID("##InputFloatWidget"),
		width:    0,
		value:    value,
		format:   "%.3f",
		flags:    0,
		onChange: nil,
	}
}

func (i *InputFloatWidget) Label(label string) *InputFloatWidget {
	i.label = tStr(label)
	return i
}

func (i *InputFloatWidget) Labelf(format string, args ...interface{}) *InputFloatWidget {
	return i.Label(fmt.Sprintf(format, args...))
}

func (i *InputFloatWidget) Size(width float32) *InputFloatWidget {
	i.width = width
	return i
}

func (i *InputFloatWidget) Flags(flags InputTextFlags) *InputFloatWidget {
	i.flags = flags
	return i
}

func (i *InputFloatWidget) Format(format string) *InputFloatWidget {
	i.format = format
	return i
}

func (i *InputFloatWidget) OnChange(onChange func()) *InputFloatWidget {
	i.onChange = onChange
	return i
}

// Build implements Widget interface.
func (i *InputFloatWidget) Build() {
	if i.width != 0 {
		PushItemWidth(i.width)
		defer PopItemWidth()
	}

	if imgui.InputFloatV(i.label, i.value, 0, 0, i.format, int(i.flags)) && i.onChange != nil {
		i.onChange()
	}
}

var _ Widget = &LabelWidget{}

type LabelWidget struct {
	label    string
	fontInfo *FontInfo
	wrapped  bool
}

func Label(label string) *LabelWidget {
	return &LabelWidget{
		label:   tStr(label),
		wrapped: false,
	}
}

func Labelf(format string, args ...interface{}) *LabelWidget {
	return Label(fmt.Sprintf(format, args...))
}

func (l *LabelWidget) Wrapped(wrapped bool) *LabelWidget {
	l.wrapped = wrapped
	return l
}

func (l *LabelWidget) Font(font *FontInfo) *LabelWidget {
	l.fontInfo = font
	return l
}

// Build implements Widget interface.
func (l *LabelWidget) Build() {
	if l.wrapped {
		PushTextWrapPos()
		defer PopTextWrapPos()
	}

	if l.fontInfo != nil {
		if PushFont(l.fontInfo) {
			defer PopFont()
		}
	}

	imgui.Text(l.label)
}
