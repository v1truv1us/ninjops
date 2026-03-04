package spec

import (
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, fmt.Sprintf("%s: %s", e.Field, e.Message))
	}
	return strings.Join(msgs, "; ")
}

func (ve ValidationErrors) ToJSON() string {
	data, _ := json.MarshalIndent(ve, "", "  ")
	return string(data)
}

type Validator struct {
	errors ValidationErrors
}

func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

func (v *Validator) Validate(spec *QuoteSpec) ValidationErrors {
	v.errors = make(ValidationErrors, 0)

	v.validateMetadata(&spec.Metadata)
	v.validateClient(&spec.Client)
	v.validateProject(&spec.Project)
	v.validateWork(&spec.Work)
	v.validatePricing(&spec.Pricing)
	v.validateSettings(&spec.Settings)

	return v.errors
}

func (v *Validator) validateMetadata(m *Metadata) {
	if m.Reference == "" {
		v.addError("metadata.reference", "reference is required")
	} else if !strings.HasPrefix(m.Reference, ReferencePrefix) {
		v.addError("metadata.reference", fmt.Sprintf("reference must start with %s", ReferencePrefix))
	}

	if m.CreatedAt.IsZero() {
		v.addError("metadata.created_at", "created_at is required")
	}

	if m.UpdatedAt.IsZero() {
		v.addError("metadata.updated_at", "updated_at is required")
	}
}

func (v *Validator) validateClient(c *ClientInfo) {
	if c.Name == "" {
		v.addError("client.name", "client name is required")
	}

	if c.Email == "" {
		v.addError("client.email", "client email is required")
	} else if _, err := mail.ParseAddress(c.Email); err != nil {
		v.addError("client.email", "invalid email format")
	}

	validOrgTypes := map[OrgType]bool{
		OrgTypeBusiness:  true,
		OrgTypeChurch:    true,
		OrgTypeNonprofit: true,
		OrgTypeTaxExempt: true,
	}

	if !validOrgTypes[c.OrgType] {
		v.addError("client.org_type", fmt.Sprintf("invalid org_type: %s (valid: business, church, nonprofit, tax_exempt)", c.OrgType))
	}
}

func (v *Validator) validateProject(p *ProjectInfo) {
	if p.Name == "" {
		v.addError("project.name", "project name is required")
	}

	if p.Description == "" {
		v.addError("project.description", "project description is required")
	}

	if p.Type == "" {
		v.addError("project.type", "project type is required")
	}
}

func (v *Validator) validateWork(w *WorkDefinition) {
	if len(w.Features) == 0 {
		v.addError("work.features", "at least one feature is required")
	}

	for i, f := range w.Features {
		if f.Name == "" {
			v.addError(fmt.Sprintf("work.features[%d].name", i), "feature name is required")
		}
		if f.Description == "" {
			v.addError(fmt.Sprintf("work.features[%d].description", i), "feature description is required")
		}
	}
}

func (v *Validator) validatePricing(p *PricingInfo) {
	if p.Currency == "" {
		v.addError("pricing.currency", "currency is required")
	}

	for i, item := range p.LineItems {
		if item.Description == "" {
			v.addError(fmt.Sprintf("pricing.line_items[%d].description", i), "line item description is required")
		}
		if item.Quantity < 0 {
			v.addError(fmt.Sprintf("pricing.line_items[%d].quantity", i), "quantity cannot be negative")
		}
		if item.Rate < 0 {
			v.addError(fmt.Sprintf("pricing.line_items[%d].rate", i), "rate cannot be negative")
		}
	}

	if p.Discount != nil {
		if p.Discount.Percentage < 0 || p.Discount.Percentage > 100 {
			v.addError("pricing.discount.percentage", "discount percentage must be between 0 and 100")
		}
		if p.Discount.Amount < 0 {
			v.addError("pricing.discount.amount", "discount amount cannot be negative")
		}
	}

	if p.Recurring != nil {
		if p.Recurring.Amount < 0 {
			v.addError("pricing.recurring.amount", "recurring amount cannot be negative")
		}
		validFrequencies := map[string]bool{
			"monthly":   true,
			"quarterly": true,
			"annually":  true,
		}
		if !validFrequencies[p.Recurring.Frequency] && p.Recurring.Frequency != "" {
			v.addError("pricing.recurring.frequency", "invalid frequency (valid: monthly, quarterly, annually)")
		}
	}

	if p.Deposit != nil {
		if p.Deposit.Percentage < 0 || p.Deposit.Percentage > 100 {
			v.addError("pricing.deposit.percentage", "deposit percentage must be between 0 and 100")
		}
		if p.Deposit.Amount < 0 {
			v.addError("pricing.deposit.amount", "deposit amount cannot be negative")
		}
	}
}

func (v *Validator) validateSettings(s *QuoteSettings) {
	validTones := map[Tone]bool{
		ToneProfessional: true,
		ToneFormal:       true,
		ToneFriendly:     true,
	}

	if !validTones[s.Tone] {
		v.addError("settings.tone", fmt.Sprintf("invalid tone: %s (valid: professional, formal, friendly)", s.Tone))
	}
}

func (v *Validator) addError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

func Validate(spec *QuoteSpec) error {
	if spec == nil {
		return fmt.Errorf("quote spec is required")
	}

	v := NewValidator()
	errors := v.Validate(spec)
	if len(errors) > 0 {
		return errors
	}
	return nil
}

func ValidateJSON(data []byte) (*QuoteSpec, error) {
	spec, err := FromJSON(data)
	if err != nil {
		return nil, fmt.Errorf("JSON parse error: %w", err)
	}

	if err := Validate(spec); err != nil {
		return nil, err
	}

	return spec, nil
}

func IsValidURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func IsValidUUID(str string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(str))
}

func ExtractReferenceID(ref string) (string, error) {
	if !strings.HasPrefix(ref, ReferencePrefix) {
		return "", fmt.Errorf("invalid reference format: must start with %s", ReferencePrefix)
	}
	id := strings.TrimPrefix(ref, ReferencePrefix)
	if !IsValidUUID(id) {
		return "", fmt.Errorf("invalid UUID in reference: %s", id)
	}
	return id, nil
}
