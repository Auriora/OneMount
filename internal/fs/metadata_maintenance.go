package fs

import (
	"encoding/json"
	goerrors "errors"
	"fmt"
	"time"

	"github.com/auriora/onemount/internal/logging"
	"github.com/auriora/onemount/internal/metadata"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// MetadataValidationReport summarizes the results of a validation pass over metadata_v2.
type MetadataValidationReport struct {
	Checked      int
	Invalid      int
	LegacyKeys   int
	MissingV2    bool
	ErrorDetails []string
}

// MetadataMigrationReport summarizes the outcome of migrating legacy metadata into metadata_v2.
type MetadataMigrationReport struct {
	Migrated      int
	LegacyKeys    int
	DroppedLegacy bool
	ErrorDetails  []string
}

// ErrNoLegacyMetadata indicates there was no legacy bucket to migrate.
var ErrNoLegacyMetadata = goerrors.New("legacy metadata bucket missing or empty")

// ValidateMetadataBucket validates the metadata_v2 bucket contents and reports any invalid rows.
func ValidateMetadataBucket(db *bolt.DB) (*MetadataValidationReport, error) {
	report := &MetadataValidationReport{}
	if db == nil {
		return report, fmt.Errorf("metadata validation: db is nil")
	}

	err := db.View(func(tx *bolt.Tx) error {
		legacy := tx.Bucket(bucketMetadata)
		if legacy != nil {
			report.LegacyKeys = legacy.Stats().KeyN
		}

		v2 := tx.Bucket(bucketMetadataV2)
		if v2 == nil {
			report.MissingV2 = true
			return errors.New("metadata_v2 bucket missing")
		}

		return v2.ForEach(func(k, v []byte) error {
			report.Checked++
			if len(v) == 0 {
				report.Invalid++
				report.ErrorDetails = append(report.ErrorDetails, fmt.Sprintf("%s: empty entry", string(k)))
				return nil
			}
			var entry metadata.Entry
			if err := json.Unmarshal(v, &entry); err != nil {
				report.Invalid++
				report.ErrorDetails = append(report.ErrorDetails, fmt.Sprintf("%s: unmarshal error: %v", string(k), err))
				return nil
			}
			if err := entry.Validate(); err != nil {
				report.Invalid++
				report.ErrorDetails = append(report.ErrorDetails, fmt.Sprintf("%s: invalid entry: %v", entry.ID, err))
			}
			return nil
		})
	})

	return report, err
}

// MigrateLegacyMetadata migrates entries from the legacy metadata bucket into metadata_v2.
// The legacy bucket is dropped when migration succeeds.
func MigrateLegacyMetadata(db *bolt.DB) (*MetadataMigrationReport, error) {
	report := &MetadataMigrationReport{}
	if db == nil {
		return report, fmt.Errorf("metadata migration: db is nil")
	}

	err := db.Update(func(tx *bolt.Tx) error {
		legacy := tx.Bucket(bucketMetadata)
		if legacy == nil || legacy.Stats().KeyN == 0 {
			return ErrNoLegacyMetadata
		}
		report.LegacyKeys = legacy.Stats().KeyN

		v2 := tx.Bucket(bucketMetadataV2)
		if v2 == nil {
			return errors.New("metadata_v2 bucket missing")
		}

		now := time.Now().UTC()
		if err := legacy.ForEach(func(k, v []byte) error {
			if len(v) == 0 {
				return nil
			}
			inode, err := NewInodeJSON(v)
			if err != nil {
				report.ErrorDetails = append(report.ErrorDetails, fmt.Sprintf("%s: legacy decode error: %v", string(k), err))
				return nil
			}
			tmpFS := &Filesystem{}
			entry := tmpFS.metadataEntryFromInode(string(k), inode, now)
			if entry == nil {
				report.ErrorDetails = append(report.ErrorDetails, fmt.Sprintf("%s: legacy decode produced nil entry", string(k)))
				return nil
			}
			if err := entry.Validate(); err != nil {
				report.ErrorDetails = append(report.ErrorDetails, fmt.Sprintf("%s: invalid converted entry: %v", entry.ID, err))
				return nil
			}
			blob, err := json.Marshal(entry)
			if err != nil {
				report.ErrorDetails = append(report.ErrorDetails, fmt.Sprintf("%s: marshal error: %v", entry.ID, err))
				return nil
			}
			if err := v2.Put(k, blob); err != nil {
				report.ErrorDetails = append(report.ErrorDetails, fmt.Sprintf("%s: persist error: %v", entry.ID, err))
				return nil
			}
			report.Migrated++
			return nil
		}); err != nil {
			return err
		}

		if err := tx.DeleteBucket(bucketMetadata); err == nil {
			report.DroppedLegacy = true
		} else {
			logging.Warn().Err(err).Msg("Failed to drop legacy metadata bucket after migration")
		}
		return nil
	})

	return report, err
}
