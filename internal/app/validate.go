package app

import (
	"fmt"
	"os"

	"github.com/ninjops/ninjops/internal/spec"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	var inputFile string
	var strict bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a QuoteSpec JSON file",
		Long:  `Validates schema and required fields of a QuoteSpec JSON file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputFile == "" {
				return fmt.Errorf("--input is required")
			}

			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			quoteSpec, err := spec.ValidateJSON(data)
			if err != nil {
				if validationErrs, ok := err.(spec.ValidationErrors); ok {
					fmt.Fprintf(os.Stderr, "Validation failed:\n\n")
					for _, e := range validationErrs {
						fmt.Fprintf(os.Stderr, "  ✗ %s: %s\n", e.Field, e.Message)
					}
					return fmt.Errorf("validation failed with %d errors", len(validationErrs))
				}
				return fmt.Errorf("validation error: %w", err)
			}

			fmt.Printf("✓ Valid QuoteSpec\n")
			fmt.Printf("  Reference: %s\n", quoteSpec.Metadata.Reference)
			fmt.Printf("  Client: %s (%s)\n", quoteSpec.Client.Name, quoteSpec.Client.OrgType)
			fmt.Printf("  Project: %s\n", quoteSpec.Project.Name)
			fmt.Printf("  Features: %d\n", len(quoteSpec.Work.Features))
			fmt.Printf("  Line Items: %d\n", len(quoteSpec.Pricing.LineItems))

			if strict {
				warnings := validateStrict(quoteSpec)
				if len(warnings) > 0 {
					fmt.Printf("\nWarnings:\n")
					for _, w := range warnings {
						fmt.Printf("  ⚠ %s\n", w)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input QuoteSpec JSON file (required)")
	cmd.Flags().BoolVar(&strict, "strict", false, "Enable strict validation with warnings")

	cmd.MarkFlagRequired("input")

	return cmd
}

func validateStrict(s *spec.QuoteSpec) []string {
	var warnings []string

	if s.Project.Timeline == "" {
		warnings = append(warnings, "project.timeline is empty")
	}

	if len(s.Work.Responsibilities) == 0 {
		warnings = append(warnings, "work.responsibilities is empty")
	}

	if len(s.Work.Assumptions) == 0 {
		warnings = append(warnings, "work.assumptions is empty - consider adding project assumptions")
	}

	if s.Pricing.Total == 0 && len(s.Pricing.LineItems) > 0 {
		warnings = append(warnings, "pricing.total is 0 but line items exist - did you forget to calculate total?")
	}

	if s.Client.Phone == "" {
		warnings = append(warnings, "client.phone is empty")
	}

	return warnings
}
