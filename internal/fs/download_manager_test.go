package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	bolt "go.etcd.io/bbolt"
)

// TestRestoreDownloadSessionsRequeues ensures sessions restored from disk are re-enqueued
// so WaitForDownload callers do not hang on queued-but-unprocessed sessions.
func TestUT_FS_DownloadManager_RestoreDownloadSessionsRequeues(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "onemount.db")
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		t.Fatalf("open bolt: %v", err)
	}
	defer db.Close()

	// Seed a queued download session in the DB.
	err = db.Update(func(tx *bolt.Tx) error {
		b, e := tx.CreateBucketIfNotExists(bucketDownloads)
		if e != nil {
			return e
		}
		session := DownloadSession{
			ID:   "restore-id",
			Path: "/restore",
			// state will be reset to queued by restoreDownloadSessions
			State: downloadStarted,
		}
		data, e := json.Marshal(session)
		if e != nil {
			return e
		}
		return b.Put([]byte(session.ID), data)
	})
	if err != nil {
		t.Fatalf("seed session: %v", err)
	}

	// Start manager with 0 workers to avoid processing; queue size 2 for safety.
	dm := NewDownloadManager(nil, nil, 0, 2, db)

	// Assert the session exists and is queued for workers.
	if _, ok := dm.sessions["restore-id"]; !ok {
		t.Fatalf("expected restored session in manager")
	}
	select {
	case id := <-dm.queue:
		if id != "restore-id" {
			t.Fatalf("expected restore-id on queue, got %s", id)
		}
	default:
		t.Fatalf("expected restored session to be re-enqueued")
	}

	// Clean up any temp DB file explicitly on Windows-style FS.
	_ = os.Remove(dbPath)
}
