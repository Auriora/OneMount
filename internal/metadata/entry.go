package metadata

import (
	"fmt"
	"time"
)

// OperationError captures context about the last failure for hydration/upload.
type OperationError struct {
	Message    string    `json:"message"`
	Temporary  bool      `json:"temporary,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

// HydrationState records information about the most recent hydration attempt.
type HydrationState struct {
	WorkerID    string          `json:"worker_id,omitempty"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	Error       *OperationError `json:"error,omitempty"`
}

// UploadState records context for uploads in flight or pending retries.
type UploadState struct {
	SessionID   string          `json:"session_id,omitempty"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	LastError   *OperationError `json:"last_error,omitempty"`
}

// PinState indicates the policy governing hydration/eviction decisions.
type PinState struct {
	Mode   PinMode    `json:"mode"`
	Policy string     `json:"policy,omitempty"`
	Since  *time.Time `json:"since,omitempty"`
}

// Entry is the canonical record persisted to BBolt for every filesystem item.
type Entry struct {
	ID            string            `json:"id"`
	RemoteID      string            `json:"remote_id,omitempty"`
	ParentID      string            `json:"parent_id,omitempty"`
	Name          string            `json:"name"`
	ItemType      ItemKind          `json:"item_type"`
	State         ItemState         `json:"item_state"`
	OverlayPolicy OverlayPolicy     `json:"overlay_policy"`
	Virtual       bool              `json:"is_virtual,omitempty"`
	Size          uint64            `json:"size,omitempty"`
	ETag          string            `json:"etag,omitempty"`
	CTag          string            `json:"ctag,omitempty"`
	ContentHash   string            `json:"content_hash,omitempty"`
	LastModified  *time.Time        `json:"last_modified,omitempty"`
	LastHydrated  *time.Time        `json:"last_hydrated,omitempty"`
	LastUploaded  *time.Time        `json:"last_uploaded,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Children      []string          `json:"children,omitempty"`
	SubdirCount   uint32            `json:"subdir_count,omitempty"`
	Mode          uint32            `json:"mode,omitempty"`
	PendingRemote bool              `json:"pending_remote,omitempty"`
	Xattrs        map[string][]byte `json:"xattrs,omitempty"`
	Hydration     HydrationState    `json:"hydration"`
	Upload        UploadState       `json:"upload"`
	Pin           PinState          `json:"pin"`
	LastError     *OperationError   `json:"last_error,omitempty"`
}

// Validate ensures the entry is internally consistent before persistence.
func (e *Entry) Validate() error {
	if e == nil {
		return fmt.Errorf("entry is nil")
	}
	if e.ID == "" {
		return fmt.Errorf("id is required")
	}
	if e.Name == "" {
		return fmt.Errorf("name is required")
	}
	if e.State == "" {
		return fmt.Errorf("item_state is required")
	}
	if err := e.State.Validate(); err != nil {
		return err
	}
	if e.ItemType == "" {
		e.ItemType = ItemKindUnknown
	}
	if err := e.ItemType.Validate(); err != nil {
		return err
	}
	if e.OverlayPolicy == "" {
		e.OverlayPolicy = OverlayPolicyRemoteWins
	}
	if err := e.OverlayPolicy.Validate(); err != nil {
		return err
	}
	if e.Pin.Mode == "" {
		e.Pin.Mode = PinModeUnset
	}
	if err := e.Pin.Mode.Validate(); err != nil {
		return err
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now().UTC()
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = e.CreatedAt
	}
	return nil
}
