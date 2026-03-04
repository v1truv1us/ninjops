package app

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/invoiceninja"
)

func TestConvertCmd_MissingTokenReturnsClearError(t *testing.T) {
	prevCfg := cfg
	cfg = &config.Config{Ninja: config.NinjaConfig{}}
	t.Cleanup(func() {
		cfg = prevCfg
	})

	cmd := newConvertCmd()
	cmd.SetArgs([]string{"q1"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "Invoice Ninja API token not configured"; got != want {
		t.Fatalf("unexpected error %q, want %q", got, want)
	}
}

func TestRunConvertWorkflow_CancelsWithoutConverting(t *testing.T) {
	stub := &stubQuoteConverter{
		quote: &invoiceninja.NinjaQuote{ID: "q1", Number: "Q-100", ClientID: "c1", Amount: 1500},
	}

	in := bytes.NewBufferString("n\n")
	out := &bytes.Buffer{}

	err := runConvertWorkflow(context.Background(), in, out, stub, "q1", false)
	if err != nil {
		t.Fatalf("runConvertWorkflow returned error: %v", err)
	}
	if stub.convertCalls != 0 {
		t.Fatalf("expected convert not to be called, got %d", stub.convertCalls)
	}
	if !strings.Contains(out.String(), "Cancelled") {
		t.Fatalf("expected output to include cancellation message, got %q", out.String())
	}
}

func TestParseEditFieldSelection_ValidAndInvalidValues(t *testing.T) {
	selection, err := parseEditFieldSelection("public_notes")
	if err != nil {
		t.Fatalf("parseEditFieldSelection returned error: %v", err)
	}
	if !selection.publicNotes || selection.terms {
		t.Fatalf("unexpected selection for public_notes: %+v", selection)
	}

	selection, err = parseEditFieldSelection("both")
	if err != nil {
		t.Fatalf("parseEditFieldSelection returned error: %v", err)
	}
	if !selection.publicNotes || !selection.terms {
		t.Fatalf("unexpected selection for both: %+v", selection)
	}

	_, err = parseEditFieldSelection("unknown")
	if err == nil {
		t.Fatal("expected error for unsupported field")
	}
	if got, want := err.Error(), "unsupported --field \"unknown\""; !strings.Contains(got, want) {
		t.Fatalf("unexpected error %q, want substring %q", got, want)
	}
}

func TestPromptMultilineReplacement_UsesSentinelAndKeepsCurrentOnEmpty(t *testing.T) {
	out := &bytes.Buffer{}
	reader := bytes.NewBufferString("Line one\nLine two\n.end\n")

	updated, err := promptMultilineReplacement(bufio.NewReader(reader), out, "public_notes", "current")
	if err != nil {
		t.Fatalf("promptMultilineReplacement returned error: %v", err)
	}
	if got, want := updated, "Line one\nLine two"; got != want {
		t.Fatalf("unexpected updated text %q, want %q", got, want)
	}

	out.Reset()
	reader = bytes.NewBufferString(".end\n")
	updated, err = promptMultilineReplacement(bufio.NewReader(reader), out, "terms", "keep me")
	if err != nil {
		t.Fatalf("promptMultilineReplacement returned error: %v", err)
	}
	if got, want := updated, "keep me"; got != want {
		t.Fatalf("unexpected keep-current behavior %q, want %q", got, want)
	}
}

func TestValidateNewInvoiceMode_RejectsConflicts(t *testing.T) {
	err := validateNewInvoiceMode("q1", "quote.json", "", "", "", "", "")
	if err == nil {
		t.Fatal("expected conflict error")
	}
	if got, want := err.Error(), "cannot be used together"; !strings.Contains(got, want) {
		t.Fatalf("unexpected error %q, want substring %q", got, want)
	}

	err = validateNewInvoiceMode("", "", "client1", "", "", "", "task1")
	if err != nil {
		t.Fatalf("unexpected error when using entity selectors: %v", err)
	}

	err = validateNewInvoiceMode("", "", "", "", "", "", "")
	if err != nil {
		t.Fatalf("unexpected error when no flags provided (interactive mode): %v", err)
	}
}

func TestNewInvoiceCmd_RejectsConflictingModesBeforeTokenCheck(t *testing.T) {
	prevCfg := cfg
	cfg = &config.Config{Ninja: config.NinjaConfig{}}
	t.Cleanup(func() {
		cfg = prevCfg
	})

	cmd := newNewInvoiceCmd()
	cmd.SetArgs([]string{"--from-quote", "q1", "--input", "quote.json"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if got, want := err.Error(), "--from-quote, --input, and entity selection flags cannot be used together"; got != want {
		t.Fatalf("unexpected error %q, want %q", got, want)
	}
}

type stubQuoteConverter struct {
	quote        *invoiceninja.NinjaQuote
	convertErr   error
	convertCalls int
}

func (s *stubQuoteConverter) GetQuote(context.Context, string) (*invoiceninja.NinjaQuote, error) {
	if s.quote == nil {
		return nil, errors.New("missing quote")
	}
	return s.quote, nil
}

func (s *stubQuoteConverter) ConvertQuoteToInvoice(context.Context, string) (*invoiceninja.NinjaInvoice, error) {
	s.convertCalls++
	if s.convertErr != nil {
		return nil, s.convertErr
	}
	return &invoiceninja.NinjaInvoice{ID: "i1", Number: "I-100", ClientID: "c1", Amount: 1500}, nil
}
