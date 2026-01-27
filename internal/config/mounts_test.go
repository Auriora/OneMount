package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMountsRegistry(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	// Create new registry
	registry, err := NewMountsRegistry(configDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Test adding a mount
	mountPoint := "/home/user/OneDrive"
	account := "user@example.com"
	label := "Personal OneDrive"

	if err := registry.SetMount(mountPoint, account, label); err != nil {
		t.Fatalf("Failed to set mount: %v", err)
	}

	// Test getting account
	gotAccount, exists := registry.GetAccount(mountPoint)
	if !exists {
		t.Fatal("Mount not found after setting")
	}
	if gotAccount != account {
		t.Errorf("GetAccount() = %q, want %q", gotAccount, account)
	}

	// Test getting mount config
	config, exists := registry.GetMountConfig(mountPoint)
	if !exists {
		t.Fatal("Mount config not found")
	}
	if config.Account != account {
		t.Errorf("Config.Account = %q, want %q", config.Account, account)
	}
	if config.Label != label {
		t.Errorf("Config.Label = %q, want %q", config.Label, label)
	}

	// Test listing mounts
	mounts := registry.ListMounts()
	if len(mounts) != 1 {
		t.Errorf("ListMounts() returned %d mounts, want 1", len(mounts))
	}

	// Test persistence - create new registry from same path
	registry2, err := NewMountsRegistry(configDir)
	if err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	gotAccount2, exists := registry2.GetAccount(mountPoint)
	if !exists {
		t.Fatal("Mount not found after reload")
	}
	if gotAccount2 != account {
		t.Errorf("After reload: GetAccount() = %q, want %q", gotAccount2, account)
	}

	// Test removing mount
	if err := registry.RemoveMount(mountPoint); err != nil {
		t.Fatalf("Failed to remove mount: %v", err)
	}

	_, exists = registry.GetAccount(mountPoint)
	if exists {
		t.Error("Mount still exists after removal")
	}
}

func TestMountsRegistry_MultipleAccounts(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	registry, err := NewMountsRegistry(configDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Add multiple mounts for different accounts
	mounts := map[string]string{
		"/home/user/OneDrive": "user@example.com",
		"/home/user/Work":     "user@company.com",
		"/mnt/shared":         "user@example.com",
	}

	for mountPoint, account := range mounts {
		if err := registry.SetMount(mountPoint, account, ""); err != nil {
			t.Fatalf("Failed to set mount %s: %v", mountPoint, err)
		}
	}

	// Test getting mounts by account
	personalMounts := registry.GetMountsByAccount("user@example.com")
	if len(personalMounts) != 2 {
		t.Errorf("GetMountsByAccount(user@example.com) returned %d mounts, want 2", len(personalMounts))
	}

	workMounts := registry.GetMountsByAccount("user@company.com")
	if len(workMounts) != 1 {
		t.Errorf("GetMountsByAccount(user@company.com) returned %d mounts, want 1", len(workMounts))
	}
}

func TestMountsRegistry_EmptyConfigDir(t *testing.T) {
	// Test with empty config dir (should use default)
	registry, err := NewMountsRegistry("")
	if err != nil {
		t.Fatalf("Failed to create registry with empty config dir: %v", err)
	}

	// Should create default path
	if registry.path == "" {
		t.Error("Registry path is empty")
	}

	// Clean up - remove the created file
	os.Remove(registry.path)
}

func TestMountsRegistry_UpdateMount(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	registry, err := NewMountsRegistry(configDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	mountPoint := "/home/user/OneDrive"

	// Set initial mount
	if err := registry.SetMount(mountPoint, "user@example.com", "Personal"); err != nil {
		t.Fatalf("Failed to set mount: %v", err)
	}

	config1, _ := registry.GetMountConfig(mountPoint)
	created := config1.Created

	// Update mount with different account
	if err := registry.SetMount(mountPoint, "user@newaccount.com", "Updated"); err != nil {
		t.Fatalf("Failed to update mount: %v", err)
	}

	config2, exists := registry.GetMountConfig(mountPoint)
	if !exists {
		t.Fatal("Mount not found after update")
	}

	if config2.Account != "user@newaccount.com" {
		t.Errorf("Account not updated: got %q, want %q", config2.Account, "user@newaccount.com")
	}

	if config2.Label != "Updated" {
		t.Errorf("Label not updated: got %q, want %q", config2.Label, "Updated")
	}

	if config2.Created != created {
		t.Error("Created timestamp should not change on update")
	}

	if config2.Updated.IsZero() {
		t.Error("Updated timestamp should be set")
	}
}
