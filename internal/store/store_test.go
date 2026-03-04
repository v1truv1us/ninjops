package store

import (
	"os"
	"testing"

	"github.com/ninjops/ninjops/internal/spec"
	"github.com/stretchr/testify/assert"
)

func TestNewStore(t *testing.T) {
	dir, err := os.MkdirTemp("", "ninjops-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	store, err := NewStore(dir)
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestStore_SetAndGetEntry(t *testing.T) {
	dir, err := os.MkdirTemp("", "ninjops-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	store, err := NewStore(dir)
	assert.NoError(t, err)

	entry := StateEntry{
		ReferenceID: "test-ref-123",
		ClientID:    "client-456",
		QuoteID:     "quote-789",
	}

	err = store.SetEntry(entry)
	assert.NoError(t, err)

	retrieved, err := store.GetEntry("test-ref-123")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "client-456", retrieved.ClientID)
	assert.Equal(t, "quote-789", retrieved.QuoteID)
}

func TestStore_UpdateFields(t *testing.T) {
	dir, err := os.MkdirTemp("", "ninjops-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	store, err := NewStore(dir)
	assert.NoError(t, err)

	entry := StateEntry{
		ReferenceID: "test-ref-123",
		ClientID:    "client-456",
	}
	_ = store.SetEntry(entry)

	err = store.UpdateQuoteID("test-ref-123", "quote-new")
	assert.NoError(t, err)

	retrieved, _ := store.GetEntry("test-ref-123")
	assert.Equal(t, "quote-new", retrieved.QuoteID)
	assert.Equal(t, "client-456", retrieved.ClientID)
}

func TestStore_DeleteEntry(t *testing.T) {
	dir, err := os.MkdirTemp("", "ninjops-test")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	store, err := NewStore(dir)
	assert.NoError(t, err)

	entry := StateEntry{
		ReferenceID: "test-ref-123",
		ClientID:    "client-456",
	}
	_ = store.SetEntry(entry)

	err = store.DeleteEntry("test-ref-123")
	assert.NoError(t, err)

	retrieved, _ := store.GetEntry("test-ref-123")
	assert.Nil(t, retrieved)
}

func TestComputeHash(t *testing.T) {
	spec1 := spec.NewQuoteSpec()
	spec1.Client.Name = "Client A"

	spec2 := spec.NewQuoteSpec()
	spec2.Client.Name = "Client B"

	hash1 := ComputeHash(spec1)
	hash2 := ComputeHash(spec2)

	assert.NotEmpty(t, hash1)
	assert.NotEmpty(t, hash2)
	assert.NotEqual(t, hash1, hash2)
}
