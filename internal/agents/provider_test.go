package agents

import (
	"context"
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
	}
	s.Project = spec.ProjectInfo{
		Name:        "Test Project",
		Description: "A test project for testing",
		Type:        "website",
	}
	s.Work = spec.WorkDefinition{
		Features: []spec.Feature{
			{Name: "Feature 1", Description: "Implement the core functionality"},
			{Name: "Feature 2", Description: "Design the user interface"},
		},
	}
	s.Pricing = spec.PricingInfo{
		Currency:  "USD",
		LineItems: []spec.LineItem{},
	}
	return s
}

func TestOfflineProvider_Clarify(t *testing.T) {
	provider := NewOfflineProvider()
	ctx := context.Background()

	req := AgentRequest{
		Role:      RoleClarify,
		Plan:      PlanDefault,
		QuoteSpec: newTestQuoteSpec(),
	}

	resp, err := provider.Execute(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp.QuoteSpec)
	assert.True(t, resp.Confidence > 0)
}

func TestOfflineProvider_Polish(t *testing.T) {
	provider := NewOfflineProvider()
	ctx := context.Background()

	req := AgentRequest{
		Role:      RolePolish,
		Plan:      PlanDefault,
		QuoteSpec: newTestQuoteSpec(),
	}

	resp, err := provider.Execute(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp.QuoteSpec)
	assert.True(t, resp.Confidence > 0)
}

func TestOfflineProvider_Boundaries(t *testing.T) {
	provider := NewOfflineProvider()
	ctx := context.Background()

	quoteSpec := newTestQuoteSpec()
	quoteSpec.Work.MinorChanges = nil
	quoteSpec.Work.OutOfScope = nil
	quoteSpec.Work.Assumptions = nil

	req := AgentRequest{
		Role:      RoleBoundaries,
		Plan:      PlanDefault,
		QuoteSpec: quoteSpec,
	}

	resp, err := provider.Execute(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp.QuoteSpec)
	assert.True(t, len(resp.QuoteSpec.Work.MinorChanges) > 0)
	assert.True(t, len(resp.QuoteSpec.Work.OutOfScope) > 0)
	assert.True(t, len(resp.QuoteSpec.Work.Assumptions) > 0)
}

func TestOfflineProvider_LineItems(t *testing.T) {
	provider := NewOfflineProvider()
	ctx := context.Background()

	quoteSpec := newTestQuoteSpec()
	quoteSpec.Pricing.LineItems = nil

	req := AgentRequest{
		Role:      RoleLineItems,
		Plan:      PlanDefault,
		QuoteSpec: quoteSpec,
	}

	resp, err := provider.Execute(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp.QuoteSpec)
	assert.True(t, len(resp.QuoteSpec.Pricing.LineItems) > 0)
}

func TestOfflineProvider_IsAvailable(t *testing.T) {
	provider := NewOfflineProvider()
	assert.True(t, provider.IsAvailable())
}

func TestIsValidRole(t *testing.T) {
	assert.True(t, IsValidRole("clarify"))
	assert.True(t, IsValidRole("polish"))
	assert.True(t, IsValidRole("boundaries"))
	assert.True(t, IsValidRole("line-items"))
	assert.False(t, IsValidRole("invalid"))
}

func TestIsValidPlan(t *testing.T) {
	assert.True(t, IsValidPlan("default"))
	assert.True(t, IsValidPlan("codex-pro"))
	assert.True(t, IsValidPlan("opencode-zen"))
	assert.True(t, IsValidPlan("zai-plan"))
	assert.False(t, IsValidPlan("invalid"))
}
