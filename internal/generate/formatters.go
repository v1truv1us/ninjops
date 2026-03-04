package generate

import (
	"encoding/json"
	"fmt"

	"github.com/ninjops/ninjops/internal/spec"
)

type OutputFormat string

const (
	FormatMarkdown OutputFormat = "md"
	FormatText     OutputFormat = "text"
	FormatJSON     OutputFormat = "json"
)

type Formatter struct {
	format OutputFormat
}

func NewFormatter(format OutputFormat) *Formatter {
	return &Formatter{format: format}
}

func (f *Formatter) FormatProposal(artifacts *spec.GeneratedArtifacts) (string, error) {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(artifacts)
	case FormatText:
		return f.markdownToText(artifacts.ProposalMarkdown)
	default:
		return artifacts.ProposalMarkdown, nil
	}
}

func (f *Formatter) FormatTerms(artifacts *spec.GeneratedArtifacts) (string, error) {
	switch f.format {
	case FormatJSON:
		return "", nil
	case FormatText:
		return f.markdownToText(artifacts.TermsMarkdown)
	default:
		return artifacts.TermsMarkdown, nil
	}
}

func (f *Formatter) FormatNotes(artifacts *spec.GeneratedArtifacts) (string, error) {
	return artifacts.PublicNotesText, nil
}

func (f *Formatter) FormatAll(artifacts *spec.GeneratedArtifacts) (string, error) {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(artifacts)
	case FormatText:
		proposal, _ := f.markdownToText(artifacts.ProposalMarkdown)
		terms, _ := f.markdownToText(artifacts.TermsMarkdown)
		return fmt.Sprintf("=== PROPOSAL ===\n%s\n\n=== TERMS ===\n%s\n\n=== NOTES ===\n%s",
			proposal, terms, artifacts.PublicNotesText), nil
	default:
		return fmt.Sprintf("=== PROPOSAL ===\n%s\n\n=== TERMS ===\n%s\n\n=== NOTES ===\n%s",
			artifacts.ProposalMarkdown, artifacts.TermsMarkdown, artifacts.PublicNotesText), nil
	}
}

func (f *Formatter) formatJSON(artifacts *spec.GeneratedArtifacts) (string, error) {
	data, err := json.MarshalIndent(artifacts, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to format as JSON: %w", err)
	}
	return string(data), nil
}

func (f *Formatter) markdownToText(md string) (string, error) {
	text := md

	replacements := []struct {
		from string
		to   string
	}{
		{"### ", "\n"},
		{"## ", "\n"},
		{"# ", "\n"},
		{"**", ""},
		{"*", ""},
		{"`", ""},
		{"  ", " "},
	}

	for _, r := range replacements {
		text = replaceAll(text, r.from, r.to)
	}

	return text, nil
}

func replaceAll(s, old, new string) string {
	result := ""
	for i := 0; i < len(s); {
		if i <= len(s)-len(old) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}

func ParseFormat(s string) (OutputFormat, error) {
	switch s {
	case "md", "markdown":
		return FormatMarkdown, nil
	case "text", "txt":
		return FormatText, nil
	case "json":
		return FormatJSON, nil
	default:
		return FormatMarkdown, fmt.Errorf("invalid format: %s (valid: md, text, json)", s)
	}
}
