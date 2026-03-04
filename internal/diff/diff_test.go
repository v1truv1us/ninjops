package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompute(t *testing.T) {
	old := "line1\nline2\nline3"
	new := "line1\nline2 modified\nline3\nline4"

	result := Compute(old, new)

	assert.True(t, result.HasDiff)
	assert.Contains(t, result.Summary, "additions")
	assert.Contains(t, result.Summary, "deletions")
}

func TestCompute_NoDiff(t *testing.T) {
	text := "line1\nline2\nline3"

	result := Compute(text, text)

	assert.False(t, result.HasDiff)
}

func TestComputeFields(t *testing.T) {
	oldFields := map[string]string{
		"field1": "value1",
		"field2": "value2",
	}

	newFields := map[string]string{
		"field1": "value1 modified",
		"field2": "value2",
		"field3": "value3",
	}

	diffs := ComputeFields(oldFields, newFields)

	assert.True(t, len(diffs) > 0)
}

func TestFormatFieldDiffs(t *testing.T) {
	diffs := []FieldDiff{
		{Field: "field1", OldValue: "old", NewValue: "new", Type: DiffChange},
		{Field: "field2", OldValue: "", NewValue: "added", Type: DiffAdd},
		{Field: "field3", OldValue: "removed", NewValue: "", Type: DiffRemove},
	}

	output := FormatFieldDiffs(diffs)
	assert.Contains(t, output, "field1")
	assert.Contains(t, output, "field2")
	assert.Contains(t, output, "field3")
}

func TestFormatFieldDiffs_Empty(t *testing.T) {
	output := FormatFieldDiffs([]FieldDiff{})
	assert.Contains(t, output, "No differences found")
}
