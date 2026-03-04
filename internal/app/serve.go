package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ninjops/ninjops/internal/agents"
	"github.com/ninjops/ninjops/internal/config"
	"github.com/ninjops/ninjops/internal/generate"
	"github.com/ninjops/ninjops/internal/invoiceninja"
	"github.com/ninjops/ninjops/internal/spec"
	"github.com/ninjops/ninjops/internal/store"
	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var listen string
	var port int
	defaults := activeConfig()

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start local HTTP API server",
		Long:  `Starts a local HTTP server with REST endpoints for generate, assist, and sync operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr := fmt.Sprintf("%s:%d", listen, port)

			http.HandleFunc("/generate", handleGenerate)
			http.HandleFunc("/assist/", handleAssist)
			http.HandleFunc("/ninja/sync", handleNinjaSync)
			http.HandleFunc("/health", handleHealth)

			fmt.Printf("Starting server on %s\n", addr)
			fmt.Println("Endpoints:")
			fmt.Println("  POST /generate     - Generate artifacts from QuoteSpec")
			fmt.Println("  POST /assist/{role} - Get AI assistance")
			fmt.Println("  POST /ninja/sync   - Sync with Invoice Ninja")
			fmt.Println("  GET  /health       - Health check")

			return http.ListenAndServe(addr, nil)
		},
	}

	cmd.Flags().StringVar(&listen, "listen", defaults.Serve.Listen, "Address to listen on")
	cmd.Flags().IntVarP(&port, "port", "p", defaults.Serve.Port, "Port to listen on")

	return cmd
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var quoteSpec spec.QuoteSpec
	if err := json.NewDecoder(r.Body).Decode(&quoteSpec); err != nil {
		http.Error(w, fmt.Sprintf("Invalid QuoteSpec: %v", err), http.StatusBadRequest)
		return
	}

	if err := spec.Validate(&quoteSpec); err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	generator := generate.NewGenerator()
	artifacts, err := generator.Generate(&quoteSpec)
	if err != nil {
		http.Error(w, fmt.Sprintf("Generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(artifacts)
}

func handleAssist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role := r.URL.Path[len("/assist/"):]
	if !agents.IsValidRole(role) {
		http.Error(w, fmt.Sprintf("Invalid role: %s", role), http.StatusBadRequest)
		return
	}

	var req struct {
		QuoteSpec *spec.QuoteSpec `json:"quote_spec"`
		Provider  string          `json:"provider,omitempty"`
		Plan      string          `json:"plan,omitempty"`
		Model     string          `json:"model,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.QuoteSpec == nil {
		http.Error(w, "Missing quote_spec", http.StatusBadRequest)
		return
	}

	appCfg := activeConfig()

	providerName := req.Provider
	if providerName == "" {
		providerName = appCfg.Agent.Provider
	}

	planName := agents.Plan(req.Plan)
	if planName == "" {
		planName = agents.Plan(appCfg.Agent.Plan)
	}

	modelName := req.Model
	if modelName == "" {
		modelName = appCfg.Agent.Model
	}
	providerName, modelName = config.ResolveProviderModel(providerName, modelName)

	apiKey := config.ResolveProviderAPIKey(providerName, appCfg.Agent.ProviderAPIKey)
	provider := agents.GetProvider(providerName, apiKey)

	if !provider.IsAvailable() {
		http.Error(w, fmt.Sprintf("Provider %s not available", providerName), http.StatusServiceUnavailable)
		return
	}

	agentReq := agents.AgentRequest{
		Role:      agents.Role(role),
		Plan:      planName,
		Model:     modelName,
		QuoteSpec: req.QuoteSpec,
	}

	ctx := context.Background()
	response, err := provider.Execute(ctx, agentReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Agent execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleNinjaSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	appCfg := activeConfig()

	if appCfg.Ninja.APIToken == "" {
		http.Error(w, "Invoice Ninja API token not configured", http.StatusServiceUnavailable)
		return
	}

	var req struct {
		QuoteSpec *spec.QuoteSpec `json:"quote_spec"`
		Mode      string          `json:"mode"`
		DryRun    bool            `json:"dry_run"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	if req.QuoteSpec == nil {
		http.Error(w, "Missing quote_spec", http.StatusBadRequest)
		return
	}

	if err := spec.Validate(req.QuoteSpec); err != nil {
		http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
		return
	}

	generator := generate.NewGenerator()
	artifacts, err := generator.Generate(req.QuoteSpec)
	if err != nil {
		http.Error(w, fmt.Sprintf("Generation failed: %v", err), http.StatusInternalServerError)
		return
	}

	client := invoiceninja.NewClient(appCfg.Ninja)
	st, err := store.NewStore(".")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to initialize state store: %v", err), http.StatusInternalServerError)
		return
	}
	syncer := invoiceninja.NewSyncer(client, st)

	syncMode := invoiceninja.SyncMode(req.Mode)
	if syncMode == "" {
		syncMode = invoiceninja.SyncModeQuote
	}

	opts := invoiceninja.SyncOptions{
		Mode:   syncMode,
		DryRun: req.DryRun,
	}

	ctx := context.Background()
	result, err := syncer.Sync(ctx, req.QuoteSpec, artifacts, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Sync failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func activeConfig() *config.Config {
	if cfg != nil {
		return cfg
	}
	return config.DefaultConfig()
}
