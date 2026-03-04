package generate

import (
	"testing"

	"github.com/ninjops/ninjops/internal/spec"
	"github.com/stretchr/testify/assert"
)

func newTestQuoteSpec() *spec.QuoteSpec {
	s := spec.NewQuoteSpec()
	s.Client = spec.ClientInfo{
		Name:    "Test Client",
		Email:   "test@example.com",
		OrgType: spec.OrgTypeBusiness,
		Address: spec.Address{
			Line1: "123 Test St",
			City:  "Test City",
			State: "TS",
		},
	}
	s.Project = spec.ProjectInfo{
		Name:        "Test Project",
		Description: "A test project for testing",
		Type:        "website",
		Timeline:    "4 weeks",
	}
	s.Work = spec.WorkDefinition{
		Features: []spec.Feature{
			{Name: "Feature 1", Description: "First feature", Priority: "high"},
			{Name: "Feature 2", Description: "Second feature", Priority: "medium"},
		},
		Responsibilities: []string{"Implement features", "Test code"},
	}
	s.Pricing = spec.PricingInfo{
		Currency: "USD",
		LineItems: []spec.LineItem{
			{Description: "Development", Quantity: 1, Rate: 1000, Amount: 1000},
		},
		Total: 1000,
	}
	s.Settings = spec.QuoteSettings{
		Tone:           spec.ToneProfessional,
		IncludePricing: true,
	}
	return s
}

func TestGenerator_Generate(t *testing.T) {
	generator := NewGenerator()
	quoteSpec := newTestQuoteSpec()

	artifacts, err := generator.Generate(quoteSpec)
	assert.NoError(t, err)
	assert.NotNil(t, artifacts)

	assert.NotEmpty(t, artifacts.ProposalMarkdown)
	assert.NotEmpty(t, artifacts.TermsMarkdown)
	assert.NotEmpty(t, artifacts.PublicNotesText)
	assert.NotEmpty(t, artifacts.Meta.Hash)
	assert.False(t, artifacts.Meta.GeneratedAt.IsZero())
}

func TestGenerator_ProposalContainsClient(t *testing.T) {
	generator := NewGenerator()
	quoteSpec := newTestQuoteSpec()
	quoteSpec.Client.Name = "Acme Corporation"

	artifacts, err := generator.Generate(quoteSpec)
	assert.NoError(t, err)
	assert.Contains(t, artifacts.ProposalMarkdown, "Acme Corporation")
}

func TestGenerator_ProposalContainsProject(t *testing.T) {
	generator := NewGenerator()
	quoteSpec := newTestQuoteSpec()
	quoteSpec.Project.Name = "Website Redesign"

	artifacts, err := generator.Generate(quoteSpec)
	assert.NoError(t, err)
	assert.Contains(t, artifacts.ProposalMarkdown, "Website Redesign")
}

func TestGenerator_TaxExemptWording(t *testing.T) {
	generator := NewGenerator()
	quoteSpec := newTestQuoteSpec()
	quoteSpec.Client.OrgType = spec.OrgTypeChurch

	artifacts, err := generator.Generate(quoteSpec)
	assert.NoError(t, err)
	assert.Contains(t, artifacts.ProposalMarkdown, "tax-exempt religious organization")
}

func TestGenerator_NonprofitWording(t *testing.T) {
	generator := NewGenerator()
	quoteSpec := newTestQuoteSpec()
	quoteSpec.Client.OrgType = spec.OrgTypeNonprofit

	artifacts, err := generator.Generate(quoteSpec)
	assert.NoError(t, err)
	assert.Contains(t, artifacts.ProposalMarkdown, "tax-exempt nonprofit organization")
}
