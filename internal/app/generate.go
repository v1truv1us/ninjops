package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ninjops/ninjops/internal/generate"
	"github.com/ninjops/ninjops/internal/spec"
	"github.com/spf13/cobra"
)

func newGenerateCmd() *cobra.Command {
	var inputFile string
	var format string
	var outDir string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate proposal, terms, and notes from QuoteSpec",
		Long:  `Generates proposal markdown, terms markdown, and public notes text from a QuoteSpec JSON file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputFile == "" {
				return fmt.Errorf("--input is required")
			}

			data, err := os.ReadFile(inputFile)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			quoteSpec, err := spec.FromJSON(data)
			if err != nil {
				return fmt.Errorf("failed to parse QuoteSpec: %w", err)
			}

			if err := spec.Validate(quoteSpec); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}

			generator := generate.NewGenerator()
			artifacts, err := generator.Generate(quoteSpec)
			if err != nil {
				return fmt.Errorf("failed to generate: %w", err)
			}

			formatter := generate.NewFormatter(generate.OutputFormat(format))

			if outDir != "" {
				if err := os.MkdirAll(outDir, 0755); err != nil {
					return fmt.Errorf("failed to create output directory: %w", err)
				}

				proposal, _ := formatter.FormatProposal(artifacts)
				if err := os.WriteFile(filepath.Join(outDir, "proposal.md"), []byte(proposal), 0644); err != nil {
					return fmt.Errorf("failed to write proposal: %w", err)
				}

				terms, _ := formatter.FormatTerms(artifacts)
				if err := os.WriteFile(filepath.Join(outDir, "terms.md"), []byte(terms), 0644); err != nil {
					return fmt.Errorf("failed to write terms: %w", err)
				}

				notes, _ := formatter.FormatNotes(artifacts)
				if err := os.WriteFile(filepath.Join(outDir, "notes.txt"), []byte(notes), 0644); err != nil {
					return fmt.Errorf("failed to write notes: %w", err)
				}

				generatedJSON, _ := json.MarshalIndent(artifacts, "", "  ")
				if err := os.WriteFile(filepath.Join(outDir, "generated.json"), generatedJSON, 0644); err != nil {
					return fmt.Errorf("failed to write generated.json: %w", err)
				}

				fmt.Printf("✓ Generated artifacts in %s/\n", outDir)
				fmt.Printf("  proposal.md\n")
				fmt.Printf("  terms.md\n")
				fmt.Printf("  notes.txt\n")
				fmt.Printf("  generated.json\n")
				fmt.Printf("\nHash: %s\n", artifacts.Meta.Hash)

			} else {
				output, err := formatter.FormatAll(artifacts)
				if err != nil {
					return fmt.Errorf("failed to format output: %w", err)
				}
				fmt.Println(output)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input QuoteSpec JSON file (required)")
	cmd.Flags().StringVarP(&format, "format", "f", "md", "Output format: md, text, json")
	cmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "Output directory for generated files")

	cmd.MarkFlagRequired("input")

	return cmd
}
