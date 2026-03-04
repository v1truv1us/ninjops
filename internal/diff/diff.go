package diff

import (
	"fmt"
	"strings"
)

type DiffType string

const (
	DiffAdd    DiffType = "+"
	DiffRemove DiffType = "-"
	DiffChange DiffType = "~"
	DiffSame   DiffType = " "
)

type DiffLine struct {
	Type    DiffType
	Content string
	LineNum int
}

type DiffResult struct {
	Lines   []DiffLine
	Summary string
	HasDiff bool
}

type FieldDiff struct {
	Field    string
	OldValue string
	NewValue string
	Type     DiffType
}

func Compute(old, new string) *DiffResult {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var result DiffResult
	result.Lines = make([]DiffLine, 0)

	maxLen := len(oldLines)
	if len(newLines) > maxLen {
		maxLen = len(newLines)
	}

	oldIdx := 0
	newIdx := 0

	for i := 0; i < maxLen; i++ {
		if oldIdx < len(oldLines) && newIdx < len(newLines) {
			if oldLines[oldIdx] == newLines[newIdx] {
				result.Lines = append(result.Lines, DiffLine{
					Type:    DiffSame,
					Content: oldLines[oldIdx],
					LineNum: i + 1,
				})
				oldIdx++
				newIdx++
			} else {
				result.Lines = append(result.Lines, DiffLine{
					Type:    DiffRemove,
					Content: oldLines[oldIdx],
					LineNum: i + 1,
				})
				result.Lines = append(result.Lines, DiffLine{
					Type:    DiffAdd,
					Content: newLines[newIdx],
					LineNum: i + 1,
				})
				result.HasDiff = true
				oldIdx++
				newIdx++
			}
		} else if oldIdx < len(oldLines) {
			result.Lines = append(result.Lines, DiffLine{
				Type:    DiffRemove,
				Content: oldLines[oldIdx],
				LineNum: i + 1,
			})
			result.HasDiff = true
			oldIdx++
		} else if newIdx < len(newLines) {
			result.Lines = append(result.Lines, DiffLine{
				Type:    DiffAdd,
				Content: newLines[newIdx],
				LineNum: i + 1,
			})
			result.HasDiff = true
			newIdx++
		}
	}

	added := 0
	removed := 0
	for _, line := range result.Lines {
		if line.Type == DiffAdd {
			added++
		} else if line.Type == DiffRemove {
			removed++
		}
	}

	result.Summary = fmt.Sprintf("%d additions, %d deletions", added, removed)

	return &result
}

func ComputeFields(oldFields, newFields map[string]string) []FieldDiff {
	var diffs []FieldDiff

	allFields := make(map[string]bool)
	for k := range oldFields {
		allFields[k] = true
	}
	for k := range newFields {
		allFields[k] = true
	}

	for field := range allFields {
		oldVal, oldExists := oldFields[field]
		newVal, newExists := newFields[field]

		if !oldExists && newExists {
			diffs = append(diffs, FieldDiff{
				Field:    field,
				OldValue: "",
				NewValue: newVal,
				Type:     DiffAdd,
			})
		} else if oldExists && !newExists {
			diffs = append(diffs, FieldDiff{
				Field:    field,
				OldValue: oldVal,
				NewValue: "",
				Type:     DiffRemove,
			})
		} else if oldVal != newVal {
			diffs = append(diffs, FieldDiff{
				Field:    field,
				OldValue: oldVal,
				NewValue: newVal,
				Type:     DiffChange,
			})
		}
	}

	return diffs
}

func (r *DiffResult) String() string {
	var sb strings.Builder

	for _, line := range r.Lines {
		sb.WriteString(fmt.Sprintf("%s %s\n", line.Type, line.Content))
	}

	sb.WriteString(fmt.Sprintf("\n--- %s ---\n", r.Summary))

	return sb.String()
}

func FormatFieldDiffs(diffs []FieldDiff) string {
	if len(diffs) == 0 {
		return "No differences found"
	}

	var sb strings.Builder
	sb.WriteString("Field differences:\n\n")

	for _, d := range diffs {
		switch d.Type {
		case DiffAdd:
			sb.WriteString(fmt.Sprintf("  + %s: %s\n", d.Field, truncate(d.NewValue, 50)))
		case DiffRemove:
			sb.WriteString(fmt.Sprintf("  - %s: (was: %s)\n", d.Field, truncate(d.OldValue, 50)))
		case DiffChange:
			sb.WriteString(fmt.Sprintf("  ~ %s:\n", d.Field))
			sb.WriteString(fmt.Sprintf("      - %s\n", truncate(d.OldValue, 50)))
			sb.WriteString(fmt.Sprintf("      + %s\n", truncate(d.NewValue, 50)))
		}
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
