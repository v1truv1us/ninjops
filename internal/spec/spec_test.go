package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewQuoteSpec(t *testing.T) {
	spec := NewQuoteSpec()

	assert.NotEmpty(t, spec.SchemaVersion)
	assert.NotEmpty(t, spec.Metadata.Reference)
	assert.Contains(t, spec.Metadata.Reference, ReferencePrefix)
	assert.False(t, spec.Metadata.CreatedAt.IsZero())
	assert.False(t, spec.Metadata.UpdatedAt.IsZero())
	assert.Equal(t, OrgTypeBusiness, spec.Client.OrgType)
	assert.Equal(t, "USD", spec.Pricing.Currency)
	assert.Equal(t, ToneProfessional, spec.Settings.Tone)
}

func TestQuoteSpec_GetOrgTypeWording(t *testing.T) {
	tests := []struct {
		orgType  OrgType
		expected string
	}{
		{OrgTypeBusiness, "business"},
		{OrgTypeChurch, "tax-exempt religious organization"},
		{OrgTypeNonprofit, "tax-exempt nonprofit organization"},
		{OrgTypeTaxExempt, "tax-exempt religious organization"},
	}

	for _, tt := range tests {
		t.Run(string(tt.orgType), func(t *testing.T) {
			spec := &QuoteSpec{Client: ClientInfo{OrgType: tt.orgType}}
			assert.Equal(t, tt.expected, spec.GetOrgTypeWording())
		})
	}
}

func TestQuoteSpec_CalculateTotal(t *testing.T) {
	spec := &QuoteSpec{
		Pricing: PricingInfo{
			LineItems: []LineItem{
				{Amount: 1000},
				{Amount: 500},
				{Amount: 250},
			},
		},
	}

	assert.Equal(t, 1750.0, spec.CalculateTotal())
}

func TestQuoteSpec_CalculateTotalWithDiscount(t *testing.T) {
	spec := &QuoteSpec{
		Pricing: PricingInfo{
			LineItems: []LineItem{
				{Amount: 1000},
				{Amount: 500},
			},
			Discount: &Discount{
				Percentage: 10,
			},
		},
	}

	assert.Equal(t, 1350.0, spec.CalculateTotal())
}

func TestQuoteSpec_ToJSON(t *testing.T) {
	spec := NewQuoteSpec()
	spec.Client.Name = "Test Client"

	json, err := spec.ToJSON()
	assert.NoError(t, err)
	assert.Contains(t, string(json), "Test Client")
	assert.Contains(t, string(json), "schema_version")
}

func TestFromJSON(t *testing.T) {
	jsonData := `{
		"schema_version": "1.0.0",
		"metadata": {
			"reference": "ninjops:550e8400-e29b-41d4-a716-446655440000",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		},
		"client": {
			"name": "Test Client",
			"email": "test@example.com",
			"org_type": "business"
		},
		"project": {
			"name": "Test Project",
			"description": "Test description",
			"type": "website"
		},
		"work": {
			"features": [
				{"name": "Feature 1", "description": "Description 1"}
			]
		},
		"pricing": {
			"currency": "USD",
			"line_items": []
		},
		"settings": {
			"tone": "professional"
		}
	}`

	spec, err := FromJSON([]byte(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, "Test Client", spec.Client.Name)
	assert.Equal(t, "Test Project", spec.Project.Name)
}

func TestFromJSON_WithOptionalLinkageIDs(t *testing.T) {
	jsonData := `{
		"schema_version": "1.0.0",
		"metadata": {
			"reference": "ninjops:550e8400-e29b-41d4-a716-446655440000",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		},
		"client": {
			"id": "client-1",
			"name": "Test Client",
			"email": "test@example.com",
			"org_type": "business"
		},
		"project": {
			"id": "project-1",
			"name": "Test Project",
			"description": "Test description",
			"type": "website"
		},
		"task_ids": ["task-1", "task-2"],
		"work": {
			"features": [
				{"name": "Feature 1", "description": "Description 1"}
			]
		},
		"pricing": {
			"currency": "USD",
			"line_items": []
		},
		"settings": {
			"tone": "professional"
		}
	}`

	parsed, err := FromJSON([]byte(jsonData))
	assert.NoError(t, err)
	assert.Equal(t, "client-1", parsed.Client.ID)
	assert.Equal(t, "project-1", parsed.Project.ID)
	assert.Equal(t, []string{"task-1", "task-2"}, parsed.TaskIDs)
}
