package metadata

import "fmt"

// ItemState represents the lifecycle state of a metadata entry.
type ItemState string

const (
	ItemStateGhost      ItemState = "GHOST"
	ItemStateHydrating  ItemState = "HYDRATING"
	ItemStateHydrated   ItemState = "HYDRATED"
	ItemStateDirtyLocal ItemState = "DIRTY_LOCAL"
	ItemStateDeleted    ItemState = "DELETED_LOCAL"
	ItemStateConflict   ItemState = "CONFLICT"
	ItemStateError      ItemState = "ERROR"
)

var validItemStates = map[ItemState]struct{}{
	ItemStateGhost:      {},
	ItemStateHydrating:  {},
	ItemStateHydrated:   {},
	ItemStateDirtyLocal: {},
	ItemStateDeleted:    {},
	ItemStateConflict:   {},
	ItemStateError:      {},
}

// Validate ensures the provided state is one of the supported values.
func (s ItemState) Validate() error {
	if _, ok := validItemStates[s]; ok {
		return nil
	}
	return fmt.Errorf("invalid item_state %q", s)
}

// ItemKind enumerates whether an entry is a file, directory, or unknown.
type ItemKind string

const (
	ItemKindUnknown   ItemKind = "UNKNOWN"
	ItemKindFile      ItemKind = "FILE"
	ItemKindDirectory ItemKind = "DIRECTORY"
)

var validItemKinds = map[ItemKind]struct{}{
	ItemKindUnknown:   {},
	ItemKindFile:      {},
	ItemKindDirectory: {},
}

// Validate ensures the kind is supported.
func (k ItemKind) Validate() error {
	if _, ok := validItemKinds[k]; ok {
		return nil
	}
	return fmt.Errorf("invalid item_type %q", k)
}

// OverlayPolicy defines how virtual entries resolve against remote content.
type OverlayPolicy string

const (
	OverlayPolicyRemoteWins OverlayPolicy = "REMOTE_WINS"
	OverlayPolicyLocalWins  OverlayPolicy = "LOCAL_WINS"
	OverlayPolicyMerged     OverlayPolicy = "MERGED"
)

var validOverlayPolicies = map[OverlayPolicy]struct{}{
	OverlayPolicyRemoteWins: {},
	OverlayPolicyLocalWins:  {},
	OverlayPolicyMerged:     {},
}

// Validate ensures the overlay policy is supported.
func (o OverlayPolicy) Validate() error {
	if o == "" {
		return fmt.Errorf("overlay_policy is required")
	}
	if _, ok := validOverlayPolicies[o]; ok {
		return nil
	}
	return fmt.Errorf("invalid overlay_policy %q", o)
}

// PinMode captures how an entry should be hydrated or evicted.
type PinMode string

const (
	PinModeUnset  PinMode = "UNSET"
	PinModeAlways PinMode = "ALWAYS"
	PinModeNever  PinMode = "NEVER"
	PinModeSmart  PinMode = "SMART"
)

var validPinModes = map[PinMode]struct{}{
	PinModeUnset:  {},
	PinModeAlways: {},
	PinModeNever:  {},
	PinModeSmart:  {},
}

// Validate ensures the pin mode is supported.
func (p PinMode) Validate() error {
	if p == "" {
		return fmt.Errorf("pin mode is required")
	}
	if _, ok := validPinModes[p]; ok {
		return nil
	}
	return fmt.Errorf("invalid pin mode %q", p)
}
