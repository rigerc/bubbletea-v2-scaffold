package config

import (
	"reflect"
	"slices"
	"strings"
)

// FieldKind classifies how a config field should be rendered in the UI.
type FieldKind uint8

const (
	FieldInput    FieldKind = iota // string              → text input
	FieldSelect                    // string + cfg_options → select dropdown
	FieldConfirm                   // bool                → confirm toggle
	FieldReadOnly                  // cfg_readonly:"true"  → read-only note
)

// FieldMeta holds UI metadata for a single config field.
type FieldMeta struct {
	Key      string // dot-path koanf key, e.g. "ui.themeName"
	Label    string // cfg_label tag, falls back to struct field name
	Desc     string // cfg_desc tag
	Kind     FieldKind
	Options  []string // non-nil only for FieldSelect
	ReadOnly bool
	Value    reflect.Value // settable Value pointing into the working *Config
}

// GroupMeta groups related fields under a label.
type GroupMeta struct {
	Label  string
	Fields []FieldMeta
}

// Schema reflects over cfg and returns ordered groups of field metadata.
// cfg MUST be a pointer so reflect.Values are settable.
func Schema(cfg *Config) []GroupMeta {
	rv := reflect.ValueOf(cfg).Elem()
	rt := rv.Type()

	var groups []GroupMeta
	var topFields []FieldMeta

	for i := range rt.NumField() {
		sf := rt.Field(i)
		fv := rv.Field(i)
		koanfKey := sf.Tag.Get("koanf")
		if koanfKey == "" {
			continue
		}
		if sf.Tag.Get("cfg_exclude") == "true" {
			continue
		}
		if fv.Kind() == reflect.Struct {
			groups = append(groups, GroupMeta{
				Label:  tagOrName(sf, "cfg_label"),
				Fields: nestedFields(fv, koanfKey),
			})
		} else {
			topFields = append(topFields, leafField(sf, fv, koanfKey))
		}
	}

	if len(topFields) > 0 {
		groups = slices.Insert(groups, 0, GroupMeta{
			Label:  "General",
			Fields: topFields,
		})
	}
	return groups
}

// FieldCount returns the total number of interactive (non-readonly) fields.
func FieldCount(groups []GroupMeta) int {
	n := 0
	for _, g := range groups {
		for _, f := range g.Fields {
			if !f.ReadOnly {
				n++
			}
		}
	}
	return n
}

func nestedFields(rv reflect.Value, prefix string) []FieldMeta {
	rt := rv.Type()
	fields := make([]FieldMeta, 0, rt.NumField())
	for i := range rt.NumField() {
		sf := rt.Field(i)
		fv := rv.Field(i)
		key := sf.Tag.Get("koanf")
		if key == "" {
			continue
		}
		fields = append(fields, leafField(sf, fv, prefix+"."+key))
	}
	return fields
}

func leafField(sf reflect.StructField, fv reflect.Value, key string) FieldMeta {
	readOnly := sf.Tag.Get("cfg_readonly") == "true"
	options := parseOptions(sf.Tag.Get("cfg_options"))
	return FieldMeta{
		Key:      key,
		Label:    tagOrName(sf, "cfg_label"),
		Desc:     sf.Tag.Get("cfg_desc"),
		ReadOnly: readOnly,
		Options:  options,
		Kind:     deriveKind(fv.Kind(), options, readOnly),
		Value:    fv,
	}
}

func deriveKind(k reflect.Kind, options []string, readOnly bool) FieldKind {
	if readOnly {
		return FieldReadOnly
	}
	switch k {
	case reflect.Bool:
		return FieldConfirm
	case reflect.String:
		if options != nil {
			return FieldSelect
		}
		return FieldInput
	default:
		return FieldInput
	}
}

func tagOrName(sf reflect.StructField, tag string) string {
	if v := sf.Tag.Get(tag); v != "" {
		return v
	}
	return sf.Name
}

func parseOptions(s string) []string {
	if s == "" {
		return nil
	}
	if s == "_themes" {
		// Non-nil but empty: signals FieldSelect; options injected by form builder.
		return []string{}
	}
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
