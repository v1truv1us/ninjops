package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ninjops/ninjops/internal/spec"
)

const (
	StateDir  = ".ninjops"
	StateFile = "state.json"
)

type StateEntry struct {
	ReferenceID  string    `json:"reference_id"`
	ClientID     string    `json:"client_id,omitempty"`
	ProjectID    string    `json:"project_id,omitempty"`
	QuoteID      string    `json:"quote_id,omitempty"`
	InvoiceID    string    `json:"invoice_id,omitempty"`
	LastSyncHash string    `json:"last_sync_hash,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type State struct {
	Version string                `json:"version"`
	Entries map[string]StateEntry `json:"entries"`
}

type Store struct {
	mu    sync.RWMutex
	path  string
	state *State
}

func NewStore(projectDir string) (*Store, error) {
	stateDir := filepath.Join(projectDir, StateDir)
	if err := os.MkdirAll(stateDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	statePath := filepath.Join(stateDir, StateFile)
	s := &Store{
		path: statePath,
		state: &State{
			Version: "1.0.0",
			Entries: make(map[string]StateEntry),
		},
	}

	if err := s.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	return s, nil
}

func (s *Store) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s.state)
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	return os.WriteFile(s.path, data, 0600)
}

func (s *Store) GetEntry(referenceID string) (*StateEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.state.Entries[referenceID]
	if !exists {
		return nil, nil
	}

	return &entry, nil
}

func (s *Store) SetEntry(entry StateEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, exists := s.state.Entries[entry.ReferenceID]
	if exists {
		entry.CreatedAt = existing.CreatedAt
	} else {
		entry.CreatedAt = time.Now().UTC()
	}
	entry.UpdatedAt = time.Now().UTC()

	s.state.Entries[entry.ReferenceID] = entry
	return s.save()
}

func (s *Store) UpdateClientID(referenceID, clientID string) error {
	return s.updateField(referenceID, func(e *StateEntry) {
		e.ClientID = clientID
	})
}

func (s *Store) UpdateProjectID(referenceID, projectID string) error {
	return s.updateField(referenceID, func(e *StateEntry) {
		e.ProjectID = projectID
	})
}

func (s *Store) UpdateQuoteID(referenceID, quoteID string) error {
	return s.updateField(referenceID, func(e *StateEntry) {
		e.QuoteID = quoteID
	})
}

func (s *Store) UpdateInvoiceID(referenceID, invoiceID string) error {
	return s.updateField(referenceID, func(e *StateEntry) {
		e.InvoiceID = invoiceID
	})
}

func (s *Store) UpdateSyncHash(referenceID, hash string) error {
	return s.updateField(referenceID, func(e *StateEntry) {
		e.LastSyncHash = hash
	})
}

func (s *Store) updateField(referenceID string, update func(*StateEntry)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.state.Entries[referenceID]
	if !exists {
		entry = StateEntry{
			ReferenceID: referenceID,
			CreatedAt:   time.Now().UTC(),
		}
	}

	update(&entry)
	entry.UpdatedAt = time.Now().UTC()
	s.state.Entries[referenceID] = entry

	return s.save()
}

func (s *Store) DeleteEntry(referenceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.state.Entries, referenceID)
	return s.save()
}

func (s *Store) ListEntries() ([]StateEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]StateEntry, 0, len(s.state.Entries))
	for _, entry := range s.state.Entries {
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *Store) FindByClientID(clientID string) ([]StateEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matches []StateEntry
	for _, entry := range s.state.Entries {
		if entry.ClientID == clientID {
			matches = append(matches, entry)
		}
	}

	return matches, nil
}

func (s *Store) FindByQuoteID(quoteID string) (*StateEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, entry := range s.state.Entries {
		if entry.QuoteID == quoteID {
			return &entry, nil
		}
	}

	return nil, nil
}

func ComputeHash(spec *spec.QuoteSpec) string {
	data, _ := json.Marshal(spec)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
