package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ninjops/ninjops/internal/spec"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize ninjops in the current directory",
		Long:  `Creates .ninjops/ directory with sample config and QuoteSpec template.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if outputDir == "" {
				wd, _ := os.Getwd()
				outputDir = wd
			}

			ninjopsDir := filepath.Join(outputDir, ".ninjops")
			if err := os.MkdirAll(ninjopsDir, 0750); err != nil {
				return fmt.Errorf("failed to create .ninjops directory: %w", err)
			}

			configContent := `# Ninjops Configuration
[ninja]
base_url = "https://invoiceninja.fergify.work"
# api_token = ""  # Set via NINJOPS_NINJA_API_TOKEN env var
# api_secret = ""  # Set via NINJOPS_NINJA_API_SECRET env var

[agent]
provider = "offline"  # offline, openai, or anthropic
plan = "default"  # default, codex-pro, opencode-zen, zai-plan

[serve]
listen = "127.0.0.1"
port = 8080
`
			configPath := filepath.Join(ninjopsDir, "config.toml")
			if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			sampleSpec := spec.NewQuoteSpec()
			sampleSpec.Client = spec.ClientInfo{
				Name:    "Acme Corporation",
				Email:   "contact@acme.com",
				Phone:   "555-123-4567",
				OrgType: spec.OrgTypeBusiness,
				Address: spec.Address{
					Line1:      "123 Business St",
					City:       "Business City",
					State:      "CA",
					PostalCode: "90210",
					Country:    "US",
				},
			}
			sampleSpec.Project = spec.ProjectInfo{
				Name:         "Corporate Website Redesign",
				Description:  "Complete redesign of corporate website with modern UI/UX",
				Type:         "website",
				Timeline:     "6-8 weeks",
				Technologies: []string{"React", "Node.js", "PostgreSQL"},
			}
			sampleSpec.Work = spec.WorkDefinition{
				Features: []spec.Feature{
					{
						Name:        "Homepage Redesign",
						Description: "Modern, responsive homepage with hero section and key features showcase",
						Priority:    "high",
						Category:    "Design",
					},
					{
						Name:        "About Page",
						Description: "Company overview, team section, and mission statement",
						Priority:    "medium",
						Category:    "Design",
					},
					{
						Name:        "Contact Form",
						Description: "Contact form with validation and email integration",
						Priority:    "high",
						Category:    "Development",
					},
				},
				Responsibilities: []string{
					"Design and implementation of all pages",
					"Responsive design for all device sizes",
					"SEO optimization",
				},
			}
			sampleSpec.Pricing = spec.PricingInfo{
				Currency: "USD",
				LineItems: []spec.LineItem{
					{Description: "Homepage Design & Development", Quantity: 1, Rate: 2500, Amount: 2500, Category: "Design"},
					{Description: "About Page", Quantity: 1, Rate: 1200, Amount: 1200, Category: "Design"},
					{Description: "Contact Form", Quantity: 1, Rate: 800, Amount: 800, Category: "Development"},
				},
				Total:        4500,
				PaymentTerms: "Net 15",
			}
			sampleSpec.Settings = spec.QuoteSettings{
				Tone:            spec.ToneProfessional,
				IncludePricing:  true,
				IncludeTimeline: true,
			}

			specJSON, _ := sampleSpec.ToJSON()
			specPath := filepath.Join(outputDir, "quote_template.json")
			if err := os.WriteFile(specPath, specJSON, 0600); err != nil {
				return fmt.Errorf("failed to write sample spec: %w", err)
			}

			fmt.Printf("✓ Created %s/\n", ninjopsDir)
			fmt.Printf("✓ Created %s\n", configPath)
			fmt.Printf("✓ Created %s\n", specPath)
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Edit quote_template.json with your project details")
			fmt.Println("  2. Run: ninjops validate --input quote_template.json")
			fmt.Println("  3. Run: ninjops generate --input quote_template.json --format md")

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory (default: current directory)")

	return cmd
}
