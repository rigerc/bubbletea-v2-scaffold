package screens

import (
	"fmt"
	"reflect"
	"strings"

	"scaffold/config"
	"scaffold/internal/ui/theme"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
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

// computeAlignmentWidths returns the maximum title and description column
// widths for a group. Title width includes the ":" suffix.
func computeAlignmentWidths(group config.GroupMeta) (titleW, descW int) {
	for _, f := range group.Fields {
		if tw := lipgloss.Width(f.Label); tw > titleW {
			titleW = tw
		}
		if dw := lipgloss.Width(f.Desc); dw > descW {
			descW = dw
		}
	}
	return titleW, descW
}

// minControlWidth is the minimum width reserved for the interactive control column.
const minControlWidth = 20

// buildFormForAllGroups constructs a huh.Form from all config groups.
// Uses LayoutDefault for pagination (one group per page) to handle many fields.
// The form width is set dynamically based on the widest group's alignment needs.
func buildFormForAllGroups(groups []config.GroupMeta) *huh.Form {
	huhGroups := make([]*huh.Group, 0, len(groups))
	var maxOverhead int
	for _, g := range groups {
		titleW, descW := computeAlignmentWidths(g)
		a := fieldAlignment{titleW: titleW, descW: descW}
		if oh := a.alignmentOverhead(); oh > maxOverhead {
			maxOverhead = oh
		}
		fields := make([]huh.Field, 0, len(g.Fields))
		for _, fm := range g.Fields {
			if f := buildField(fm, titleW, descW); f != nil {
				fields = append(fields, f)
			}
		}
		if len(fields) > 0 {
			huhGroups = append(huhGroups, huh.NewGroup(fields...))
		}
	}
	if len(huhGroups) > 0 {
		formWidth := maxOverhead + minControlWidth
		return huh.NewForm(huhGroups...).
			WithLayout(huh.LayoutDefault).
			WithWidth(formWidth)
	}
	return huh.NewForm()
}

// buildField maps a single FieldMeta to a huh.Field wrapped in an aligned
// container so that title, description, and control columns align vertically
// across all fields in a group.
func buildField(m config.FieldMeta, titleW, descW int) huh.Field {
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
		sel := huh.NewSelect[string]().
			Key(m.Key).
			Options(opts...).Inline(true).
			Accessor(&reflectAccessor[string]{v: m.Value})
		return newInlineSelect(m.Label, m.Desc, titleW, descW, sel)
	case config.FieldConfirm:
		confirm := huh.NewConfirm().
			Key(m.Key).
			Affirmative("Yes").Negative("No").Inline(true).
			Accessor(&reflectAccessor[bool]{v: m.Value})
		return newAlignedField(m.Label, m.Desc, titleW, descW, confirm)
	case config.FieldReadOnly:
		note := huh.NewNote().
			Title(fmt.Sprint(m.Value.Interface()))
		return newAlignedField(m.Label, m.Desc, titleW, descW, note)
	default: // FieldInput
		switch m.Value.Kind() {
		case reflect.Int:
			input := huh.NewInput().
				Key(m.Key).Inline(true).
				Accessor(&intAccessor{v: m.Value})
			return newAlignedField(m.Label, m.Desc, titleW, descW, input)
		case reflect.Bool:
			confirm := huh.NewConfirm().
				Key(m.Key).Inline(true).
				Affirmative("Yes").Negative("No").
				Accessor(&reflectAccessor[bool]{v: m.Value})
			return newAlignedField(m.Label, m.Desc, titleW, descW, confirm)
		default: // string and others
			input := huh.NewInput().
				Key(m.Key).Inline(true).
				Accessor(&reflectAccessor[string]{v: m.Value})
			return newAlignedField(m.Label, m.Desc, titleW, descW, input)
		}
	}
}
