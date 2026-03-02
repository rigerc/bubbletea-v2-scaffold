package screens

import (
	"fmt"
	"reflect"
	"strings"

	"scaffold/config"
	"scaffold/internal/ui/theme"

	"charm.land/huh/v2"
)

// reflectAccessor bridges reflect.Value to huh.Accessor[T].
type reflectAccessor[T any] struct {
	v reflect.Value
}

func (a *reflectAccessor[T]) Get() T {
	return a.v.Interface().(T)
}

func (a *reflectAccessor[T]) Set(val T) {
	a.v.Set(reflect.ValueOf(val))
}

// intAccessor bridges reflect.Value for int fields to huh.Accessor[string].
// It converts between int and string representation for huh.Input.
type intAccessor struct {
	v reflect.Value
}

func (a *intAccessor) Get() string {
	return fmt.Sprintf("%d", a.v.Int())
}

func (a *intAccessor) Set(val string) {
	var intVal int
	fmt.Sscanf(val, "%d", &intVal)
	a.v.SetInt(int64(intVal))
}

// buildFormForAllGroups constructs a huh.Form from all config groups.
// Uses LayoutDefault for pagination (one group per page) to handle many fields.
// Fields within each group are sized to the width of the largest field.
func buildFormForAllGroups(groups []config.GroupMeta) *huh.Form {
	huhGroups := make([]*huh.Group, 0, len(groups))
	for _, g := range groups {
		fields := make([]huh.Field, 0, len(g.Fields))
		for _, fm := range g.Fields {
			if f := buildField(fm); f != nil {
				fields = append(fields, f)
			}
		}
		if len(fields) > 0 {
			// Calculate max field width in this group and apply to all fields
			maxW := maxFieldWidth(g.Fields)
			for _, f := range fields {
				f.WithWidth(maxW)
			}
			huhGroups = append(huhGroups, huh.NewGroup(fields...))
		}
	}
	if len(huhGroups) > 0 {
		return huh.NewForm(huhGroups...).WithLayout(huh.LayoutDefault)
	}
	return huh.NewForm()
}

// buildField maps a single FieldMeta to a huh.Field.
func buildField(m config.FieldMeta) huh.Field {
	switch m.Kind {
	case config.FieldSelect:
		options := m.Options
		if m.Key == "ui.themeName" {
			options = theme.AvailableThemes()
		}
		opts := make([]huh.Option[string], len(options))
		for i, o := range options {
			opts[i] = huh.NewOption(strings.ToUpper(o[:1])+o[1:], o)
		}
		// Use inlineSelect wrapper to render label/desc on same line as options
		sel := huh.NewSelect[string]().
			Key(m.Key).
			Options(opts...).Inline(true).
			Accessor(&reflectAccessor[string]{v: m.Value})
		return newInlineSelect(m.Label, m.Desc, sel)
	case config.FieldConfirm:
		return huh.NewConfirm().
			Key(m.Key).Title(m.Label).Description(m.Desc).
			Affirmative("Yes").Negative("No").Inline(true).
			Accessor(&reflectAccessor[bool]{v: m.Value})
	case config.FieldReadOnly:
		return huh.NewNote().
			Title(m.Label + ": " + fmt.Sprint(m.Value.Interface()))
	default: // FieldInput
		// Handle different types for input fields
		switch m.Value.Kind() {
		case reflect.Int:
			return huh.NewInput().
				Key(m.Key).Title(m.Label).Description(m.Desc).Inline(true).
				Accessor(&intAccessor{v: m.Value})
		case reflect.Bool:
			return huh.NewConfirm().
				Key(m.Key).Title(m.Label).Description(m.Desc).Inline(true).
				Affirmative("Yes").Negative("No").
				Accessor(&reflectAccessor[bool]{v: m.Value})
		default: // string and others
			return huh.NewInput().
				Key(m.Key).Title(m.Label).Description(m.Desc).Inline(true).
				Accessor(&reflectAccessor[string]{v: m.Value})
		}
	}
}

// maxLabelWidth returns the longest field label length across all groups.
func maxLabelWidth(groups []config.GroupMeta) int {
	max := 0
	for _, g := range groups {
		for _, f := range g.Fields {
			if len(f.Label) > max {
				max = len(f.Label)
			}
		}
	}
	return max
}

// maxDescWidth returns the longest field description length across all groups.
func maxDescWidth(groups []config.GroupMeta) int {
	max := 0
	for _, g := range groups {
		for _, f := range g.Fields {
			if len(f.Desc) > max {
				max = len(f.Desc)
			}
		}
	}
	return max
}

// maxFieldWidth returns the width of the largest field in the group.
// This is used to align all fields to the same width for a consistent layout.
func maxFieldWidth(fields []config.FieldMeta) int {
	max := 0
	for _, f := range fields {
		w := fieldContentWidth(f)
		if w > max {
			max = w
		}
	}
	return max
}

// fieldContentWidth estimates the natural width of a field based on its content.
// Layout: "Label: description [value]" for inline fields.
func fieldContentWidth(f config.FieldMeta) int {
	const (
		labelSep    = 2 // ": " after label
		descSep     = 1 // space before/after description
		indicators  = 5 // "‹ " and " ›" for inline select
		framePad    = 4 // Base style padding
		valuePad    = 4 // Padding around value
	)

	w := len(f.Label) + labelSep

	// Add description width
	if f.Desc != "" {
		w += len(f.Desc) + descSep
	}

	// Add value width based on field type
	switch f.Kind {
	case config.FieldSelect:
		// Find the longest option
		maxOpt := 0
		for _, opt := range f.Options {
			if len(opt) > maxOpt {
				maxOpt = len(opt)
			}
		}
		w += maxOpt + indicators
	case config.FieldConfirm:
		// "Yes" or "No" buttons
		w += 6 // "Yes" + space + "No"
	case config.FieldReadOnly:
		// Display the actual value
		w += len(fmt.Sprint(f.Value.Interface()))
	default:
		// Input fields - estimate based on current value or default
		val := ""
		if f.Value.Kind() == reflect.String {
			val = f.Value.String()
		} else if f.Value.Kind() == reflect.Int {
			val = fmt.Sprintf("%d", f.Value.Int())
		}
		if len(val) == 0 {
			val = "____" // placeholder estimate
		}
		w += len(val) + valuePad
	}

	return w + framePad
}
