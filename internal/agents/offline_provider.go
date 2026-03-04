package agents

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/ninjops/ninjops/internal/spec"
)

type OfflineProvider struct{}

func NewOfflineProvider() *OfflineProvider {
	return &OfflineProvider{}
}

func (p *OfflineProvider) Name() string {
	return "offline"
}

func (p *OfflineProvider) IsAvailable() bool {
	return true
}

func (p *OfflineProvider) Execute(ctx context.Context, req AgentRequest) (*AgentResponse, error) {
	switch req.Role {
	case RoleClarify:
		return p.clarify(req.QuoteSpec)
	case RolePolish:
		return p.polish(req.QuoteSpec)
	case RoleBoundaries:
		return p.boundaries(req.QuoteSpec)
	case RoleLineItems:
		return p.lineItems(req.QuoteSpec)
	default:
		return nil, fmt.Errorf("unknown role: %s", req.Role)
	}
}

func (p *OfflineProvider) clarify(quoteSpec *spec.QuoteSpec) (*AgentResponse, error) {
	result := *quoteSpec
	var suggestions []string

	for i, feature := range result.Work.Features {
		normalized := p.normalizeFeatureName(feature.Name)
		if normalized != feature.Name {
			suggestions = append(suggestions, fmt.Sprintf("Normalized feature name: '%s' -> '%s'", feature.Name, normalized))
			result.Work.Features[i].Name = normalized
		}

		if feature.Priority == "" {
			result.Work.Features[i].Priority = "medium"
			suggestions = append(suggestions, fmt.Sprintf("Added default priority 'medium' to feature: %s", feature.Name))
		}

		if feature.Category == "" {
			category := p.inferCategory(feature.Name, feature.Description)
			result.Work.Features[i].Category = category
			suggestions = append(suggestions, fmt.Sprintf("Inferred category '%s' for feature: %s", category, feature.Name))
		}
	}

	responsibilities := p.extractResponsibilities(quoteSpec.Work.Features)
	if len(responsibilities) > 0 && len(result.Work.Responsibilities) == 0 {
		result.Work.Responsibilities = responsibilities
		suggestions = append(suggestions, "Extracted responsibilities from feature descriptions")
	}

	if result.Project.Timeline == "" && result.Project.Deadline != "" {
		result.Project.Timeline = "To be determined based on scope"
		suggestions = append(suggestions, "Added placeholder timeline")
	}

	return &AgentResponse{
		QuoteSpec:   &result,
		Suggestions: suggestions,
		Confidence:  0.85,
		Metadata: map[string]interface{}{
			"transformations": len(suggestions),
		},
	}, nil
}

func (p *OfflineProvider) polish(quoteSpec *spec.QuoteSpec) (*AgentResponse, error) {
	result := *quoteSpec
	var suggestions []string

	result.Project.Description = p.polishText(result.Project.Description)
	suggestions = append(suggestions, "Polished project description")

	for i := range result.Work.Features {
		result.Work.Features[i].Description = p.polishText(result.Work.Features[i].Description)
	}

	if len(result.Work.Responsibilities) > 0 {
		polishedResp := make([]string, len(result.Work.Responsibilities))
		for i, resp := range result.Work.Responsibilities {
			polishedResp[i] = p.polishText(resp)
		}
		result.Work.Responsibilities = polishedResp
		suggestions = append(suggestions, "Polished responsibilities")
	}

	if result.Client.OrgType == spec.OrgTypeChurch || result.Client.OrgType == spec.OrgTypeTaxExempt {
		suggestions = append(suggestions, "Verified tax-exempt wording for religious organization")
	} else if result.Client.OrgType == spec.OrgTypeNonprofit {
		suggestions = append(suggestions, "Verified tax-exempt wording for nonprofit organization")
	}

	return &AgentResponse{
		QuoteSpec:   &result,
		Suggestions: suggestions,
		Confidence:  0.90,
		Metadata: map[string]interface{}{
			"polished_sections": []string{"description", "features", "responsibilities"},
		},
	}, nil
}

func (p *OfflineProvider) boundaries(quoteSpec *spec.QuoteSpec) (*AgentResponse, error) {
	result := *quoteSpec
	var suggestions []string

	if len(result.Work.MinorChanges) == 0 {
		result.Work.MinorChanges = p.generateMinorChanges(quoteSpec.Project.Type)
		suggestions = append(suggestions, "Generated default minor changes list")
	}

	if len(result.Work.OutOfScope) == 0 {
		result.Work.OutOfScope = p.generateOutOfScope(quoteSpec.Project.Type)
		suggestions = append(suggestions, "Generated default out-of-scope list")
	}

	if len(result.Work.Assumptions) == 0 {
		result.Work.Assumptions = p.generateAssumptions(quoteSpec)
		suggestions = append(suggestions, "Generated default assumptions list")
	}

	return &AgentResponse{
		QuoteSpec:   &result,
		Suggestions: suggestions,
		Confidence:  0.80,
		Metadata: map[string]interface{}{
			"minor_changes_count": len(result.Work.MinorChanges),
			"out_of_scope_count":  len(result.Work.OutOfScope),
			"assumptions_count":   len(result.Work.Assumptions),
		},
	}, nil
}

func (p *OfflineProvider) lineItems(quoteSpec *spec.QuoteSpec) (*AgentResponse, error) {
	result := *quoteSpec
	var suggestions []string

	if len(result.Pricing.LineItems) == 0 {
		items := p.generateLineItems(quoteSpec.Work.Features, quoteSpec.Project.Type)
		result.Pricing.LineItems = items
		suggestions = append(suggestions, fmt.Sprintf("Generated %d line items from features", len(items)))
	}

	for i, item := range result.Pricing.LineItems {
		if item.Quantity == 0 {
			result.Pricing.LineItems[i].Quantity = 1
		}
		if item.Amount == 0 && item.Quantity > 0 && item.Rate > 0 {
			result.Pricing.LineItems[i].Amount = item.Quantity * item.Rate
		}
	}

	return &AgentResponse{
		QuoteSpec:   &result,
		Suggestions: suggestions,
		Confidence:  0.75,
		Metadata: map[string]interface{}{
			"line_items_count": len(result.Pricing.LineItems),
		},
	}, nil
}

func (p *OfflineProvider) normalizeFeatureName(name string) string {
	name = strings.TrimSpace(name)
	name = regexp.MustCompile(`\s+`).ReplaceAllString(name, " ")

	words := strings.Fields(name)
	if len(words) > 0 {
		words[0] = strings.Title(strings.ToLower(words[0]))
	}

	return strings.Join(words, " ")
}

func (p *OfflineProvider) inferCategory(name, description string) string {
	combined := strings.ToLower(name + " " + description)

	keywords := map[string]string{
		"ui": "Design", "design": "Design", "style": "Design", "css": "Design",
		"api": "Development", "backend": "Development", "database": "Development",
		"auth": "Security", "login": "Security", "permission": "Security",
		"test": "Quality Assurance", "testing": "Quality Assurance", "qa": "Quality Assurance",
		"mobile": "Mobile", "ios": "Mobile", "android": "Mobile", "app": "Mobile",
		"analytics": "Analytics", "report": "Analytics", "dashboard": "Analytics",
		"hosting": "Infrastructure", "deploy": "Infrastructure", "server": "Infrastructure",
	}

	for keyword, category := range keywords {
		if strings.Contains(combined, keyword) {
			return category
		}
	}

	return "Development"
}

func (p *OfflineProvider) extractResponsibilities(features []spec.Feature) []string {
	responsibilities := make(map[string]bool)

	for _, f := range features {
		desc := strings.ToLower(f.Description)

		if strings.Contains(desc, "implement") || strings.Contains(desc, "develop") || strings.Contains(desc, "build") {
			responsibilities["Implementation of core functionality"] = true
		}
		if strings.Contains(desc, "design") || strings.Contains(desc, "create") {
			responsibilities["Design and prototyping"] = true
		}
		if strings.Contains(desc, "test") || strings.Contains(desc, "qa") {
			responsibilities["Testing and quality assurance"] = true
		}
		if strings.Contains(desc, "deploy") || strings.Contains(desc, "launch") {
			responsibilities["Deployment and launch support"] = true
		}
	}

	result := make([]string, 0, len(responsibilities))
	for r := range responsibilities {
		result = append(result, r)
	}

	return result
}

func (p *OfflineProvider) polishText(text string) string {
	text = strings.TrimSpace(text)

	replacements := []struct {
		from string
		to   string
	}{
		{"very ", ""},
		{"really ", ""},
		{"basically ", ""},
		{"essentially ", ""},
		{"actually ", ""},
		{"  ", " "},
		{"\n\n\n", "\n\n"},
	}

	for _, r := range replacements {
		text = strings.ReplaceAll(text, r.from, r.to)
	}

	return text
}

func (p *OfflineProvider) generateMinorChanges(projectType string) []string {
	base := []string{
		"Minor text and copy changes",
		"Color scheme adjustments",
		"Layout tweaks within existing sections",
		"Small bug fixes",
	}

	switch strings.ToLower(projectType) {
	case "website", "web":
		return append(base,
			"Image placeholder updates",
			"Contact information updates",
			"Social media link updates",
		)
	case "web app", "application":
		return append(base,
			"UI element positioning adjustments",
			"Form field label changes",
			"Button text modifications",
		)
	case "mobile app", "mobile":
		return append(base,
			"Icon adjustments",
			"Screen flow minor changes",
			"Notification text updates",
		)
	default:
		return base
	}
}

func (p *OfflineProvider) generateOutOfScope(projectType string) []string {
	base := []string{
		"Major feature additions not listed in scope",
		"Third-party integrations not specified",
		"Content writing or copywriting services",
		"Stock photography or asset licensing",
		"SEO or marketing services",
	}

	switch strings.ToLower(projectType) {
	case "website", "web":
		return append(base,
			"E-commerce functionality (unless specified)",
			"Custom CMS development",
			"Server administration",
		)
	case "web app", "application":
		return append(base,
			"Native mobile app development",
			"Desktop application development",
			"Hardware integration",
		)
	default:
		return base
	}
}

func (p *OfflineProvider) generateAssumptions(quoteSpec *spec.QuoteSpec) []string {
	assumptions := []string{
		"Client will provide all necessary content and assets in a timely manner",
		"Client will provide feedback within 3 business days",
		"Project scope is as described; major changes will require a change order",
	}

	if len(quoteSpec.Project.Technologies) > 0 {
		assumptions = append(assumptions,
			"Specified technologies will be used as the primary stack",
		)
	}

	if quoteSpec.Project.Timeline != "" {
		assumptions = append(assumptions,
			"Timeline assumes no extended delays in client feedback or asset delivery",
		)
	}

	return assumptions
}

func (p *OfflineProvider) generateLineItems(features []spec.Feature, projectType string) []spec.LineItem {
	items := make([]spec.LineItem, 0, len(features))

	for _, f := range features {
		quantity := p.estimateQuantity(f)
		category := f.Category
		if category == "" {
			category = p.inferCategory(f.Name, f.Description)
		}

		items = append(items, spec.LineItem{
			Description: f.Name + " - " + f.Description,
			Quantity:    quantity,
			Rate:        0,
			Amount:      0,
			Category:    category,
		})
	}

	items = append(items, spec.LineItem{
		Description: "Project management and communication",
		Quantity:    1,
		Rate:        0,
		Amount:      0,
		Category:    "Management",
	})

	items = append(items, spec.LineItem{
		Description: "Testing and QA",
		Quantity:    1,
		Rate:        0,
		Amount:      0,
		Category:    "Quality Assurance",
	})

	return items
}

func (p *OfflineProvider) estimateQuantity(feature spec.Feature) float64 {
	return 1.0
}
