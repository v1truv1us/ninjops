package agents

import (
	"fmt"
	"strings"
)

type PromptBuilder struct {
	plan Plan
	role Role
}

func NewPromptBuilder(plan Plan, role Role) *PromptBuilder {
	return &PromptBuilder{plan: plan, role: role}
}

func (pb *PromptBuilder) BuildSystemPrompt() string {
	var systemPrompt strings.Builder

	switch pb.plan {
	case PlanCodexPro:
		systemPrompt.WriteString("You are an expert business consultant and technical writer. ")
		systemPrompt.WriteString("Your responses are precise, data-driven, and focused on measurable outcomes. ")
		systemPrompt.WriteString("You excel at creating clear, actionable proposals with detailed specifications. ")

	case PlanOpencodeZen:
		systemPrompt.WriteString("You are a thoughtful, experienced developer advocate. ")
		systemPrompt.WriteString("Your approach balances technical excellence with practical simplicity. ")
		systemPrompt.WriteString("You communicate with warmth and clarity, making complex topics accessible. ")

	case PlanZAIPlan:
		systemPrompt.WriteString("You are a strategic planning expert specializing in software projects. ")
		systemPrompt.WriteString("You focus on risk mitigation, milestone planning, and resource optimization. ")
		systemPrompt.WriteString("Your proposals include contingency plans and clear success criteria. ")

	default:
		systemPrompt.WriteString("You are a professional proposal writer and business analyst. ")
		systemPrompt.WriteString("Your responses are clear, professional, and tailored to the client's needs. ")
	}

	systemPrompt.WriteString(pb.getRoleSpecificInstructions())

	return systemPrompt.String()
}

func (pb *PromptBuilder) getRoleSpecificInstructions() string {
	switch pb.role {
	case RoleClarify:
		return `
Your task is to CLARIFY the project specification:
1. Normalize feature descriptions for consistency
2. Extract implicit requirements from descriptions
3. Identify missing information that should be clarified
4. Ensure all required sections are present and complete
5. Standardize terminology and naming conventions

Output a structured JSON with clarified features, extracted responsibilities, and suggested questions.`

	case RolePolish:
		return `
Your task is to POLISH the proposal content:
1. Improve tone consistency throughout the document
2. Tighten verbose sentences while preserving meaning
3. Ensure professional language appropriate for the client type
4. Add section headings where helpful for readability
5. Standardize formatting and structure

Output a structured JSON with polished content and improvement notes.`

	case RoleBoundaries:
		return `
Your task is to define SCOPE BOUNDARIES:
1. Generate appropriate minor_changes items based on project type
2. Create out_of_scope items to protect against scope creep
3. List reasonable assumptions for the project
4. Identify potential risks or dependencies
5. Suggest client responsibilities

Output a structured JSON with minor_changes, out_of_scope, assumptions, and client_responsibilities.`

	case RoleLineItems:
		return `
Your task is to SUGGEST LINE ITEMS:
1. Break down features into billable line items
2. Estimate reasonable quantities based on complexity
3. Group items by category (development, design, infrastructure, etc.)
4. Do NOT include pricing unless provided in the spec
5. Include brief descriptions for each line item

Output a structured JSON with suggested line items grouped by category.`

	default:
		return "Process the input and provide a structured response."
	}
}

func (pb *PromptBuilder) BuildUserPrompt(quoteSpec interface{}) string {
	return fmt.Sprintf("Process the following QuoteSpec:\n\n%+v", quoteSpec)
}

func (pb *PromptBuilder) GetOutputSchema() string {
	switch pb.role {
	case RoleClarify:
		return `{
  "features": [{"name": "string", "description": "string", "priority": "string"}],
  "responsibilities": ["string"],
  "questions": ["string"],
  "confidence": 0.0-1.0
}`

	case RolePolish:
		return `{
  "polished_sections": {"section_name": "polished_content"},
  "improvements": ["string"],
  "confidence": 0.0-1.0
}`

	case RoleBoundaries:
		return `{
  "minor_changes": ["string"],
  "out_of_scope": ["string"],
  "assumptions": ["string"],
  "client_responsibilities": ["string"],
  "confidence": 0.0-1.0
}`

	case RoleLineItems:
		return `{
  "line_items": [{"category": "string", "description": "string", "quantity": number}],
  "notes": ["string"],
  "confidence": 0.0-1.0
}`

	default:
		return `{"result": "object", "confidence": 0.0-1.0}`
	}
}

func GetPlanDescription(plan Plan) string {
	switch plan {
	case PlanCodexPro:
		return "Data-driven, precise proposals with detailed specifications"
	case PlanOpencodeZen:
		return "Balanced, warm communication making complex topics accessible"
	case PlanZAIPlan:
		return "Strategic planning with risk mitigation and milestone focus"
	default:
		return "Professional, clear proposals tailored to client needs"
	}
}

func GetRoleDescription(role Role) string {
	switch role {
	case RoleClarify:
		return "Normalize features, extract responsibilities, identify gaps"
	case RolePolish:
		return "Improve tone, tighten prose, enhance professionalism"
	case RoleBoundaries:
		return "Generate scope boundaries, assumptions, and exclusions"
	case RoleLineItems:
		return "Suggest billable line items from features"
	default:
		return "Unknown role"
	}
}
