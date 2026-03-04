package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate_ValidSpec(t *testing.T) {
	spec := NewQuoteSpec()
	spec.Client.Name = "Test Client"
	spec.Client.Email = "test@example.com"
	spec.Project.Name = "Test Project"
	spec.Project.Description = "Test description"
	spec.Project.Type = "website"
	spec.Work.Features = []Feature{
		{Name: "Feature 1", Description: "Description 1"},
	}

	err := Validate(spec)
	assert.NoError(t, err)
}

func TestValidate_MissingRequired(t *testing.T) {
	spec := &QuoteSpec{}

	err := Validate(spec)
	assert.Error(t, err)

	validationErrs, ok := err.(ValidationErrors)
	assert.True(t, ok)
	assert.True(t, len(validationErrs) > 0)
}

func TestValidate_NilSpec(t *testing.T) {
	err := Validate(nil)
	assert.EqualError(t, err, "quote spec is required")
}

func TestValidate_InvalidEmail(t *testing.T) {
	spec := NewQuoteSpec()
	spec.Client.Name = "Test Client"
	spec.Client.Email = "invalid-email"
	spec.Project.Name = "Test Project"
	spec.Project.Description = "Test description"
	spec.Project.Type = "website"
	spec.Work.Features = []Feature{
		{Name: "Feature 1", Description: "Description 1"},
	}

	err := Validate(spec)
	assert.Error(t, err)
}

func TestValidate_InvalidOrgType(t *testing.T) {
	spec := NewQuoteSpec()
	spec.Client.Name = "Test Client"
	spec.Client.Email = "test@example.com"
	spec.Client.OrgType = OrgType("invalid")
	spec.Project.Name = "Test Project"
	spec.Project.Description = "Test description"
	spec.Project.Type = "website"
	spec.Work.Features = []Feature{
		{Name: "Feature 1", Description: "Description 1"},
	}

	err := Validate(spec)
	assert.Error(t, err)
}

func TestValidate_NegativeQuantity(t *testing.T) {
	spec := NewQuoteSpec()
	spec.Client.Name = "Test Client"
	spec.Client.Email = "test@example.com"
	spec.Project.Name = "Test Project"
	spec.Project.Description = "Test description"
	spec.Project.Type = "website"
	spec.Work.Features = []Feature{
		{Name: "Feature 1", Description: "Description 1"},
	}
	spec.Pricing.LineItems = []LineItem{
		{Description: "Item 1", Quantity: -1, Rate: 100, Amount: -100},
	}

	err := Validate(spec)
	assert.Error(t, err)
}

func TestExtractReferenceID(t *testing.T) {
	id, err := ExtractReferenceID("ninjops:550e8400-e29b-41d4-a716-446655440000")
	assert.NoError(t, err)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", id)
}

func TestExtractReferenceID_Invalid(t *testing.T) {
	_, err := ExtractReferenceID("invalid-reference")
	assert.Error(t, err)
}
