package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/spec"
)

func TestLoadLookupSpec_WithRefOnly(t *testing.T) {
	lookup, err := loadLookupSpec("", "ninjops:550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("loadLookupSpec returned error: %v", err)
	}
	if lookup.Metadata.Reference != "ninjops:550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected reference: %s", lookup.Metadata.Reference)
	}
}

func TestLoadLookupSpec_OverrideReferenceFromFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "quote.json")

	quoteSpec := spec.NewQuoteSpec()
	quoteSpec.Client.Name = "Test Client"
	quoteSpec.Client.Email = "test@example.com"
	quoteSpec.Project.Name = "Project"
	quoteSpec.Project.Description = "Description"
	quoteSpec.Project.Type = "web"
	quoteSpec.Work.Features = []spec.Feature{{Name: "Feature", Description: "Description"}}

	data, err := quoteSpec.ToJSON()
	if err != nil {
		t.Fatalf("quoteSpec.ToJSON returned error: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("os.WriteFile returned error: %v", err)
	}

	lookup, err := loadLookupSpec(path, "ninjops:550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("loadLookupSpec returned error: %v", err)
	}
	if lookup.Metadata.Reference != "ninjops:550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected reference: %s", lookup.Metadata.Reference)
	}
}

func TestNinjaPullCmd_RejectsConflictingEntityIDs(t *testing.T) {
	prevCfg := cfg
	cfg = &config.Config{Ninja: config.NinjaConfig{APIToken: "token"}}
	t.Cleanup(func() {
		cfg = prevCfg
	})

	cmd := newNinjaPullCmd()
	cmd.SetArgs([]string{"--quote-id", "q1", "--invoice-id", "i1"})

	err := cmd.Execute()
	if err == nil || err.Error() != "--quote-id and --invoice-id cannot be used together" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNinjaDiffCmd_RejectsConflictingEntityIDs(t *testing.T) {
	prevCfg := cfg
	cfg = &config.Config{Ninja: config.NinjaConfig{APIToken: "token"}}
	t.Cleanup(func() {
		cfg = prevCfg
	})

	cmd := newNinjaDiffCmd()
	cmd.SetArgs([]string{"--input", "quote.json", "--quote-id", "q1", "--invoice-id", "i1"})

	err := cmd.Execute()
	if err == nil || err.Error() != "--quote-id and --invoice-id cannot be used together" {
		t.Fatalf("unexpected error: %v", err)
	}
}
