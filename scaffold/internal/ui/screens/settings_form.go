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
		return huh.NewSelect[string]().
			Key(m.Key).Title(m.Label).Description(m.Desc).
			Options(opts...).Inline(true).
			Accessor(&reflectAccessor[string]{v: m.Value})
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
