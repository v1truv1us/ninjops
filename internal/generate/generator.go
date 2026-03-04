package generate

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"text/template"
	"time"

	"github.com/ninjops/ninjops/internal/spec"
	"github.com/ninjops/ninjops/internal/templates"
)

type Generator struct {
	templateVersion string
	funcMap         template.FuncMap
}

func NewGenerator() *Generator {
	return &Generator{
		templateVersion: "1.0.0",
		funcMap: template.FuncMap{
			"formatDate": func(layout string, t time.Time) string {
				return t.Format(layout)
			},
			"lower": func(s string) string {
				return fmt.Sprintf("%s", s)
			},
		},
	}
}

func (g *Generator) Generate(quoteSpec *spec.QuoteSpec) (*spec.GeneratedArtifacts, error) {
	data := &TemplateData{
		QuoteSpec:      quoteSpec,
		OrgTypeWording: quoteSpec.GetOrgTypeWording(),
	}

	proposalMD, err := g.renderTemplate("proposal.md.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate proposal: %w", err)
	}

	termsMD, err := g.renderTemplate("terms.md.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate terms: %w", err)
	}

	notesText, err := g.renderTemplate("notes.txt.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to generate notes: %w", err)
	}

	hash := computeHash(proposalMD, termsMD, notesText)

	return &spec.GeneratedArtifacts{
		ProposalMarkdown: proposalMD,
		TermsMarkdown:    termsMD,
		PublicNotesText:  notesText,
		Meta: spec.GenMeta{
			GeneratedAt: time.Now().UTC(),
			TemplateVer: g.templateVersion,
			Hash:        hash,
		},
	}, nil
}

func (g *Generator) renderTemplate(name string, data *TemplateData) (string, error) {
	tmplBytes, err := templates.ReadTemplate(name)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Funcs(g.funcMap).Parse(string(tmplBytes))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", name, err)
	}

	return buf.String(), nil
}

type TemplateData struct {
	*spec.QuoteSpec
	OrgTypeWording string
}

func computeHash(proposal, terms, notes string) string {
	h := sha256.New()
	h.Write([]byte(proposal))
	h.Write([]byte(terms))
	h.Write([]byte(notes))
	return hex.EncodeToString(h.Sum(nil))[:16]
}
