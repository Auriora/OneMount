package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/auriora/onemount/internal/logging"
)

// MountConfig represents the configuration for a single mount point
type MountConfig struct {
	Account string    `json:"account"`           // OneDrive account email
	Created time.Time `json:"created"`           // When this mount was created
	Label   string    `json:"label,omitempty"`   // Optional user-friendly label
	Updated time.Time `json:"updated,omitempty"` // Last updated timestamp
}

// MountsRegistry manages the mapping between mount points and OneDrive accounts
type MountsRegistry struct {
	Mounts map[string]*MountConfig `json:"mounts"`
	mu     sync.RWMutex
	path   string
}

// NewMountsRegistry creates or loads the mounts registry from the config directory
func NewMountsRegistry(configDir string) (*MountsRegistry, error) {
	if configDir == "" {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			return nil, err
		}
		configDir = filepath.Join(userConfigDir, "onemount")
	}

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, err
	}

	registryPath := filepath.Join(configDir, "mounts.json")
	registry := &MountsRegistry{
		Mounts: make(map[string]*MountConfig),
		path:   registryPath,
	}

	// Try to load existing registry
	if _, err := os.Stat(registryPath); err == nil {
		if err := registry.load(); err != nil {
			logging.Warn().
				Err(err).
				Str("path", registryPath).
				Msg("Failed to load mounts registry, starting with empty registry")
		}
	}

	return registry, nil
}

// load reads the registry from disk
func (r *MountsRegistry) load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	data, err := os.ReadFile(r.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, r)
}

// save writes the registry to disk
// NOTE: Caller must hold the lock (either read or write)
func (r *MountsRegistry) save() error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(r.path, data, 0600)
}

// GetAccount returns the account associated with a mount point
func (r *MountsRegistry) GetAccount(mountPoint string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, exists := r.Mounts[mountPoint]
	if !exists {
		return "", false
	}
	return config.Account, true
}

// SetMount registers or updates a mount point with its associated account
func (r *MountsRegistry) SetMount(mountPoint, account, label string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if existing, exists := r.Mounts[mountPoint]; exists {
		// Update existing mount
		existing.Account = account
		existing.Updated = now
		if label != "" {
			existing.Label = label
		}
	} else {
		// Create new mount
		r.Mounts[mountPoint] = &MountConfig{
			Account: account,
			Created: now,
			Label:   label,
		}
	}

	return r.save()
}

// RemoveMount removes a mount point from the registry
func (r *MountsRegistry) RemoveMount(mountPoint string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.Mounts, mountPoint)
	return r.save()
}

// ListMounts returns all registered mount points
func (r *MountsRegistry) ListMounts() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mounts := make([]string, 0, len(r.Mounts))
	for mountPoint := range r.Mounts {
		mounts = append(mounts, mountPoint)
	}
	return mounts
}

// GetMountConfig returns the full configuration for a mount point
func (r *MountsRegistry) GetMountConfig(mountPoint string) (*MountConfig, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	config, exists := r.Mounts[mountPoint]
	return config, exists
}

// GetMountsByAccount returns all mount points for a given account
func (r *MountsRegistry) GetMountsByAccount(account string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mounts := make([]string, 0)
	for mountPoint, config := range r.Mounts {
		if config.Account == account {
			mounts = append(mounts, mountPoint)
		}
	}
	return mounts
}
