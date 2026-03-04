package spec

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	SchemaVersion   = "1.0.0"
	ReferencePrefix = "ninjops:"
)

type OrgType string

const (
	OrgTypeBusiness  OrgType = "business"
	OrgTypeChurch    OrgType = "church"
	OrgTypeNonprofit OrgType = "nonprofit"
	OrgTypeTaxExempt OrgType = "tax_exempt"
)

type Tone string

const (
	ToneProfessional Tone = "professional"
	ToneFormal       Tone = "formal"
	ToneFriendly     Tone = "friendly"
)

type QuoteSpec struct {
	SchemaVersion string         `json:"schema_version" yaml:"schema_version"`
	Metadata      Metadata       `json:"metadata" yaml:"metadata"`
	Client        ClientInfo     `json:"client" yaml:"client"`
	Project       ProjectInfo    `json:"project" yaml:"project"`
	TaskIDs       []string       `json:"task_ids,omitempty" yaml:"task_ids,omitempty"`
	Work          WorkDefinition `json:"work" yaml:"work"`
	Pricing       PricingInfo    `json:"pricing" yaml:"pricing"`
	Settings      QuoteSettings  `json:"settings" yaml:"settings"`
}

type Metadata struct {
	Reference   string    `json:"reference" yaml:"reference"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
	GeneratedBy string    `json:"generated_by,omitempty" yaml:"generated_by,omitempty"`
}

type ClientInfo struct {
	ID      string  `json:"id,omitempty" yaml:"id,omitempty"`
	Name    string  `json:"name" yaml:"name"`
	Email   string  `json:"email" yaml:"email"`
	Phone   string  `json:"phone,omitempty" yaml:"phone,omitempty"`
	Address Address `json:"address,omitempty" yaml:"address,omitempty"`
	OrgType OrgType `json:"org_type" yaml:"org_type"`
}

type Address struct {
	Line1      string `json:"line1,omitempty" yaml:"line1,omitempty"`
	Line2      string `json:"line2,omitempty" yaml:"line2,omitempty"`
	City       string `json:"city,omitempty" yaml:"city,omitempty"`
	State      string `json:"state,omitempty" yaml:"state,omitempty"`
	PostalCode string `json:"postal_code,omitempty" yaml:"postal_code,omitempty"`
	Country    string `json:"country,omitempty" yaml:"country,omitempty"`
}

type ProjectInfo struct {
	ID           string   `json:"id,omitempty" yaml:"id,omitempty"`
	Name         string   `json:"name" yaml:"name"`
	Description  string   `json:"description" yaml:"description"`
	Type         string   `json:"type" yaml:"type"`
	Timeline     string   `json:"timeline,omitempty" yaml:"timeline,omitempty"`
	Deadline     string   `json:"deadline,omitempty" yaml:"deadline,omitempty"`
	Technologies []string `json:"technologies,omitempty" yaml:"technologies,omitempty"`
}

type WorkDefinition struct {
	Features         []Feature `json:"features" yaml:"features"`
	Responsibilities []string  `json:"responsibilities,omitempty" yaml:"responsibilities,omitempty"`
	MinorChanges     []string  `json:"minor_changes,omitempty" yaml:"minor_changes,omitempty"`
	OutOfScope       []string  `json:"out_of_scope,omitempty" yaml:"out_of_scope,omitempty"`
	Assumptions      []string  `json:"assumptions,omitempty" yaml:"assumptions,omitempty"`
}

type Feature struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Priority    string   `json:"priority,omitempty" yaml:"priority,omitempty"`
	Category    string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tasks       []string `json:"tasks,omitempty" yaml:"tasks,omitempty"`
}

type PricingInfo struct {
	Total        float64       `json:"total,omitempty" yaml:"total,omitempty"`
	Currency     string        `json:"currency" yaml:"currency"`
	LineItems    []LineItem    `json:"line_items" yaml:"line_items"`
	Discount     *Discount     `json:"discount,omitempty" yaml:"discount,omitempty"`
	PaymentTerms string        `json:"payment_terms,omitempty" yaml:"payment_terms,omitempty"`
	Recurring    *RecurringFee `json:"recurring,omitempty" yaml:"recurring,omitempty"`
	Deposit      *Deposit      `json:"deposit,omitempty" yaml:"deposit,omitempty"`
}

type LineItem struct {
	Description string  `json:"description" yaml:"description"`
	Quantity    float64 `json:"quantity" yaml:"quantity"`
	Rate        float64 `json:"rate" yaml:"rate"`
	Amount      float64 `json:"amount" yaml:"amount"`
	Category    string  `json:"category,omitempty" yaml:"category,omitempty"`
}

type Discount struct {
	Type        string  `json:"type" yaml:"type"`
	Percentage  float64 `json:"percentage,omitempty" yaml:"percentage,omitempty"`
	Amount      float64 `json:"amount,omitempty" yaml:"amount,omitempty"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
}

type RecurringFee struct {
	Type        string  `json:"type" yaml:"type"`
	Amount      float64 `json:"amount" yaml:"amount"`
	Frequency   string  `json:"frequency" yaml:"frequency"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
}

type Deposit struct {
	Percentage float64 `json:"percentage,omitempty" yaml:"percentage,omitempty"`
	Amount     float64 `json:"amount,omitempty" yaml:"amount,omitempty"`
	DueDate    string  `json:"due_date,omitempty" yaml:"due_date,omitempty"`
}

type QuoteSettings struct {
	Tone            Tone   `json:"tone" yaml:"tone"`
	IncludeTimeline bool   `json:"include_timeline,omitempty" yaml:"include_timeline,omitempty"`
	IncludePricing  bool   `json:"include_pricing,omitempty" yaml:"include_pricing,omitempty"`
	Template        string `json:"template,omitempty" yaml:"template,omitempty"`
	Locale          string `json:"locale,omitempty" yaml:"locale,omitempty"`
}

type GeneratedArtifacts struct {
	ProposalMarkdown string  `json:"proposal_markdown" yaml:"proposal_markdown"`
	TermsMarkdown    string  `json:"terms_markdown" yaml:"terms_markdown"`
	PublicNotesText  string  `json:"public_notes_text" yaml:"public_notes_text"`
	Meta             GenMeta `json:"meta" yaml:"meta"`
}

type GenMeta struct {
	GeneratedAt time.Time `json:"generated_at" yaml:"generated_at"`
	TemplateVer string    `json:"template_version" yaml:"template_version"`
	Hash        string    `json:"hash" yaml:"hash"`
}

func NewQuoteSpec() *QuoteSpec {
	now := time.Now().UTC()
	return &QuoteSpec{
		SchemaVersion: SchemaVersion,
		Metadata: Metadata{
			Reference: ReferencePrefix + uuid.New().String(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Client: ClientInfo{
			OrgType: OrgTypeBusiness,
		},
		Project: ProjectInfo{},
		Work: WorkDefinition{
			Features:         []Feature{},
			Responsibilities: []string{},
			MinorChanges:     []string{},
			OutOfScope:       []string{},
			Assumptions:      []string{},
		},
		Pricing: PricingInfo{
			Currency:  "USD",
			LineItems: []LineItem{},
		},
		Settings: QuoteSettings{
			Tone:           ToneProfessional,
			IncludePricing: true,
		},
	}
}

func (q *QuoteSpec) UpdateTimestamp() {
	q.Metadata.UpdatedAt = time.Now().UTC()
}

func (q *QuoteSpec) CalculateTotal() float64 {
	var total float64
	for _, item := range q.Pricing.LineItems {
		total += item.Amount
	}
	if q.Pricing.Discount != nil {
		if q.Pricing.Discount.Percentage > 0 {
			total = total * (1 - q.Pricing.Discount.Percentage/100)
		} else if q.Pricing.Discount.Amount > 0 {
			total = total - q.Pricing.Discount.Amount
		}
	}
	return total
}

func (q *QuoteSpec) GetOrgTypeWording() string {
	switch q.Client.OrgType {
	case OrgTypeChurch, OrgTypeTaxExempt:
		return "tax-exempt religious organization"
	case OrgTypeNonprofit:
		return "tax-exempt nonprofit organization"
	default:
		return "business"
	}
}

func (q *QuoteSpec) ToJSON() ([]byte, error) {
	return json.MarshalIndent(q, "", "  ")
}

func FromJSON(data []byte) (*QuoteSpec, error) {
	var spec QuoteSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse QuoteSpec: %w", err)
	}
	return &spec, nil
}

func (a *GeneratedArtifacts) ToJSON() ([]byte, error) {
	return json.MarshalIndent(a, "", "  ")
}
