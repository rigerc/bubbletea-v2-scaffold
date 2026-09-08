package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchema_ExcludedFieldsAbsent verifies that fields tagged cfg_exclude:"true"
// are not present in any group returned by Schema.
func TestSchema_ExcludedFieldsAbsent(t *testing.T) {
	cfg := DefaultConfig()
	groups := Schema(cfg)

	// ConfigVersion and App are cfg_exclude:"true" on the top-level Config struct.
	// Neither their key names nor their group labels should appear.
	for _, g := range groups {
		assert.NotEqual(t, "Application", g.Label, "excluded App group must not appear")
		for _, f := range g.Fields {
			assert.NotEqual(t, "configVersion", f.Key, "excluded configVersion must not appear")
			assert.NotContains(t, f.Key, "app.", "fields under excluded App must not appear")
		}
	}
}

// TestSchema_ThemesOptionNonNilEmptySlice verifies that cfg_options:"_themes"
// produces a non-nil empty slice (signals FieldSelect with dynamic population).
func TestSchema_ThemesOptionNonNilEmptySlice(t *testing.T) {
	cfg := DefaultConfig()
	groups := Schema(cfg)

	var themeField *FieldMeta
	for i := range groups {
		for j := range groups[i].Fields {
			if groups[i].Fields[j].Key == "ui.themeName" {
				themeField = &groups[i].Fields[j]
				break
			}
		}
	}
	require.NotNil(t, themeField, "ui.themeName field must be present")
	assert.NotNil(t, themeField.Options, "Options must be non-nil (signals FieldSelect)")
	assert.Empty(t, themeField.Options, "Options must be empty (populated dynamically by form builder)")
	assert.Equal(t, FieldSelect, themeField.Kind)
}

// TestSchema_NestedStructBecomesGroup verifies that struct fields (UIConfig)
// are promoted to their own GroupMeta rather than appearing as leaf fields.
func TestSchema_NestedStructBecomesGroup(t *testing.T) {
	cfg := DefaultConfig()
	groups := Schema(cfg)

	var uiGroup *GroupMeta
	for i := range groups {
		if groups[i].Label == "UI Settings" {
			uiGroup = &groups[i]
			break
		}
	}
	require.NotNil(t, uiGroup, "UIConfig must produce a 'UI Settings' group")
	assert.NotEmpty(t, uiGroup.Fields, "UI Settings group must have fields")

	// All fields in the group should have the "ui." prefix
	for _, f := range uiGroup.Fields {
		assert.Contains(t, f.Key, "ui.", "UI group fields must have ui. prefix")
	}
}

// TestFieldCount_ExcludesReadOnly verifies that FieldCount counts only
// interactive (non-readonly) fields.
func TestFieldCount_ExcludesReadOnly(t *testing.T) {
	// Count total fields vs readonly fields
	cfg := DefaultConfig()
	groups := Schema(cfg)

	total := 0
	readonlyCount := 0
	for _, g := range groups {
		for _, f := range g.Fields {
			total++
			if f.ReadOnly {
				readonlyCount++
			}
		}
	}

	count := FieldCount(groups)
	assert.Equal(t, total-readonlyCount, count,
		"FieldCount must equal total fields minus readonly fields")
}

// TestSchema_TopLevelNonStructFieldsGroupedAsGeneral verifies that top-level
// non-struct fields (LogLevel, Debug) end up in the "General" group.
func TestSchema_TopLevelNonStructFieldsGroupedAsGeneral(t *testing.T) {
	cfg := DefaultConfig()
	groups := Schema(cfg)

	require.NotEmpty(t, groups)
	assert.Equal(t, "General", groups[0].Label, "first group must be General")

	keys := make(map[string]bool)
	for _, f := range groups[0].Fields {
		keys[f.Key] = true
	}
	assert.True(t, keys["logLevel"], "logLevel must be in General group")
	assert.True(t, keys["debug"], "debug must be in General group")
}
