package metadata

import "testing"

func TestEntryValidateDefaults(t *testing.T) {
	nowEntry := &Entry{
		ID:    "123",
		Name:  "file.txt",
		State: ItemStateHydrated,
	}
	if err := nowEntry.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if nowEntry.ItemType != ItemKindUnknown {
		t.Fatalf("expected default item type UNKNOWN, got %s", nowEntry.ItemType)
	}
	if nowEntry.OverlayPolicy != OverlayPolicyRemoteWins {
		t.Fatalf("expected default overlay REMOTE_WINS, got %s", nowEntry.OverlayPolicy)
	}
	if nowEntry.Pin.Mode != PinModeUnset {
		t.Fatalf("expected default pin mode UNSET, got %s", nowEntry.Pin.Mode)
	}
	if nowEntry.CreatedAt.IsZero() || nowEntry.UpdatedAt.IsZero() {
		t.Fatalf("expected timestamps populated")
	}
}

func TestEntryValidateRejectsBadState(t *testing.T) {
	entry := &Entry{
		ID:    "123",
		Name:  "file.txt",
		State: ItemState("INVALID"),
	}
	if err := entry.Validate(); err == nil {
		t.Fatalf("expected error for invalid state")
	}
}

func TestEntryValidateRejectsOverlayPolicy(t *testing.T) {
	entry := &Entry{
		ID:            "123",
		Name:          "file.txt",
		State:         ItemStateHydrated,
		OverlayPolicy: OverlayPolicy("BAD"),
	}
	if err := entry.Validate(); err == nil {
		t.Fatalf("expected overlay policy validation error")
	}
}
